wh31e
=========

[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/akerl/wh31e/build.yml?branch=main)](https://github.com/akerl/wh31e/actions)
[![GitHub release](https://img.shields.io/github/release/akerl/wh31e.svg)](https://github.com/akerl/wh31e/releases)
[![MIT Licensed](https://img.shields.io/badge/license-MIT-green.svg)](https://tldrlegal.com/license/mit-license)

Transform [WH31E](https://www.amazon.com/dp/B01MG4HW8C/) readings from [rtl_433](https://github.com/merbanan/rtl_433) for publishing as Prometheus-format metrics.

## Usage

Configue rtl_433 to emit syslog events with `-F syslog:127.0.0.1:1514` (to send to wh31e on localhost port 1514). In the rtl_433 config file, this is configured with `output syslog:127.0.0.1:1514`.

To configure wh31e, create a YAML file:

```
syslog_host: 127.0.0.1
syslog_port: 1514
port: 9131
sensor_names:
  SENSOR_ID: HUMAN_NAME
```

Check the output from rtl_433 to get the sensor IDs for each sensor (this is a separate integer from the channel ID).

Then, start the forwarder with `./wh31e serve config.yaml`

## Installation

```
git clone https://github.com/akerl/wh31e
cd wh31e
go build .
```

## License

wh31e is released under the MIT License. See the bundled LICENSE file for details.
