# Xovis User Guide

### Introduction

> The Xovis app provides integration and synchronization between Eliona and Xovis services.

## Overview

This guide provides instructions on configuring, installing, and using the Xovis app to manage resources and synchronize data between Eliona and Xovis services.

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
| `baseURL`         | URL of the Xovis services.                                                   |
| `clientSecrets`   | Client secrets obtained from the Xovis service.                              |
| `assetFilter`     | Filtering asset during [Continuous Asset Creation](#continuous-asset-creation). |
| `enable`          | Flag to enable or disable this configuration.                                   |
| `refreshInterval` | Interval in seconds for data synchronization.                                   |
| `requestTimeout`  | API query timeout in seconds.                                                   |
| `projectIDs`      | List of Eliona project IDs for data collection.                                 |

Example configuration JSON:

```json
{
  "baseURL": "http://service/v1",
  "clientSecrets": "random-cl13nt-s3cr3t",
  "filter": "",
  "enable": true,
  "refreshInterval": 60,
  "requestTimeout": 120,
  "projectIDs": [
    "10"
  ]
}
```

## Continuous Asset Creation

Once configured, the app starts Continuous Asset Creation (CAC). Discovered resources are automatically created as assets in Eliona, and users are notified via Eliona’s notification system.

<mark>TODO: Describe what resources are created, the hierarchy and the data points.</mark>

## Additional Features

<mark>TODO: Describe all other features of the app.</mark>

### Dashboard templates

The app offers a predefined dashboard that clearly displays the most important information. YOu can create such a dashboard under `Dashboards > Copy Dashboard > From App > Xovis`.

### <mark>TODO: Other features</mark>
