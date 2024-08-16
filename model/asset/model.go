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
	"context"
	"fmt"
	"xovis/conf"
	confmodel "xovis/model/conf"

	"github.com/eliona-smart-building-assistant/go-eliona/asset"
)

type Zone struct {
	ID       int    `eliona:"id" subtype:"info"`
	Name     string `eliona:"name" subtype:"info"`
	Presence int    `eliona:"presence" subtype:"input"`

	Config *confmodel.Configuration
}

func (d *Zone) GetName() string {
	return d.Name
}

func (d *Zone) GetDescription() string {
	return "Xovis People Counter Zone" + d.Name
}

func (d *Zone) GetAssetType() string {
	return "xovis_zone"
}

func (d *Zone) GetGAI() string {
	return d.GetAssetType() + "_" + string(d.ID)
}

func (d *Zone) GetAssetID(projectID string) (*int32, error) {
	return conf.GetAssetId(context.Background(), *d.Config, projectID, d.GetGAI())
}

func (d *Zone) SetAssetID(assetID int32, projectID string) error {
	if err := conf.InsertAsset(context.Background(), *d.Config, projectID, d.GetGAI(), assetID, string(d.ID)); err != nil {
		return fmt.Errorf("inserting asset to config db: %v", err)
	}
	return nil
}

func (d *Zone) GetLocationalChildren() []asset.LocationalNode {
	return []asset.LocationalNode{}
}

func (d *Zone) GetFunctionalChildren() []asset.FunctionalNode {
	return []asset.FunctionalNode{}
}

type Line struct {
	ID       int    `eliona:"id" subtype:"info"`
	Name     string `eliona:"name" subtype:"info"`
	Forward  int    `eliona:"forward" subtype:"input"`
	Backward int    `eliona:"backward" subtype:"input"`

	Config *confmodel.Configuration
}

func (d *Line) GetName() string {
	return d.Name
}

func (d *Line) GetDescription() string {
	return "Xovis People Counter Line" + d.Name
}

func (d *Line) GetAssetType() string {
	return "xovis_line"
}

func (d *Line) GetGAI() string {
	return d.GetAssetType() + "_" + string(d.ID)
}

func (d *Line) GetAssetID(projectID string) (*int32, error) {
	return conf.GetAssetId(context.Background(), *d.Config, projectID, d.GetGAI())
}

func (d *Line) SetAssetID(assetID int32, projectID string) error {
	if err := conf.InsertAsset(context.Background(), *d.Config, projectID, d.GetGAI(), assetID, string(d.ID)); err != nil {
		return fmt.Errorf("inserting asset to config db: %v", err)
	}
	return nil
}

func (d *Line) GetLocationalChildren() []asset.LocationalNode {
	return []asset.LocationalNode{}
}

func (d *Line) GetFunctionalChildren() []asset.FunctionalNode {
	return []asset.FunctionalNode{}
}

type PeopleCounter struct {
	MAC   string `eliona:"mac" subtype:"info"`
	Name  string `eliona:"name" subtype:"info"`
	Model string `eliona:"model" subtype:"info"`

	Group string // Group name used just for pairing

	Lines []Line
	Zones []Zone

	Config *confmodel.Configuration
}

func (d *PeopleCounter) GetName() string {
	return d.Name
}

func (d *PeopleCounter) GetDescription() string {
	return "Xovis People Counter" + d.Name
}

func (d *PeopleCounter) GetAssetType() string {
	return "xovis_people_counter"
}

func (d *PeopleCounter) GetGAI() string {
	return d.GetAssetType() + "_" + d.MAC
}

func (d *PeopleCounter) GetAssetID(projectID string) (*int32, error) {
	return conf.GetAssetId(context.Background(), *d.Config, projectID, d.GetGAI())
}

func (d *PeopleCounter) SetAssetID(assetID int32, projectID string) error {
	if err := conf.InsertAsset(context.Background(), *d.Config, projectID, d.GetGAI(), assetID, d.MAC); err != nil {
		return fmt.Errorf("inserting asset to config db: %v", err)
	}
	return nil
}

func (d *PeopleCounter) GetLocationalChildren() []asset.LocationalNode {
	locationalChildren := make([]asset.LocationalNode, 0, len(d.Lines)+len(d.Zones))
	for i := range d.Lines {
		locationalChildren[i] = &d.Lines[i]
	}
	for i := range d.Zones {
		locationalChildren[i] = &d.Zones[i]
	}
	return locationalChildren
}

func (d *PeopleCounter) GetFunctionalChildren() []asset.FunctionalNode {
	functionalChildren := make([]asset.FunctionalNode, 0, len(d.Lines)+len(d.Zones))
	for i := range d.Lines {
		functionalChildren[i] = &d.Lines[i]
	}
	for i := range d.Zones {
		functionalChildren[i] = &d.Zones[i]
	}
	return functionalChildren
}

type Group struct {
	Name string `eliona:"name" subtype:"info"`

	sensors []PeopleCounter

	Config *confmodel.Configuration
}

func (d *Group) GetName() string {
	return d.Name
}

func (d *Group) GetDescription() string {
	return "Xovis group " + d.Name
}

func (d *Group) GetAssetType() string {
	return "xovis_group"
}

func (d *Group) GetGAI() string {
	return d.GetAssetType() + "_" + d.Name
}

func (d *Group) GetAssetID(projectID string) (*int32, error) {
	return conf.GetAssetId(context.Background(), *d.Config, projectID, d.GetGAI())
}

func (d *Group) SetAssetID(assetID int32, projectID string) error {
	if err := conf.InsertAsset(context.Background(), *d.Config, projectID, d.GetGAI(), assetID, d.Name); err != nil {
		return fmt.Errorf("inserting asset to config db: %v", err)
	}
	return nil
}

func (d *Group) GetLocationalChildren() []asset.LocationalNode {
	locationalChildren := make([]asset.LocationalNode, 0, len(d.sensors))
	for i := range d.sensors {
		locationalChildren[i] = &d.sensors[i]
	}
	return locationalChildren
}

func (d *Group) GetFunctionalChildren() []asset.FunctionalNode {
	functionalChildren := make([]asset.FunctionalNode, 0, len(d.sensors))
	for i := range d.sensors {
		functionalChildren[i] = &d.sensors[i]
	}
	return functionalChildren
}

type Root struct {
	groups []Group

	Config *confmodel.Configuration
}

func (r *Root) GetName() string {
	return "xovis"
}

func (r *Root) GetDescription() string {
	return "Root asset for Xovis devices"
}

func (r *Root) GetAssetType() string {
	return "xovis_root"
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
	locationalChildren := make([]asset.LocationalNode, 0, len(r.groups))
	for i := range r.groups {
		locationalChildren[i] = &r.groups[i]
	}
	return locationalChildren
}

func (r *Root) GetFunctionalChildren() []asset.FunctionalNode {
	functionalChildren := make([]asset.FunctionalNode, 0, len(r.groups))
	for i := range r.groups {
		functionalChildren[i] = &r.groups[i]
	}
	return functionalChildren
}
