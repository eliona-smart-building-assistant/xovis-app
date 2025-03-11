//  This file is part of the eliona project.
//  Copyright © 2025 LEICOM iTEC AG. All Rights Reserved.
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

	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
	"github.com/eliona-smart-building-assistant/go-eliona/client"
	"github.com/eliona-smart-building-assistant/go-utils/common"
)

func findChildren(parent api.Asset, assets []api.Asset) (children []api.Asset) {
	for _, asset := range assets {
		if asset.ParentFunctionalAssetId == parent.Id {
			children = append(children, asset)
		}
	}
	return children
}

func GetDashboard(projectId string) (api.Dashboard, error) {
	dashboard := api.Dashboard{}
	dashboard.Name = "Xovis"
	dashboard.ProjectId = projectId
	dashboard.Widgets = []api.Widget{}

	devices, _, err := client.NewClient().AssetsAPI.
		GetAssets(client.AuthenticationContext()).
		AssetTypeName("xovis_people_counter").
		ProjectId(projectId).
		Execute()
	if err != nil {
		return api.Dashboard{}, fmt.Errorf("fetching devices: %v", err)
	}

	zones, _, err := client.NewClient().AssetsAPI.
		GetAssets(client.AuthenticationContext()).
		AssetTypeName("xovis_zone").
		ProjectId(projectId).
		Execute()
	if err != nil {
		return api.Dashboard{}, fmt.Errorf("fetching zones: %v", err)
	}

	lines, _, err := client.NewClient().AssetsAPI.
		GetAssets(client.AuthenticationContext()).
		AssetTypeName("xovis_line").
		ProjectId(projectId).
		Execute()
	if err != nil {
		return api.Dashboard{}, fmt.Errorf("fetching lines: %v", err)
	}
	widgetSequence := int32(0)
	for _, device := range devices {
		deviceZones := findChildren(device, zones)
		deviceLines := findChildren(device, lines)

		var zonesData []api.WidgetData
		for i, zone := range deviceZones {
			zonesData = append(zonesData, api.WidgetData{
				ElementSequence: nullableInt32(1),
				AssetId:         zone.Id,
				Data: map[string]interface{}{
					"aggregatedDataField":  "avg",
					"aggregatedDataRaster": "M30",
					"aggregatedDataType":   "heap",
					"attribute":            "presence",
					"description":          zone.Name.Get(),
					"key":                  "",
					"seq":                  i,
					"subtype":              "input",
				},
			})
		}
		dashboard.Widgets = append(dashboard.Widgets, api.Widget{
			WidgetTypeName: "CombinedTrends",
			AssetId:        device.Id,
			Sequence:       nullableInt32(widgetSequence),
			Details: map[string]any{
				"size":     1,
				"timespan": 7,
			},
			Data: zonesData,
		})
		widgetSequence++

		var linesData []api.WidgetData
		for i, line := range deviceLines {
			linesData = append(linesData, api.WidgetData{
				ElementSequence: nullableInt32(1),
				AssetId:         line.Id,
				Data: map[string]interface{}{
					"aggregatedDataField":  "avg",
					"aggregatedDataRaster": "H1",
					"aggregatedDataType":   "heap",
					"attribute":            "backward",
					"description":          line.Name.Get(),
					// "key": "1741360302119",
					"seq":     i,
					"subtype": "input",
				},
			})
		}
		dashboard.Widgets = append(dashboard.Widgets, api.Widget{
			WidgetTypeName: "GeneralDisplay",
			AssetId:        device.Id,
			Sequence:       nullableInt32(widgetSequence),
			Details: map[string]any{
				"size":     1,
				"timespan": 7,
			},
			Data: linesData,
		})
		widgetSequence++
	}

	return dashboard, nil
}

func nullableInt32(val int32) api.NullableInt32 {
	return *api.NewNullableInt32(common.Ptr[int32](val))
}
