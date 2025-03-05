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

package apiservices

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"xovis/apiserver"
	"xovis/broker"
	"xovis/conf"
	confmodel "xovis/model/conf"
)

// ConfigurationAPIService is a service that implements the logic for the ConfigurationAPIServicer
// This service should implement the business logic for every endpoint for the ConfigurationAPI API.
// Include any external packages or services that will be required by this service.
type ConfigurationAPIService struct {
}

// NewConfigurationAPIService creates a default API service
func NewConfigurationAPIService() apiserver.ConfigurationAPIServicer {
	return &ConfigurationAPIService{}
}

// Configuration methods
func (s *ConfigurationAPIService) GetConfigurations(ctx context.Context) (apiserver.ImplResponse, error) {
	appConfigs, err := conf.GetConfigs(ctx)
	if err != nil {
		return apiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}
	var configs []apiserver.Configuration
	for _, appConfig := range appConfigs {
		configs = append(configs, toAPIConfig(appConfig))
	}
	return apiserver.Response(http.StatusOK, configs), nil
}

func (s *ConfigurationAPIService) PostConfiguration(ctx context.Context, config apiserver.Configuration) (apiserver.ImplResponse, error) {
	appConfig := toAppConfig(config)
	discovered, err := discoverDevices(appConfig)
	if err != nil {
		err = fmt.Errorf("testing configuration: %v", err)
		return apiserver.ImplResponse{Code: http.StatusBadRequest, Body: err}, err
	}
	insertedConfig, err := conf.InsertConfig(ctx, appConfig)
	if err != nil {
		return apiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}
	resp, err := formatResponse(fmt.Sprintf("configuration successfully created and discovered %v new sensors", discovered), toAPIConfig(insertedConfig))
	if err != nil {
		return apiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}
	return apiserver.Response(http.StatusCreated, resp), nil
}

func (s *ConfigurationAPIService) GetConfigurationById(ctx context.Context, configId int64) (apiserver.ImplResponse, error) {
	config, err := conf.GetConfig(ctx, configId)
	if errors.Is(err, conf.ErrNotFound) {
		return apiserver.ImplResponse{Code: http.StatusNotFound}, nil
	}
	if err != nil {
		return apiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}
	return apiserver.Response(http.StatusOK, toAPIConfig(config)), nil
}

func (s *ConfigurationAPIService) PutConfigurationById(ctx context.Context, configId int64, config apiserver.Configuration) (apiserver.ImplResponse, error) {
	config.Id = &configId
	appConfig := toAppConfig(config)
	discovered, err := discoverDevices(appConfig)
	if err != nil {
		err = fmt.Errorf("testing configuration: %v", err)
		return apiserver.ImplResponse{Code: http.StatusBadRequest, Body: err}, err
	}
	upsertedConfig, err := conf.UpsertConfig(ctx, appConfig)
	if err != nil {
		return apiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}
	resp, err := formatResponse(fmt.Sprintf("configuration successfully updated and discovered %v new sensors", discovered), toAPIConfig(upsertedConfig))
	if err != nil {
		return apiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}
	return apiserver.Response(http.StatusCreated, resp), nil
}

func (s *ConfigurationAPIService) DeleteConfigurationById(ctx context.Context, configId int64) (apiserver.ImplResponse, error) {
	err := conf.DeleteConfig(ctx, configId)
	if errors.Is(err, conf.ErrNotFound) {
		return apiserver.ImplResponse{Code: http.StatusNotFound}, nil
	}
	if err != nil {
		return apiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}
	return apiserver.ImplResponse{Code: http.StatusNoContent}, nil
}

// Sensor methods
func (s *ConfigurationAPIService) SensorsGet(ctx context.Context) (apiserver.ImplResponse, error) {
	appSensors, err := conf.GetSensors(ctx)
	if err != nil {
		return apiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}
	var sensors []apiserver.Sensor
	for _, appSensor := range appSensors {
		sensors = append(sensors, toAPISensor(appSensor))
	}
	return apiserver.Response(http.StatusOK, sensors), nil
}

func (s *ConfigurationAPIService) SensorsPost(ctx context.Context, sensor apiserver.SensorCreateUpdate) (apiserver.ImplResponse, error) {
	appSensor := toAppSensor(sensor)
	insertedSensor, err := conf.InsertSensor(ctx, appSensor)
	if err != nil {
		return apiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}
	return apiserver.Response(http.StatusCreated, toAPISensor(insertedSensor)), nil
}

func (s *ConfigurationAPIService) SensorsIdGet(ctx context.Context, sensorId int32) (apiserver.ImplResponse, error) {
	sensor, err := conf.GetSensor(ctx, int64(sensorId))
	if errors.Is(err, conf.ErrNotFound) {
		return apiserver.ImplResponse{Code: http.StatusNotFound}, nil
	}
	if err != nil {
		return apiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}
	return apiserver.Response(http.StatusOK, toAPISensor(sensor)), nil
}

