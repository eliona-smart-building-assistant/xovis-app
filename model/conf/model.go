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

package confmodel

type Configuration struct {
	ID               int64
	CheckCertificate bool
	RefreshInterval  int32
	RequestTimeout   int32
	Enable           bool
	Active           bool
	ProjectIDs       []string
	UserId           string
}

type Sensor struct {
	ID       int64
	Config   Configuration
	Username string
	Password string
	Hostname string
	Port     int32

	DiscoveryMode string
	L3FirstIP     *string
	L3Count       *int32

	MACAddress *string
}

type Asset struct {
	ID            int64
	Config        Configuration
	ProjectID     string
	GlobalAssetID string
	ProviderID    string
	AssetID       int32
}
