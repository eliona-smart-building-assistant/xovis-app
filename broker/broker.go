//  This file is part of the Eliona project.
//  Copyright © 2024 IoTEC AG. All Rights Reserved.
//  ______ _ _
// |  ____| (_)
// | |__  | |_  ___  _ __   __ _
// |  __| | | |/ _ \| '_ \ / _` |
// | |____| | | (_) | | | | (_| |
// |______|_|_|\___/|_| |_|\__,_|
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
//  BUT NOT LIMITED  TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
//  NON INFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
//  DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package broker

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
	assetmodel "xovis/model/asset"
	confmodel "xovis/model/conf"

	"github.com/eliona-smart-building-assistant/go-utils/log"
)

const (
	module = "broker"

	UserViewer = "viewer"
	UserAdmin  = "admin"

	InfoTypeLineLegacy = "XLT_4X_LINE_IN_OUT_COUNT"
	InfoTypeZoneLegacy = "XLT_4X_ZONE_COUNT"
	InfoTypeLine       = "XLT_LINE_IN_OUT_COUNT"
	InfoTypeZone       = "XLT_ZONE_OCCUPANCY_COUNT"

	ApiPath = "/api/v5"

	AllCountersPath      = ApiPath + "/singlesensor/data/live/logics"
	ResetAllCountersPath = ApiPath + "/singlesensor/data/live/counts/reset"
)

type LineData struct {
	ForwardTotal  int
	BackwardTotal int
}

type ZoneData struct {
	FillLevel int
}

type Line struct {
	Name string
	ID   int
	Data LineData
	Time string
}

type Zone struct {
	Name string
	ID   int
	Data ZoneData
	Time string
}

type XovisHttp struct {
	host      string
	port      string
	timeout   time.Duration
	checkCert bool
}

func (httpClient *XovisHttp) Request(method, apiPath string, headers map[string]string) ([]byte, error) {
	url := "https://" + httpClient.host + ":" + httpClient.port + apiPath

	client := &http.Client{
		Timeout: httpClient.timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: !httpClient.checkCert},
		},
	}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to %s %w", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading body from %s: %w", url, err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusCreated {
		log.Debug(module, " -> with: %v, %v", headers, string(body))
		return body, fmt.Errorf("%s not ok: status code: %d", url, resp.StatusCode)
	}

	return body, nil
}

