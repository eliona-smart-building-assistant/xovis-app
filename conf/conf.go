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

package conf

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"xovis/appdb"
	confmodel "xovis/model/conf"

	"github.com/eliona-smart-building-assistant/go-eliona/frontend"
	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

var ErrBadRequest = errors.New("bad request")
var ErrNotFound = errors.New("not found")

func InsertConfig(ctx context.Context, config confmodel.Configuration) (confmodel.Configuration, error) {
	dbConfig, err := toDbConfig(ctx, config)
	if err != nil {
		return confmodel.Configuration{}, fmt.Errorf("creating DB config from App config: %v", err)
	}
	if err := dbConfig.InsertG(ctx, boil.Infer()); err != nil {
		return confmodel.Configuration{}, fmt.Errorf("inserting DB config: %v", err)
	}
	return config, nil
}

func UpsertConfig(ctx context.Context, config confmodel.Configuration) (confmodel.Configuration, error) {
	dbConfig, err := toDbConfig(ctx, config)
	if err != nil {
		return confmodel.Configuration{}, fmt.Errorf("creating DB config from App config: %v", err)
	}
	if err := dbConfig.UpsertG(ctx, true, []string{"id"}, boil.Blacklist("id"), boil.Infer()); err != nil {
		return confmodel.Configuration{}, fmt.Errorf("upserting DB config: %v", err)
	}
	return config, nil
}

func GetConfig(ctx context.Context, configID int64) (confmodel.Configuration, error) {
	dbConfig, err := appdb.Configurations(
		appdb.ConfigurationWhere.ID.EQ(configID),
	).OneG(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return confmodel.Configuration{}, ErrNotFound
	}
	if err != nil {
		return confmodel.Configuration{}, fmt.Errorf("fetching config from database: %v", err)
	}
	appConfig, err := toAppConfig(dbConfig)
	if err != nil {
		return confmodel.Configuration{}, fmt.Errorf("creating App config from DB config: %v", err)
	}
	return appConfig, nil
}

