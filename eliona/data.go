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

package eliona

import (
	"fmt"
	confmodel "xovis/model/conf"

	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
	"github.com/eliona-smart-building-assistant/go-eliona/asset"
)

const ClientReference string = "xovis"

func UpsertAssetData(config confmodel.Configuration, assetID int32, data map[string]any) error {
	apiData := api.Data{
		AssetId:         assetID,
		Data:            data,
		ClientReference: *api.NewNullableString(api.PtrString(ClientReference)),
		Subtype:         api.SUBTYPE_INPUT,
	}
	if err := asset.UpsertDataIfAssetExists(apiData); err != nil {
		return fmt.Errorf("upserting data: %v", err)
	}
	return nil
}
