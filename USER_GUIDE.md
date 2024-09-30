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

### Configure the Xovis app 

Configurations can be created in Eliona under `Apps > Xovis > Settings` which opens the app's [Generic Frontend](https://doc.eliona.io/collection/v/eliona-english/manuals/settings/apps). Here you can use the appropriate endpoint with the POST method. Each configuration requires the following data:

| Attribute         | Description                                                                     |
|-------------------|---------------------------------------------------------------------------------|
| `username`         | Xovis device username.                                                   |
| `password`   | Xovis device password                              |
| `hostname`     | Xovis device hostname |
| `port`     | Xovis device port |
| `checkCertificate`     | Specifies whether the device certificate should be verified (should be `true` for devices publicly accessible, can be `false` for devices inaccessible from the Internet). |
| `enable`          | Flag to enable or disable this configuration.                                   |
| `refreshInterval` | Interval in seconds for data synchronization.                                   |
| `requestTimeout`  | API query timeout in seconds.                                                   |
| `projectIDs`      | List of Eliona project IDs for data collection.                                 |

Example configuration JSON:

```json
{
  "username": "Username",
  "password": "Password",
  "hostname": "192.168.1.x",
  "port": 443,
  "checkCertificate": true,
  "enable": true,
  "refreshInterval": 60,
  "requestTimeout": 120,
  "projectIDs": [
    "10"
  ]
}
```

## Continuous Asset Creation

### Xovis Sensor Discovery and Configuration

To add Xovis sensors to the Eliona application, follow these guidelines:

#### **Initial Configuration**

1. **Manual Sensor Configuration**:
   - At least one Xovis sensor must be configured manually in the Eliona app. This sensor will serve as the discovery device, responsible for finding other Xovis sensors on the network.

#### **Sensor Discovery Methods**

Xovis sensors support two methods for discovering other sensors on the network:

1. **Layer 2 (L2) Discovery**:
   - **Description**: This method scans the sensor’s own subnet (e.g., 192.168.1.0/24) to find other Xovis devices.
   - **Configuration**: No additional configuration is required for L2 discovery.

2. **Layer 3 (L3) Discovery**:
   - **Description**: This method performs an active scan across a specified IP range to identify Xovis devices. It is useful for discovering devices beyond the local subnet.
   - **Configuration**: Requires configuration of the following:
     - **First IP Address**: The starting IP address for the scan.
     - **IP Count**: The number of IP addresses to scan.

#### **Continuous Asset Creation (CAC)**

- **Automatic Asset Creation**: Once the discovery configuration is set, the application will initiate Continuous Asset Creation (CAC). Discovered sensors are automatically added as assets in Eliona.
- **Notifications**: Users will be notified through Eliona’s notification system when new assets are created.

#### **Handling NAT and Address Modifications**

- **Outside NAT**: If the application is running outside the company's NAT, discovered addresses and ports may not correspond to the correct external addresses.
- **Address Modifications**: Adjust the app configuration to reflect the correct external addresses if necessary.

#### **Password Management**

- **Default Passwords**: The application assumes that all discovered devices use the same password as the dicsovering device.
- **Individual Passwords**: If devices use unique passwords, update the configuration for each device accordingly.

### Dashboard templates

The app offers a predefined dashboard that clearly displays the most important information. YOu can create such a dashboard under `Dashboards > Copy Dashboard > From App > Xovis`.

### <mark>TODO: Other features</mark>