func DeleteConfig(ctx context.Context, configID int64) error {
	_, err := appdb.Sensors(
		appdb.SensorWhere.ConfigurationID.EQ(configID),
	).DeleteAllG(ctx)
	if err != nil {
		return fmt.Errorf("deleting sensors from database: %v", err)
	}

	count, err := appdb.Configurations(
		appdb.ConfigurationWhere.ID.EQ(configID),
	).DeleteAllG(ctx)
	if err != nil {
		return fmt.Errorf("deleting config from database: %v", err)
	}
	if count > 1 {
		return fmt.Errorf("shouldn't happen: deleted more (%v) configs by ID", count)
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

func GetConfigs(ctx context.Context) ([]confmodel.Configuration, error) {
	dbConfigs, err := appdb.Configurations().AllG(ctx)
	if err != nil {
		return nil, err
	}
	var appConfigs []confmodel.Configuration
	for _, dbConfig := range dbConfigs {
		ac, err := toAppConfig(dbConfig)
		if err != nil {
			return nil, fmt.Errorf("creating App config from DB config: %v", err)
		}
		appConfigs = append(appConfigs, ac)
	}
	return appConfigs, nil
}

func toDbConfig(ctx context.Context, appConfig confmodel.Configuration) (appdb.Configuration, error) {
	dbConfig := appdb.Configuration{
		ID:               appConfig.ID,
		CheckCertificate: appConfig.CheckCertificate,
		RefreshInterval:  appConfig.RefreshInterval,
		RequestTimeout:   appConfig.RequestTimeout,
		Active:           appConfig.Active,
		Enable:           appConfig.Enable,
		ProjectIds:       appConfig.ProjectIDs,
		UserID:           appConfig.UserId,
	}

	env := frontend.GetEnvironment(ctx)
	if env != nil {
		dbConfig.UserID = env.UserId
	}

	return dbConfig, nil
}

func toAppConfig(dbConfig *appdb.Configuration) (confmodel.Configuration, error) {
	appConfig := confmodel.Configuration{
		ID:               dbConfig.ID,
		CheckCertificate: dbConfig.CheckCertificate,
		RefreshInterval:  dbConfig.RefreshInterval,
		RequestTimeout:   dbConfig.RequestTimeout,
		Active:           dbConfig.Active,
		Enable:           dbConfig.Enable,
		ProjectIDs:       dbConfig.ProjectIds,
		UserId:           dbConfig.UserID,
	}
	return appConfig, nil
}

func InsertSensor(ctx context.Context, sensor confmodel.Sensor) (confmodel.Sensor, error) {
	dbSensor, err := toDbSensor(ctx, sensor)
	if err != nil {
		return confmodel.Sensor{}, fmt.Errorf("creating DB sensor from App sensor: %v", err)
	}
	if err := dbSensor.InsertG(ctx, boil.Infer()); err != nil {
		return confmodel.Sensor{}, fmt.Errorf("inserting DB sensor: %v", err)
	}
	return sensor, nil
}

func UpsertSensor(ctx context.Context, sensor confmodel.Sensor) (confmodel.Sensor, error) {
	dbSensor, err := toDbSensor(ctx, sensor)
	if err != nil {
		return confmodel.Sensor{}, fmt.Errorf("creating DB sensor from App sensor: %v", err)
	}
	if err := dbSensor.UpsertG(ctx, true, []string{"id"}, boil.Blacklist("id"), boil.Infer()); err != nil {
		return confmodel.Sensor{}, fmt.Errorf("upserting DB sensor: %v", err)
	}
	return sensor, nil
}

func GetSensor(ctx context.Context, sensorID int64) (confmodel.Sensor, error) {
	dbSensor, err := appdb.Sensors(
		appdb.SensorWhere.ID.EQ(sensorID),
	).OneG(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return confmodel.Sensor{}, ErrNotFound
	}
	if err != nil {
		return confmodel.Sensor{}, fmt.Errorf("fetching sensor from database: %v", err)
	}
	appSensor, err := toAppSensor(ctx, dbSensor)
	if err != nil {
		return confmodel.Sensor{}, fmt.Errorf("creating App sensor from DB sensor: %v", err)
	}
	return appSensor, nil
}

func DeleteSensor(ctx context.Context, sensorID int64) error {
	count, err := appdb.Sensors(
		appdb.SensorWhere.ID.EQ(sensorID),
	).DeleteAllG(ctx)
	if err != nil {
		return fmt.Errorf("deleting sensor from database: %v", err)
	}
	if count > 1 {
		return fmt.Errorf("shouldn't happen: deleted more (%v) sensors by ID", count)
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

func GetSensors(ctx context.Context) ([]confmodel.Sensor, error) {
	dbSensors, err := appdb.Sensors().AllG(ctx)
	if err != nil {
		return nil, err
	}
	var appSensors []confmodel.Sensor
	for _, dbSensor := range dbSensors {
		as, err := toAppSensor(ctx, dbSensor)
		if err != nil {
			return nil, fmt.Errorf("creating App sensor from DB sensor: %v", err)
		}
		appSensors = append(appSensors, as)
	}
	return appSensors, nil
}

func GetSensorsOfConfig(ctx context.Context, configID int64) ([]confmodel.Sensor, error) {
	dbSensors, err := appdb.Sensors(
		appdb.SensorWhere.ConfigurationID.EQ(configID),
	).AllG(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching sensors from database: %v", err)
	}

	var appSensors []confmodel.Sensor
	for _, dbSensor := range dbSensors {
		appSensor, err := toAppSensor(ctx, dbSensor)
		if err != nil {
			return nil, fmt.Errorf("converting DB sensor to app sensor: %v", err)
		}
		appSensors = append(appSensors, appSensor)
	}

	return appSensors, nil
}

func toDbSensor(ctx context.Context, appSensor confmodel.Sensor) (appdb.Sensor, error) {
	dbSensor := appdb.Sensor{
		ID:              appSensor.ID,
		ConfigurationID: appSensor.Config.ID,
		Username:        appSensor.Username,
		Password:        appSensor.Password,
		Hostname:        appSensor.Hostname,
		Port:            appSensor.Port,
		DiscoveryMode:   appSensor.DiscoveryMode,
		L3FirstIP:       null.StringFromPtr(appSensor.L3FirstIP),
		L3Count:         null.Int32FromPtr(appSensor.L3Count),
	}

	return dbSensor, nil
}

func toAppSensor(ctx context.Context, dbSensor *appdb.Sensor) (confmodel.Sensor, error) {
	appConfig, err := GetConfig(ctx, dbSensor.ConfigurationID)
	if err != nil {
		return confmodel.Sensor{}, fmt.Errorf("fetching related config: %v", err)
	}

	appSensor := confmodel.Sensor{
		ID:            dbSensor.ID,
		Config:        appConfig,
		Username:      dbSensor.Username,
		Password:      dbSensor.Password,
		Hostname:      dbSensor.Hostname,
		Port:          dbSensor.Port,
		DiscoveryMode: dbSensor.DiscoveryMode,
	}

	if dbSensor.L3FirstIP.Valid {
		appSensor.L3FirstIP = &dbSensor.L3FirstIP.String
	}
	if dbSensor.L3Count.Valid {
		appSensor.L3Count = &dbSensor.L3Count.Int32
	}

	return appSensor, nil
}

func SetConfigActiveState(ctx context.Context, config confmodel.Configuration, state bool) (int64, error) {
	return appdb.Configurations(
		appdb.ConfigurationWhere.ID.EQ(config.ID),
	).UpdateAllG(ctx, appdb.M{
		appdb.ConfigurationColumns.Active: state,
	})
}

func SetAllConfigsInactive(ctx context.Context) (int64, error) {
	return appdb.Configurations().UpdateAllG(ctx, appdb.M{
		appdb.ConfigurationColumns.Active: false,
	})
}

func InsertAsset(ctx context.Context, config confmodel.Configuration, projId string, globalAssetID string, assetId int32, providerId string) error {
	var dbAsset appdb.Asset
	dbAsset.ConfigurationID = config.ID
	dbAsset.ProjectID = projId
	dbAsset.GlobalAssetID = globalAssetID
	dbAsset.AssetID = null.Int32From(assetId)
	dbAsset.ProviderID = providerId
	return dbAsset.InsertG(ctx, boil.Infer())
}

func GetAssetId(ctx context.Context, config confmodel.Configuration, projId string, globalAssetID string) (*int32, error) {
	dbAsset, err := appdb.Assets(
		appdb.AssetWhere.ConfigurationID.EQ(config.ID),
		appdb.AssetWhere.ProjectID.EQ(projId),
		appdb.AssetWhere.GlobalAssetID.EQ(globalAssetID),
	).AllG(ctx)
	if err != nil || len(dbAsset) == 0 {
		return nil, err
	}
	return common.Ptr(dbAsset[0].AssetID.Int32), nil
}

func toAppAsset(dbAsset appdb.Asset, config confmodel.Configuration) confmodel.Asset {
	return confmodel.Asset{
		ID:            dbAsset.ID,
		Config:        config,
		ProjectID:     dbAsset.ProjectID,
		GlobalAssetID: dbAsset.GlobalAssetID,
		ProviderID:    dbAsset.ProviderID,
		AssetID:       dbAsset.AssetID.Int32,
	}
}

func GetAssetById(assetId int32) (confmodel.Asset, error) {
	asset, err := appdb.Assets(
		appdb.AssetWhere.AssetID.EQ(null.Int32From(assetId)),
	).OneG(context.Background())
	if err != nil {
		return confmodel.Asset{}, fmt.Errorf("fetching asset: %v", err)
	}
	if !asset.AssetID.Valid {
		return confmodel.Asset{}, fmt.Errorf("shouldn't happen: assetID is nil")
	}
	c, err := asset.Configuration().OneG(context.Background())
	if errors.Is(err, sql.ErrNoRows) {
		return confmodel.Asset{}, ErrNotFound
	}
	if err != nil {
		return confmodel.Asset{}, fmt.Errorf("fetching configuration: %v", err)
	}
	config, err := toAppConfig(c)
	if err != nil {
		return confmodel.Asset{}, fmt.Errorf("translating configuration: %v", err)
	}
	return toAppAsset(*asset, config), nil
}
