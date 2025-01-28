//  This file is part of the Eliona project.
//  Copyright © 2025 IoTEC AG. All Rights Reserved.
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

package webhook

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"xovis/conf"
	"xovis/eliona"

	"github.com/eliona-smart-building-assistant/go-utils/log"
)

type webhookServer struct {
	mux *http.ServeMux
}

func newWebhookServer() *webhookServer {
	return &webhookServer{
		mux: http.NewServeMux(),
	}
}

func (s *webhookServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Debug("webhook", "Received request for URL: %s, Method: %s", r.URL.Path, r.Method)

	configID, err := parseConfigIDFromPath(r.URL.Path)
	if err != nil {
		log.Warn("webhook", "Invalid URL path, missing or invalid config ID: %s", r.URL.Path)
		http.Error(w, "Invalid URL path, missing or invalid config ID", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, "configID", configID)
	r = r.WithContext(ctx)

	r.URL.Path = removeConfigIDFromPath(r.URL.Path)

	// Use a custom ResponseWriter to capture all status codes
	lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
	s.mux.ServeHTTP(lrw, r)

	// Log all errors (non-2xx status codes)
	if lrw.statusCode >= 400 {
		log.Error("webhook", "Error response: Status=%d, URL=%s, Method=%s", lrw.statusCode, r.URL.Path, r.Method)
	}
}

// loggingResponseWriter is a wrapper for http.ResponseWriter to capture the status code.
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

type WebhookData struct {
	LiveData struct {
		PackageInfo struct {
			Version string `json:"version"`
			ID      int    `json:"id"`
			AgentID int    `json:"agent_id"`
		} `json:"package_info"`
		SensorInfo struct {
			SerialNumber string `json:"serial_number"`
			Type         string `json:"type"`
		} `json:"sensor_info"`
		Config struct {
			Logics []struct {
				ID           int    `json:"id"`
				Name         string `json:"name"`
				OptionalData string `json:"optional_data"`
				Geometries   []int  `json:"geometries"`
			} `json:"logics"`
			Counts []struct {
				ID      int    `json:"id"`
				Name    string `json:"name"`
				LogicID int    `json:"logic_id"`
				Type    string `json:"type"`
			} `json:"counts"`
			Geometries []struct {
				ID       int         `json:"id"`
				Name     string      `json:"name"`
				Type     string      `json:"type"`
				Geometry [][]float64 `json:"geometry"`
			} `json:"geometries"`
		} `json:"config"`
		Frames []struct {
			FrameNumber    int    `json:"framenumber"`
			FrameType      string `json:"frametype"`
			Time           int64  `json:"time"`
			Illumination   string `json:"illumination"`
			TrackedObjects []struct {
				TrackID    int       `json:"track_id"`
				Type       string    `json:"type"`
				Position   []float64 `json:"position"`
				Attributes struct {
					PersonHeight float64 `json:"person_height"`
					Members      int     `json:"members"`
				} `json:"attributes"`
			} `json:"tracked_objects"`
			Events []struct {
				Category   string `json:"category"`
				Type       string `json:"type"`
				Attributes struct {
					CounterID    int `json:"counter_id"`
					CounterValue int `json:"counter_value"`
					TrackID      int `json:"track_id"`
				} `json:"attributes"`
			} `json:"events"`
		} `json:"frames"`
	} `json:"live_data"`
}

func (s *webhookServer) handleDatapush(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	log.Trace("datapush", "raw datapush:\n%s\n", string(body))

	var data WebhookData
	if err := json.Unmarshal(body, &data); err != nil {
		http.Error(w, "Failed to parse request body", http.StatusInternalServerError)
		return
	}

	for _, frame := range data.LiveData.Frames {
		for _, event := range frame.Events {
			switch event.Category {
			case "COUNT":
				counterValue := event.Attributes.CounterValue
				rawCounterID := event.Attributes.CounterID
				logicID := rawCounterID / 1000   // Get the first part (e.g., 1008 from 1008001)
				counterID := rawCounterID % 1000 // Get the last part (e.g., 001 from 1008001)

				dataToUpsert := map[string]any{"presence": counterValue}
				gai := fmt.Sprintf("xovis_zone_%v_%v", data.LiveData.SensorInfo.SerialNumber, logicID)
				asset, err := conf.GetAssetByGAI(gai)

				// Looks like there is no better way now to distinguish lines and zones...
				if errors.Is(err, conf.ErrNotFound) {
					gai = fmt.Sprintf("xovis_line_%v_%v", data.LiveData.SensorInfo.SerialNumber, logicID)
					asset, err = conf.GetAssetByGAI(gai)
					if err != nil {
						log.Error("datapush", "getting asset by GAI %s: %v", gai, err)
						err = nil
						continue
					}

					// Determine the key based on counterID (001 -> "forward", 002 -> "backward")
					var key string
					switch counterID {
					case 1:
						key = "forward"
					case 2:
						key = "backward"
					default:
						log.Warn("datapush", "unknown counter ID %v for logic %v, skipping", counterID, logicID)
						continue
					}

					dataToUpsert = map[string]any{key: counterValue}
				}
				if err != nil {
					log.Error("datapush", "getting asset by GAI %s: %v", gai, err)
					continue
				}
				if err := eliona.UpsertAssetData(asset.Config, asset.AssetID, dataToUpsert); err != nil {
					log.Error("datapush", "upserting data: %v", err)
					continue
				}
				log.Debug("datapush", "set %v data %+v", asset.AssetID, dataToUpsert)
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}

func parseConfigIDFromPath(path string) (int64, error) {
	// Matches "/{configID}/rest-of-path"
	re := regexp.MustCompile(`^/(\d+)/`)
	matches := re.FindStringSubmatch(path)
	if len(matches) < 2 {
		return 0, fmt.Errorf("config ID not found in path")
	}
	return strconv.ParseInt(matches[1], 10, 64)
}

func removeConfigIDFromPath(path string) string {
	re := regexp.MustCompile(`^/\d+/`)
	return re.ReplaceAllString(path, "/")
}

func StartWebhookListener() {
	server := newWebhookServer()

	server.mux.HandleFunc("/datapush", server.handleDatapush)

	http.Handle("/", server)
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal("webhook", "Error starting server on port 8081: %v\n", err)
	}
}
