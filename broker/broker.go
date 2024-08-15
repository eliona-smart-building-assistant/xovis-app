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
	basicAuth string
	http      XovisHttp
	login     Login
}

func NewXovisConnector(userName, password, host string, port int, checkCert bool, requestTimeout int) *Xovis {
	return &Xovis{
		basicAuth: encodeBase64(userName + ":" + password),
		login:     Login{},
		http: XovisHttp{
			host:      host,
			port:      strconv.Itoa(port),
			timeout:   time.Duration(requestTimeout) * time.Second,
			checkCert: checkCert,
		},
	}
}

func (x *Xovis) ResetAllCounters() error {
	_, err := x.request(ResetAllCountersPath, http.MethodPost)
	if err != nil {
		return fmt.Errorf("resetting all counters: %w", err)
	}
	return nil
}

func (x *Xovis) GetAllCounters() ([]Line, []Zone, error) {
	var lines []Line
	var zones []Zone

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

			lines = append(lines, Line{
				Name: logic.Name,
				ID:   logic.ID,
				Time: logics.Time,
				Data: lineData,
			})

		case InfoTypeZone, InfoTypeZoneLegacy:
			if len(logic.Counts) != 1 || logic.Counts[0].Name != "balance" {
				log.Debug(module, "unknown counter field in zone: %v", logic.Counts[0])
				continue
			}
			zones = append(zones, Zone{
				Name: logic.Name,
				ID:   logic.ID,
				Time: logics.Time,
				Data: ZoneData{
					FillLevel: logic.Counts[0].Value,
				},
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
