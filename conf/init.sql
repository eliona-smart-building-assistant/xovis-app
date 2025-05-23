--  This file is part of the Eliona project.
--  Copyright © 2024 IoTEC AG. All Rights Reserved.
--  ______ _ _
-- |  ____| (_)
-- | |__  | |_  ___  _ __   __ _
-- |  __| | | |/ _ \| '_ \ / _` |
-- | |____| | | (_) | | | | (_| |
-- |______|_|_|\___/|_| |_|\__,_|
--
--  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
--  BUT NOT LIMITED  TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
--  NON INFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
--  DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
--  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

create schema if not exists xovis2;

-- Should be editable by eliona frontend.
create table if not exists xovis2.configuration
(
	id                   bigserial primary key,
	check_certificate    boolean not null,
	refresh_interval     integer not null default 60,
	request_timeout      integer not null default 120,
	active               boolean not null default false,
	enable               boolean not null default false,
	project_ids          text[] not null,
	user_id              text not null
);

-- Should be editable by eliona frontend.
create table if not exists xovis2.sensor
(
	id                   bigserial primary key,
	configuration_id     bigserial not null references xovis2.configuration(id) ON DELETE CASCADE,
	username             text not null,
	password             text not null,
	hostname             text not null,
	port                 int not null,

	discovery_mode       TEXT NOT NULL CHECK (discovery_mode IN ('disabled', 'L2', 'L3')),
	l3_first_ip          TEXT,  -- For L3 discovery, the starting IP to scan
	l3_count             INT,   -- For L3 discovery, the number of IPs to scan

	mac_address text unique
);

create table if not exists xovis2.asset
(
	id               bigserial primary key,
	configuration_id bigserial not null references xovis2.configuration(id) ON DELETE CASCADE,
	project_id       text      not null,
	global_asset_id  text      not null,
	provider_id      text      not null,
	asset_id         integer
);

-- There is a transaction started in app.Init(). We need to commit to make the
-- new objects available for all other init steps.
-- Chain starts the same transaction again.
commit and chain;
