# Xovis User Guide

### Introduction

> The Xovis app provides integration and data gathering from Xovis People Counter sensors to Eliona.

## Overview

This guide provides instructions on configuring, installing, and using the Xovis app to manage resources and synchronize data between Eliona and Xovis People counters.

## Installation

Install the Xovis app via the Eliona App Store.

## Configuration

The Xovis app requires configuration through Eliona’s settings interface. Below are the general steps and details needed to configure the app effectively.

### Registering the app in Xovis Service

Create credentials in Xovis Service to connect the Xovis services from Eliona. All required credentials are listed below in the [configuration section](#configure-the-xovis-app).  

<mark>TODO: Describe the steps where you can get or create the necessary credentials.</mark> 

## Xovis App Configuration and Sensor Discovery Workflow

Configurations can be created in Eliona under `Apps > Xovis > Settings` which opens the app's [Generic Frontend](https://doc.eliona.io/collection/v/eliona-english/manuals/settings/apps). Below is the complete workflow to guide you through configuring Xovis devices and sensors.

---

### Step 1: Create a Xovis App Configuration

First, create a configuration that defines how Eliona interacts with the Xovis device. This configuration contains important details such as whether the device certificate should be verified, the frequency of data collection, and API timeouts.

**Endpoint**: `/configs`

**Method**: `POST`

**Required Data**:
| Attribute          | Description                                                                                                                                                                   |
|--------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `checkCertificate` | Specifies whether the device certificate should be verified (`true` for publicly accessible devices, `false` for devices that are not publicly accessible).                    |
| `enable`           | Flag to enable or disable data synchronization for this configuration.                                                                                                         |
| `refreshInterval`  | Interval in seconds for collecting data from the Xovis device (default: 60 seconds).                                                                                           |
| `requestTimeout`   | Timeout in seconds for the API request to the Xovis device (default: 120 seconds).                                                                                             |
| `projectIDs`       | List of Eliona project IDs for which this device should collect data. For each project ID, smart devices are automatically created as assets in Eliona.                         |

### Example Configuration Request:

```json
{
  "checkCertificate": true,
  "enable": true,
  "refreshInterval": 60,
  "requestTimeout": 120,
  "projectIDs": ["10"]
}
```

### Response:

Once the configuration is successfully created, you will receive a response with the internal ID (`id`) of the newly created configuration. This ID is important as it will be used when configuring sensors in the next step.

**Example Response**:

```json
{
  "id": 1,
  "enable": true,
  "refreshInterval": 60,
  "requestTimeout": 120,
  "active": true,
  "projectIDs": ["10"],
  "userId": "585"
}
```

You will see the above JSON on the other side of your screen, which includes the `id` field that is automatically generated. **You need to use this `id` when configuring sensors** in the next step.

---

### Step 2: Add Xovis Sensors Using the Configuration ID

With the configuration ID from the previous step (e.g., `"id": 1`), you can now proceed to configure Xovis sensors. Each sensor is associated with a configuration and supports discovery methods such as Layer 2 (L2) or Layer 3 (L3).

**Endpoint**: `/sensors`

**Method**: `POST`

**Required Data**:
| Attribute          | Description                                                                                                                                                                   |
|--------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `configuration_id` | The ID of the configuration created in the previous step. This associates the sensor with the configuration.                                                                  |
| `username`         | Xovis sensor username.                                                                                                                                                         |
| `password`         | Xovis sensor password.                                                                                                                                                         |
| `hostname`         | Hostname or IP address of the Xovis sensor.                                                                                                                                    |
| `port`             | Network port used to communicate with the sensor (e.g., 443 for HTTPS).                                                                                                        |
| `discovery_mode`   | The mode used for discovering other sensors (`disabled`, `L2`, or `L3`).                                                                                                       |
| `l3_first_ip`      | (Optional) The starting IP address for Layer 3 (L3) discovery, used to find Xovis sensors in a specific range.                                                                 |
| `l3_count`         | (Optional) The number of IP addresses to scan when using L3 discovery mode.                                                                                                     |

### Example Sensor Request:

Use the configuration `id` (e.g., `1`) from the previous step when posting the sensor configuration.

```json
{
  "configuration_id": 1,
  "username": "sensor_user",
  "password": "securepassword123",
  "hostname": "sensor1.local",
  "port": 8080,
  "discovery_mode": "L3",
  "l3_first_ip": "192.168.1.10",
  "l3_count": 50
}
```

### Response:

Upon successfully creating the sensor, the system will return the details of the created sensor, allowing you to monitor and manage it.

---

### Continuous Asset Creation (CAC)

Once the configuration and sensor discovery settings are complete, Eliona will begin Continuous Asset Creation (CAC). Discovered sensors will be automatically added as assets in Eliona, and the following will occur:

- **Automatic Asset Creation**: Sensors identified through discovery (L2 or L3) will be automatically added to Eliona as assets.
- **Notifications**: Users will be notified through Eliona’s notification system when new assets (sensors) are created, ensuring that newly discovered sensors are visible and actionable.

---

### Summary Workflow

1. **Create Configuration**: POST to `/configs`, receive a configuration ID in the response.
2. **Add Sensors**: Use the configuration ID when posting sensor data to `/sensors`.
3. **Asset Creation**: Eliona automatically creates assets and notifies you when new sensors are discovered.

#### **Handling NAT and Address Modifications**

- **Outside NAT**: If the application is running outside the company's NAT, discovered addresses and ports may not correspond to the correct external addresses.
- **Address Modifications**: Adjust the app configuration to reflect the correct external addresses if necessary.

#### **Password Management**

- **Default Passwords**: The application assumes that all discovered devices use the same password as the dicsovering device.
- **Individual Passwords**: If devices use unique passwords, update the configuration for each device accordingly.

### Dashboard templates

The app offers a predefined dashboard that clearly displays the most important information. YOu can create such a dashboard under `Dashboards > Copy Dashboard > From App > Xovis`.

### <mark>TODO: Other features</mark>
