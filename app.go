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

package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
	"xovis/apiserver"
	"xovis/apiservices"
	"xovis/broker"
	"xovis/conf"
	confmodel "xovis/model/conf"

	"github.com/eliona-smart-building-assistant/go-eliona/app"
	"github.com/eliona-smart-building-assistant/go-eliona/asset"
	"github.com/eliona-smart-building-assistant/go-eliona/dashboard"
	"github.com/eliona-smart-building-assistant/go-eliona/frontend"
	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/eliona-smart-building-assistant/go-utils/db"
	utilshttp "github.com/eliona-smart-building-assistant/go-utils/http"
	"github.com/eliona-smart-building-assistant/go-utils/log"
)

func initialization() {
	ctx := context.Background()

	// Necessary to close used init resources
	conn := db.NewInitConnectionWithContextAndApplicationName(ctx, app.AppName())
	defer conn.Close(ctx)

	// Init the app before the first run.
	app.Init(conn, app.AppName(),
		app.ExecSqlFile("conf/init.sql"),
		asset.InitAssetTypeFiles("resources/asset-types/*.json"),
		dashboard.InitWidgetTypeFiles("resources/widget-types/*.json"),
	)
}

var once sync.Once

func collectData() {
	configs, err := conf.GetConfigs(context.Background())
	if err != nil {
		log.Fatal("conf", "Couldn't read configs from DB: %v", err)
		return
	}
	if len(configs) == 0 {
		once.Do(func() {
			log.Info("conf", "No configs in DB. Please configure the app in Eliona.")
		})
		return
	}

	for _, config := range configs {
		if !config.Enable {
			if config.Active {
				conf.SetConfigActiveState(context.Background(), config, false)
			}
			continue
		}

		if !config.Active {
			conf.SetConfigActiveState(context.Background(), config, true)
			log.Info("conf", "Collecting initialized with Configuration %d:\n"+
				"Enable: %t\n"+
				"Refresh Interval: %d\n"+
				"Request Timeout: %d\n"+
				"Project IDs: %v\n",
				config.ID,
				config.Enable,
				config.RefreshInterval,
				config.RequestTimeout,
				config.ProjectIDs)
		}

		sensors, err := conf.GetSensorsOfConfig(context.Background(), config.ID)
		if err != nil {
			log.Fatal("conf", "Couldn't read sensors from DB: %v", err)
			return
		}
		for _, sensor := range sensors {
			common.RunOnceWithParam(func(sensor confmodel.Sensor) {
				log.Info("main", "Collecting %d started.", sensor.ID)
				if err := collectResources(sensor); err != nil {
					return // Error is handled in the method itself.
				}
				log.Info("main", "Collecting %d finished.", sensor.ID)

				time.Sleep(time.Second * time.Duration(config.RefreshInterval))
			}, sensor, sensor.ID)
		}
	}
}

func collectResources(sensor confmodel.Sensor) error {
	xovis := broker.NewXovisConnector(sensor.Username, sensor.Password, sensor.Hostname, int(sensor.Port), sensor.Config.CheckCertificate, int(sensor.Config.RequestTimeout))
	peopleCounter, err := xovis.GetDevice()
	if err != nil {
		log.Error("broker", "getting peopleCounter: %v", err)
		return err
	}
	peopleCounter.Config = &sensor.Config
	peopleCounter.Lines, peopleCounter.Zones, err = xovis.GetAllCounters()
	if err != nil {
		log.Error("broker", "getting all counters: %v", err)
		return err
	}
	fmt.Println(peopleCounter)
	// Do the magic here
	return nil
}

// listenApi starts the API server and listen for requests
func listenApi() {
	err := http.ListenAndServe(":"+common.Getenv("API_SERVER_PORT", "3000"),
		frontend.NewEnvironmentHandler(
			utilshttp.NewCORSEnabledHandler(
				apiserver.NewRouter(
					apiserver.NewConfigurationAPIController(apiservices.NewConfigurationAPIService()),
					apiserver.NewVersionAPIController(apiservices.NewVersionAPIService()),
					apiserver.NewCustomizationAPIController(apiservices.NewCustomizationAPIService()),
				))))
	log.Fatal("main", "API server: %v", err)
}