func (s *ConfigurationAPIService) SensorsIdPut(ctx context.Context, sensorId int32, sensor apiserver.SensorCreateUpdate) (apiserver.ImplResponse, error) {
	sensor.Id = sensorId
	appSensor := toAppSensor(sensor)
	upsertedSensor, err := conf.UpsertSensor(ctx, appSensor)
	if err != nil {
		return apiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}
	return apiserver.Response(http.StatusOK, toAPISensor(upsertedSensor)), nil
}

func (s *ConfigurationAPIService) SensorsIdDelete(ctx context.Context, sensorId int32) (apiserver.ImplResponse, error) {
	err := conf.DeleteSensor(ctx, int64(sensorId))
	if errors.Is(err, conf.ErrNotFound) {
		return apiserver.ImplResponse{Code: http.StatusNotFound}, nil
	}
	if err != nil {
		return apiserver.ImplResponse{Code: http.StatusInternalServerError}, err
	}
	return apiserver.ImplResponse{Code: http.StatusNoContent}, nil
}

// Conversion functions
func toAPIConfig(appConfig confmodel.Configuration) apiserver.Configuration {
	return apiserver.Configuration{
		Id:               &appConfig.ID,
		CheckCertificate: appConfig.CheckCertificate,
		Enable:           &appConfig.Enable,
		RefreshInterval:  appConfig.RefreshInterval,
		RequestTimeout:   &appConfig.RequestTimeout,
		Active:           &appConfig.Active,
		ProjectIDs:       &appConfig.ProjectIDs,
		UserId:           &appConfig.UserId,
	}
}

func toAppConfig(apiConfig apiserver.Configuration) confmodel.Configuration {
	appConfig := confmodel.Configuration{
		CheckCertificate: apiConfig.CheckCertificate,
		RefreshInterval:  apiConfig.RefreshInterval,
	}
	if apiConfig.Id != nil {
		appConfig.ID = *apiConfig.Id
	}
	if apiConfig.RequestTimeout != nil {
		appConfig.RequestTimeout = *apiConfig.RequestTimeout
	}
	if apiConfig.Active != nil {
		appConfig.Active = *apiConfig.Active
	}
	if apiConfig.Enable != nil {
		appConfig.Enable = *apiConfig.Enable
	}
	if apiConfig.ProjectIDs != nil {
		appConfig.ProjectIDs = *apiConfig.ProjectIDs
	}
	return appConfig
}

func toAPISensor(appSensor confmodel.Sensor) apiserver.Sensor {
	return apiserver.Sensor{
		Id:              int32(appSensor.ID),
		ConfigurationId: int32(appSensor.Config.ID),
		Username:        appSensor.Username,
		Password:        appSensor.Password,
		Hostname:        appSensor.Hostname,
		Port:            appSensor.Port,
		DiscoveryMode:   appSensor.DiscoveryMode,
		L3FirstIp:       appSensor.L3FirstIP,
		L3Count:         appSensor.L3Count,
	}
}

func toAppSensor(apiSensor apiserver.SensorCreateUpdate) confmodel.Sensor {
	return confmodel.Sensor{
		ID:            int64(apiSensor.Id),
		Config:        confmodel.Configuration{ID: int64(apiSensor.ConfigurationId)},
		Username:      apiSensor.Username,
		Password:      apiSensor.Password,
		Hostname:      apiSensor.Hostname,
		Port:          apiSensor.Port,
		DiscoveryMode: apiSensor.DiscoveryMode,
		L3FirstIP:     apiSensor.L3FirstIp,
		L3Count:       apiSensor.L3Count,
	}
}

func discoverDevices(config confmodel.Configuration) (int, error) {
	sensors, err := conf.GetSensorsOfConfig(context.Background(), config.ID)
	if err != nil {
		return 0, fmt.Errorf("Couldn't read sensors from DB: %v", err)
	}

	discoveredSensors := 0
	for _, sensor := range sensors {
		xovis := broker.NewXovisConnector(sensor)
		discovereds, err := xovis.DiscoverDevices()
		if err != nil {
			return 0, fmt.Errorf("discovering devices: %v", err)
		}
		discoveredSensors += len(discovereds)
	}

	return discoveredSensors, nil
}

// formatResponse marshals the struct and appends it to the text in a nicely formatted way.
func formatResponse(message string, data interface{}) (string, error) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal response data: %w", err)
	}
	return fmt.Sprintf("%s\n\n%s", message, string(jsonData)), nil
}
