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

package assetmodel

import (
	"app-name/conf"
	confmodel "app-name/model/conf"
	"context"
	"fmt"

	"github.com/eliona-smart-building-assistant/go-eliona/asset"
	"github.com/eliona-smart-building-assistant/go-eliona/utils"
	"github.com/eliona-smart-building-assistant/go-utils/common"
)

// TODO: define the asset structure here

type ExampleDevice struct {
	ID   string `eliona:"id" subtype:"info"`
	Name string `eliona:"name,filterable" subtype:"info"`

	Config *confmodel.Configuration
}

func (d *ExampleDevice) AdheresToFilter(filter [][]confmodel.FilterRule) (bool, error) {
	f := apiFilterToCommonFilter(filter)
	fp, err := utils.StructToMap(d)
	if err != nil {
		return false, fmt.Errorf("converting struct to map: %v", err)
	}
	adheres, err := common.Filter(f, fp)
	if err != nil {
		return false, err
	}
	return adheres, nil
}

func (d *ExampleDevice) GetName() string {
	return d.Name
}

func (d *ExampleDevice) GetDescription() string {
	return ""
}

func (d *ExampleDevice) GetAssetType() string {
	return "app_name_device"
}

func (d *ExampleDevice) GetGAI() string {
	return d.GetAssetType() + "_" + d.ID
}

func (d *ExampleDevice) GetAssetID(projectID string) (*int32, error) {
	return conf.GetAssetId(context.Background(), *d.Config, projectID, d.GetGAI())
}

func (d *ExampleDevice) SetAssetID(assetID int32, projectID string) error {
	if err := conf.InsertAsset(context.Background(), *d.Config, projectID, d.GetGAI(), assetID, d.ID); err != nil {
		return fmt.Errorf("inserting asset to config db: %v", err)
	}
	return nil
}

func (d *ExampleDevice) GetLocationalChildren() []asset.LocationalNode {
	return []asset.LocationalNode{}
}

func (d *ExampleDevice) GetFunctionalChildren() []asset.FunctionalNode {
	return []asset.FunctionalNode{}
}

type Root struct {
	locationsMap map[string]ExampleDevice
	devicesSlice []ExampleDevice

	Config *confmodel.Configuration
}

func (r *Root) GetName() string {
	return "app_name"
}

func (r *Root) GetDescription() string {
	return "Root asset for App Name devices"
}

func (r *Root) GetAssetType() string {
	return "app_name_root"
}

func (r *Root) GetGAI() string {
	return r.GetAssetType()
}

func (r *Root) GetAssetID(projectID string) (*int32, error) {
	return conf.GetAssetId(context.Background(), *r.Config, projectID, r.GetGAI())
}

func (r *Root) SetAssetID(assetID int32, projectID string) error {
	if err := conf.InsertAsset(context.Background(), *r.Config, projectID, r.GetGAI(), assetID, ""); err != nil {
		return fmt.Errorf("inserting asset to config db: %v", err)
	}
	return nil
}

func (r *Root) GetLocationalChildren() []asset.LocationalNode {
	locationalChildren := make([]asset.LocationalNode, 0, len(r.locationsMap))
	for _, room := range r.locationsMap {
		roomCopy := room // Create a copy of room
		locationalChildren = append(locationalChildren, &roomCopy)
	}
	return locationalChildren
}

func (r *Root) GetFunctionalChildren() []asset.FunctionalNode {
	functionalChildren := make([]asset.FunctionalNode, 0, len(r.devicesSlice))
	for i := range r.devicesSlice {
		functionalChildren[i] = &r.devicesSlice[i]
	}
	return functionalChildren
}

//

func apiFilterToCommonFilter(input [][]confmodel.FilterRule) [][]common.FilterRule {
	result := make([][]common.FilterRule, len(input))
	for i := 0; i < len(input); i++ {
		result[i] = make([]common.FilterRule, len(input[i]))
		for j := 0; j < len(input[i]); j++ {
			result[i][j] = common.FilterRule{
				Parameter: input[i][j].Parameter,
				Regex:     input[i][j].Regex,
			}
		}
	}
	return result
}
