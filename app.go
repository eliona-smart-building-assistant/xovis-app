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
	"xovis/eliona"
	assetmodel "xovis/model/asset"
	confmodel "xovis/model/conf"
	"xovis/webhook"

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

		common.RunOnceWithParam(func(config confmodel.Configuration) {
			log.Info("main", "Discovering %d started.", config.ID)
			discovered, err := discoverDevices(config)
			if err != nil {
				return // Error is handled in the method itself.
			}
			log.Info("main", "Discovered %d devices for config %d.", discovered, config.ID)

			time.Sleep(time.Second * 100 * time.Duration(config.RefreshInterval))
		}, config, fmt.Sprintf("discovery %d", config.ID))

		common.RunOnceWithParam(func(config confmodel.Configuration) {
			log.Info("main", "Collecting %d started.", config.ID)
			if err := collectResources(config); err != nil {
				return // Error is handled in the method itself.
			}
			log.Info("main", "Collecting %d finished.", config.ID)

			time.Sleep(time.Second * time.Duration(config.RefreshInterval))
		}, config, config.ID)
	}
}

func discoverDevices(config confmodel.Configuration) (int, error) {
	sensors, err := conf.GetSensorsOfConfig(context.Background(), config.ID)
	if err != nil {
		log.Error("conf", "Couldn't read sensors from DB: %v", err)
		return 0, err
	}

	discoveredSensors := 0
	for _, sensor := range sensors {
		xovis := broker.NewXovisConnector(sensor)
		discovereds, err := xovis.DiscoverDevices()
		if err != nil {
			log.Error("broker", "discovering devices: %v", err)
			return discoveredSensors, err
		}

		for _, discovered := range discovereds {
			if _, err := conf.UpsertSensorDiscovery(context.Background(), discovered); err != nil {
				log.Error("conf", "upserting discovered sensor %+v: %v", discovered, err)
				return discoveredSensors, err
			}
		}
		discoveredSensors += len(discovereds)
	}

	return discoveredSensors, nil
}

func collectResources(config confmodel.Configuration) error {
	sensors, err := conf.GetSensorsOfConfig(context.Background(), config.ID)
	if err != nil {
		log.Error("conf", "Couldn't read sensors from DB: %v", err)
		return err
	}

	root := assetmodel.Root{
		Groups: map[string]assetmodel.Group{},
		Config: &config,
	}
	for _, sensor := range sensors {
		xovis := broker.NewXovisConnector(sensor)
		peopleCounter, err := xovis.GetDevice()
		if err != nil {
			log.Error("broker", "getting peopleCounter: %v", err)
			return err
		}
		peopleCounter.Lines, peopleCounter.Zones, err = xovis.GetAllCounters()
		if err != nil {
			log.Error("broker", "getting all counters: %v", err)
			return err
		}

		groupName := peopleCounter.Group
		group, ok := root.Groups[groupName]
		if !ok {
			group = assetmodel.Group{
				Name:    groupName,
				Sensors: []assetmodel.PeopleCounter{},
				Config:  &config,
			}
		}

		group.Sensors = append(group.Sensors, peopleCounter)
		root.Groups[groupName] = group
	}

	if err := eliona.CreateAssetsAndUpsertData(config, &root); err != nil {
		log.Error("eliona", "creating assets: %v", err)
		return err
	}

	return nil
}

func listenApi() {
	mux := http.NewServeMux()

	// Add API Server routes
	apiRouter := apiserver.NewRouter(
		apiserver.NewConfigurationAPIController(apiservices.NewConfigurationAPIService()),
		apiserver.NewVersionAPIController(apiservices.NewVersionAPIService()),
		apiserver.NewCustomizationAPIController(apiservices.NewCustomizationAPIService()),
	)
	mux.Handle("/", apiRouter)

	// Register Webhook handler under /webhook
	mux.Handle("/webhook/", webhook.NewWebhookHandler())

	// Wrap with middleware
	handler := frontend.NewEnvironmentHandler(
		utilshttp.NewCORSEnabledHandler(mux),
	)

	// Start the server
	port := common.Getenv("API_SERVER_PORT", "3030")
	err := http.ListenAndServe(":"+port, handler)
	log.Fatal("main", "API server: %v", err)
}
