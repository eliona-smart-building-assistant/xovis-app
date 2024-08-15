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
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/eliona-smart-building-assistant/go-utils/log"
)

const (
	MODULE = "broker"

	UserViewer = "viewer"
	UserAdmin  = "admin"

	InfoTypeLineLegacy = "XLT_4X_LINE_IN_OUT_COUNT"
	InfoTypeZoneLegacy = "XLT_4X_ZONE_COUNT"
	InfoTypeLine       = "XLT_LINE_IN_OUT_COUNT"
	InfoTypeZone       = "XLT_ZONE_OCCUPANCY_COUNT"

	ApiPath = "/api/v5"

	AllCounterPath      = ApiPath + "/singlesensor/data/live/logics"
	ResetAllCounterPath = ApiPath + "/singlesensor/data/live/counts/reset"
)

type LineData struct {
	ForwardTotal   int `json:"fw_tot"`
	BackwardTotal  int `json:"bw_tot"`
	ForwardDiv     int `json:"rtl_counter"`
	BackwardDiv    int `json:"ltr_counter"`
	ForwardMale    int `json:"fw_male"`
	BackwardMale   int `json:"bw_male"`
	ForwardFemale  int `json:"fw_female"`
	BackwardFemale int `json:"bw_female"`
	ForwardMask    int `json:"fw_mask"`
	BackwardMask   int `json:"bw_mask"`
	ForwardNoMask  int `json:"fw_no_mask"`
	BackwardNoMask int `json:"bw_no_mask"`
}

type ZoneData struct {
	ForwardTotal  int `json:"fw_tot"`
	BackwardTotal int `json:"bw_tot"`
	ForwardDiv    int `json:"rtl_counter"`
	BackwardDiv   int `json:"ltr_counter"`
	ObjectCount   int `json:"object_cnt"`
	FillLevel     int `json:"fill_level"`
}

type Line struct {
	Name string
	Id   int
	Data LineData
	Time string
}

type Zone struct {
	Name string
	Id   int
	Data ZoneData
	Time string
}

type XovisHttp struct {
	host      string
	port      string
	timeOut   int
	checkCert bool
}

func (con *XovisHttp) Request(method, apiPath string, headers map[string]string) ([]byte, error) {
	url := "https://" + con.host + ":" + con.port + apiPath
	timeout := time.Duration(con.timeOut) * time.Second

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !con.checkCert},
	}

	httpClient := http.Client{
		Timeout:   timeout,
		Transport: tr,
	}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		log.Error(MODULE, "error creating request: %v", err)
		return nil, err
	}
	for header, value := range headers {
		req.Header.Set(header, value)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Error(MODULE, "error request to %s %v", url, err)
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error(MODULE, "error reading body from %s", url)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusCreated {
		log.Warn(MODULE, "%s not ok: status code: %d", url, resp.StatusCode)
		log.Debug(MODULE, " -> with: %v, %v", headers, string(respBody))
		return respBody, err
	}

	return respBody, err
}

type Geometrie struct {
	Id   int    `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}

type Count struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type Logic struct {
	Id         int         `json:"id"`
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
	basicAuth string
	con       XovisHttp
	login     Login
}

func NewXovisConnector(userName, password, host string, port int, checkCert bool, requestTimeout int) *Xovis {
	return &Xovis{
		basicAuth: encodeBase64(userName + ":" + password),
		login:     Login{},
		con: XovisHttp{
			host:      host,
			port:      strconv.Itoa(port),
			timeOut:   requestTimeout,
			checkCert: checkCert,
		},
	}
}

func (x *Xovis) ResetAllCounters() error {
	_, err := x.request(ResetAllCounterPath, http.MethodPost)
	if err != nil {
		log.Error(MODULE, "error resetting all counters: %v", err)
	}
	return err
}

func (x *Xovis) GetAllCounters() ([]Line, []Zone, error) {
	var lines []Line
	var zones []Zone

	logics, err := x.getCountersRaw()
	if err != nil {
		log.Error(MODULE, "error getting counter data: %v", err)
		return nil, nil, err
	}

	for _, logic := range logics.Logics {
		switch logic.Info {
		case InfoTypeLine, InfoTypeLineLegacy:
			lData := LineData{
				ForwardTotal:   -1,
				BackwardTotal:  -1,
				ForwardDiv:     0,
				BackwardDiv:    0,
				ForwardMale:    -1,
				BackwardMale:   -1,
				ForwardFemale:  -1,
				BackwardFemale: -1,
				ForwardMask:    -1,
				BackwardMask:   -1,
				ForwardNoMask:  -1,
				BackwardNoMask: -1,
			}

			for _, count := range logic.Counts {
				switch count.Name {
				case "bw":
					lData.BackwardTotal = count.Value
				case "fw":
					lData.ForwardTotal = count.Value
				case "fw-male":
					lData.ForwardMale = count.Value
				case "bw-male":
					lData.BackwardMale = count.Value
				case "fw-female":
					lData.ForwardFemale = count.Value
				case "bw-female":
					lData.BackwardFemale = count.Value
				case "fw-mask":
					lData.ForwardMask = count.Value
				case "bw-mask":
					lData.BackwardMask = count.Value
				case "fw-no_mask":
					lData.ForwardNoMask = count.Value
				case "bw-no_mask":
					lData.BackwardNoMask = count.Value
				default:
					log.Debug(MODULE, "unknown counter type: %v", count.Name)
				}
			}

			lines = append(lines, Line{
				Name: logic.Name,
				Id:   logic.Id,
				Time: logics.Time,
				Data: lData,
			})

		case InfoTypeZone, InfoTypeZoneLegacy:
			if len(logic.Counts) != 1 || logic.Counts[0].Name != "balance" {
				log.Debug(MODULE, "unknown counter field in zone: %v", logic.Counts[0])
				break
			}
			zones = append(zones, Zone{
				Name: logic.Name,
				Id:   logic.Id,
				Time: logics.Time,
				Data: ZoneData{
					FillLevel:     logic.Counts[0].Value,
					ObjectCount:   0,
					ForwardDiv:    0,
					BackwardDiv:   0,
					ForwardTotal:  0,
					BackwardTotal: 0,
				},
			})

		default:
			log.Debug(MODULE, "unknown counter type: %v", logic.Info)
		}
	}

	return lines, zones, nil
}

func (x *Xovis) getCountersRaw() (Logics, error) {
	var logics Logics
	rawData, err := x.request(AllCounterPath, http.MethodGet)
	if err != nil {
		log.Error(MODULE, "get counter data: %v", err)
		return logics, err
	}
	err = json.Unmarshal(rawData, &logics)
	if err != nil {
		log.Error(MODULE, "decode logics: %v", err)
		return logics, err
	}
	return logics, nil
}

func (x *Xovis) request(path, method string) ([]byte, error) {
	headers := map[string]string{
		"Authorization": "Basic " + x.basicAuth,
		"Accept":        "application/json",
	}

	jsonBody, err := x.con.Request(method, path, headers)
	if err != nil {
		x.login.LastUsedAt = 0
		x.login.ReceivedAt = 0
		return jsonBody, err
	}

	x.login.LastUsedAt = time.Now().Unix()
	return jsonBody, nil
}

func (x *Xovis) isTokenValid() bool {
	now := time.Now().Unix()
	if x.login.ReceivedAt+int64(x.login.ValidFor) <= now+240 || x.login.LastUsedAt+int64(x.login.MaxUnusedFor) <= now+240 {
		log.Debug(MODULE, "token expired")
		return false
	}
	return true
}

func encodeBase64(plain string) string {
	return base64.StdEncoding.EncodeToString([]byte(plain))
}