type Geometrie struct {
	ID   int    `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}

type Count struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type Logic struct {
	ID         int         `json:"id"`
	Name       string      `json:"name"`
	Info       string      `json:"info"`
	Geometries []Geometrie `json:"geometries"`
	Counts     []Count     `json:"counts"`
}

type Logics struct {
	Time   string  `json:"time"`
	Logics []Logic `json:"logics"`
}

type Login struct {
	Token        string `json:"token"`
	ValidFor     int    `json:"valid_for"`
	MaxUnusedFor int    `json:"max_unused_for"`
	ReceivedAt   int64
	LastUsedAt   int64
}

type Xovis struct {
	basicAuth  string
	http       XovisHttp
	login      Login
	sensorConf confmodel.Sensor
}

func NewXovisConnector(sensorConf confmodel.Sensor) *Xovis {
	return &Xovis{
		basicAuth: encodeBase64(sensorConf.Username + ":" + sensorConf.Password),
		login:     Login{},
		http: XovisHttp{
			host:      sensorConf.Hostname,
			port:      strconv.Itoa(int(sensorConf.Port)),
			timeout:   time.Duration(sensorConf.Config.RequestTimeout) * time.Second,
			checkCert: sensorConf.Config.CheckCertificate,
		},
		sensorConf: sensorConf,
	}
}

func (x *Xovis) DiscoverDevices() ([]confmodel.Sensor, error) {
	var resp []byte
	var err error
	switch x.sensorConf.DiscoveryMode {
	case "L2":
		resp, err = x.http.Request(http.MethodGet, ApiPath+"/discover/localnetwork", nil)
		if err != nil {
			return nil, fmt.Errorf("making L2 request: %w", err)
		}
	case "L3":
		if x.sensorConf.L3FirstIP == nil || x.sensorConf.L3Count == nil {
			return nil, fmt.Errorf("L3 discovery mode requires L3FirstIP and L3Count to be set")
		}
		body := map[string]string{
			"first_ip": *x.sensorConf.L3FirstIP,
			"count":    string(*x.sensorConf.L3Count),
		}
		resp, err = x.http.Request(http.MethodPost, ApiPath+"/discover/scan", body)
		if err != nil {
			return nil, fmt.Errorf("making L3 request: %w", err)
		}
	case "disabled":
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown discovery mode: %s", x.sensorConf.DiscoveryMode)
	}

	deviceItself, err := x.getDeviceInfo()
	if err != nil {
		return nil, fmt.Errorf("making request to get the device itself: %w", err)
	}

	var discoveryResult struct {
		Sensors []struct {
			MAC   string   `json:"mac"`
			IP    string   `json:"ip"`
			IPv6  []string `json:"ipv6"`
			Ports []struct {
				Number  int32  `json:"number"`
				Service string `json:"service"`
			} `json:"ports"`
			Model     string `json:"model"`
			Name      string `json:"name"`
			Group     string `json:"group"`
			FWVersion string `json:"fw_version"`
		} `json:"sensors"`
	}

	if err := json.Unmarshal(resp, &discoveryResult); err != nil {
		return nil, fmt.Errorf("parsing discovery response: %w", err)
	}

	var sensors []confmodel.Sensor
	for _, responseSensor := range discoveryResult.Sensors {
		hostname := responseSensor.IP
		if hostname == "" && len(responseSensor.IPv6) > 0 {
			hostname = responseSensor.IPv6[0]
		}
		port := int32(443)
		for _, p := range responseSensor.Ports {
			if p.Service == "https" {
				port = p.Number
			}
		}
		sensor := confmodel.Sensor{
			Config:        x.sensorConf.Config,
			Username:      x.sensorConf.Username,
			Password:      x.sensorConf.Password,
			Hostname:      hostname,
			Port:          port,
			DiscoveryMode: "disabled", // no value in discovering devices in the same range
			MACAddress:    &responseSensor.MAC,
		}

		// We need to identify the device itself to avoid overwriting it.
		if responseSensor.MAC == deviceItself.MAC {
			sensor.ID = x.sensorConf.ID
		}

		sensors = append(sensors, sensor)
	}

	return sensors, nil
}

func (x *Xovis) GetDevice() (assetmodel.PeopleCounter, error) {
	idResp, err := x.getDeviceID()
	if err != nil {
		return assetmodel.PeopleCounter{}, fmt.Errorf("getting device ID: %v", err)
	}

	deviceInfoResp, err := x.getDeviceInfo()
	if err != nil {
		return assetmodel.PeopleCounter{}, fmt.Errorf("getting device info: %v", err)
	}

	return assetmodel.PeopleCounter{
		Name:   idResp.Name,
		Group:  idResp.Group,
		MAC:    deviceInfoResp.MAC,
		Model:  deviceInfoResp.Type,
		Config: &x.sensorConf.Config,
	}, nil
}

type idResponse struct {
	Group string `json:"group"`
	Name  string `json:"name"`
}

func (x *Xovis) getDeviceID() (idResponse, error) {
	resp, err := x.http.Request(http.MethodGet, ApiPath+"/device/id", nil)
	if err != nil {
		return idResponse{}, fmt.Errorf("making request to get device id: %w", err)
	}

	var idResp idResponse
	if err := json.Unmarshal(resp, &idResp); err != nil {
		return idResponse{}, fmt.Errorf("parsing device id response: %w", err)
	}

	return idResp, nil
}

type deviceInfoResponse struct {
	MAC  string `json:"serial"`
	Type string `json:"type"`
}

func (x *Xovis) getDeviceInfo() (deviceInfoResponse, error) {
	resp, err := x.http.Request(http.MethodGet, ApiPath+"/device/info", nil)
	if err != nil {
		return deviceInfoResponse{}, fmt.Errorf("making request to get device info: %w", err)
	}

	var deviceInfoResp deviceInfoResponse
	if err := json.Unmarshal(resp, &deviceInfoResp); err != nil {
		return deviceInfoResponse{}, fmt.Errorf("parsing device info response: %w", err)
	}

	return deviceInfoResp, nil
}

func (x *Xovis) ResetAllCounters() error {
	_, err := x.request(ResetAllCountersPath, http.MethodPost)
	if err != nil {
		return fmt.Errorf("resetting all counters: %w", err)
	}
	return nil
}

func (x *Xovis) GetAllCounters() ([]assetmodel.Line, []assetmodel.Zone, error) {
	var lines []assetmodel.Line
	var zones []assetmodel.Zone

	deviceInfoResp, err := x.getDeviceInfo()
	if err != nil {
		return nil, nil, fmt.Errorf("getting device info: %v", err)
	}

	logics, err := x.getCountersRaw()
	if err != nil {
		return nil, nil, fmt.Errorf("getting counter data: %w", err)
	}

	for _, logic := range logics.Logics {
		switch logic.Info {
		case InfoTypeLine, InfoTypeLineLegacy:
			lineData := LineData{
				ForwardTotal:  -1,
				BackwardTotal: -1,
			}

			for _, count := range logic.Counts {
				switch count.Name {
				case "bw":
					lineData.BackwardTotal = count.Value
				case "fw":
					lineData.ForwardTotal = count.Value
				default:
					log.Debug(module, "unknown counter type: %v", count.Name)
				}
			}

			lines = append(lines, assetmodel.Line{
				Name:      logic.Name,
				ID:        logic.ID,
				Forward:   lineData.ForwardTotal,
				Backward:  lineData.BackwardTotal,
				DeviceMac: deviceInfoResp.MAC,
				Config:    &x.sensorConf.Config,
			})

		case InfoTypeZone, InfoTypeZoneLegacy:
			if len(logic.Counts) != 1 || logic.Counts[0].Name != "balance" {
				log.Debug(module, "unknown counter field in zone: %v", logic.Counts[0])
				continue
			}
			zones = append(zones, assetmodel.Zone{
				Name:      logic.Name,
				ID:        logic.ID,
				Presence:  logic.Counts[0].Value,
				DeviceMac: deviceInfoResp.MAC,
				Config:    &x.sensorConf.Config,
			})

		default:
			log.Debug(module, "unknown counter type: %v", logic.Info)
		}
	}

	return lines, zones, nil
}

func (x *Xovis) getCountersRaw() (Logics, error) {
	var logics Logics
	rawData, err := x.request(AllCountersPath, http.MethodGet)
	if err != nil {
		return logics, fmt.Errorf("getting counter data: %w", err)
	}
	if err := json.Unmarshal(rawData, &logics); err != nil {
		return logics, fmt.Errorf("decoding logics: %w", err)
	}
	return logics, nil
}

func (x *Xovis) request(path, method string) ([]byte, error) {
	headers := map[string]string{
		"Authorization": "Basic " + x.basicAuth,
		"Accept":        "application/json",
	}
	jsonBody, err := x.http.Request(method, path, headers)
	if err != nil {
		x.login.LastUsedAt = 0
		x.login.ReceivedAt = 0
		return jsonBody, fmt.Errorf("request error: %w", err)
	}
	x.login.LastUsedAt = time.Now().Unix()
	return jsonBody, nil
}

func (x *Xovis) isTokenValid() bool {
	now := time.Now().Unix()
	if x.login.ReceivedAt+int64(x.login.ValidFor) <= now+240 || x.login.LastUsedAt+int64(x.login.MaxUnusedFor) <= now+240 {
		log.Debug(module, "token expired")
		return false
	}
	return true
}

func encodeBase64(plain string) string {
	return base64.StdEncoding.EncodeToString([]byte(plain))
}
