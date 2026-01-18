# RaspiWeather

A lightweight weather monitoring service for Raspberry Pi that reads temperature, humidity, and pressure data from a BME280 sensor and exposes it via simple HTTP API and Prometheus metrics.

## Features

- Reads temperature, humidity, and pressure from BME280 sensor(s) via I2C
- Supports multiple BME280 sensors simultaneously
- Exposes Prometheus metrics at `/metrics`
- Provides JSON endpoint at `/sensor-data`
- LCD1602 display support with backlight control
- Configurable data collection interval
- Cross-platform builds for multiple architectures

## Requirements

- Go 1.25.5 or later
- BME280 sensor(s) connected via I2C
- Raspberry Pi (tested on old Model 1B)
- LCD1602 (Optional, for display output)

## Building

```bash
# Build for current platform
make build

# Build for specific architecture
make build-linux-arm5
make build-linux-arm6
make build-linux-arm7
make build-linux-arm64
make build-linux-amd64

# Build for all architectures
make build-all

# Clean build artifacts
make clean
```

## Installation

1. Copy the binary to your Raspberry Pi:
   ```bash
   scp raspiweather-linux-arm7 pi@raspberrypi:/usr/bin/raspiweather
   ```

2. Make it executable (just in case):
   ```bash
   chmod +x /usr/bin/raspiweather
   ```

3. Enable I2C on your Raspberry Pi (if not already enabled):
   ```bash
   sudo raspi-config
   # Navigate to: Interfacing Options -> I2C -> Enable
   ```

## Usage

```bash
# Run with default settings (reads from /dev/i2c-1:0x76 every minute)
./raspiweather

# Run with multiple sensors
./raspiweather -bmeSensors "out:/dev/i2c-1:0x76,in:/dev/i2c-1:0x77"

# Run with LCD display enabled
./raspiweather -lcd -lcdBacklight

# Run with custom interval and HTTP address
./raspiweather -interval 30s -httpAddress ":8080"
```

### Command Line Options

```
$ raspiweather -h
Usage of raspiweather:
  -bmeSensors string
        Comma-separated list of BME280 sensors in format id:devPath:address 
        (e.g., 'out:/dev/i2c-1:0x76,sensor2:/dev/i2c-1:0x77') (default "out:/dev/i2c-1:0x76")
  -httpAddress string
        Address for HTTP Server (default ":9111")
  -interval duration
        Interval of collecting sensors data (default 1m0s)
  -lcd
        Enable LCD1602
  -lcdAddr int
        Address of lcd1602 (default 39)
  -lcdBacklight
        Turn on LCD backlight
  -lcdCols int
        Number of LCD columns (default 16)
  -lcdDevPath string
        Path to i2c lcd device (default "/dev/i2c-1")
  -lcdRows int
        Number of LCD rows (default 2)
  -v    Show version and exit
```

#### Sensor Configuration Format

The `-bmeSensors` flag accepts a comma-separated list of sensors in the format `id:devPath:address`:
- `id`: Unique identifier for the sensor (e.g., "out", "in", "sensor1")
- `devPath`: Path to the I2C device (e.g., "/dev/i2c-1")
- `address`: I2C address in decimal or hexadecimal format (e.g., "0x76" or "118")

Examples:
- Single sensor: `-bmeSensors "out:/dev/i2c-1:0x76"`
- Multiple sensors: `-bmeSensors "out:/dev/i2c-1:0x76,in:/dev/i2c-1:0x77"`

## API Endpoints

### GET /sensor-data

Returns current sensor readings from all configured sensors in JSON format:

```json
[
  {
    "sensor_id": "out",
    "temperature": 25.0,
    "humidity": 44.0,
    "pressure": 1000.0
  },
  {
    "sensor_id": "in",
    "temperature": 22.5,
    "humidity": 50.0,
    "pressure": 1001.2
  }
]
```

### GET /metrics

Prometheus metrics endpoint with the following metrics:

- `sensor_temperature` - Current temperature in Celsius (labeled by sensor_id)
- `sensor_humidity` - Current humidity in percent (labeled by sensor_id)
- `sensor_pressure` - Current pressure in hPa (labeled by sensor_id)

### GET /toggle-leds

Toggles the LCD backlight on/off. Returns "OK" on success. Requires LCD to be enabled.

## Examples

The [examples](examples) directory contains configuration files and templates to help you get started:

- [examples/systemd/raspiweather.service](examples/systemd/raspiweather.service) - Systemd service file for running raspiweather as a system service. 

- [examples/grafana/dashboard.json](examples/grafana/dashboard.json) - Grafana dashboard configuration for visualizing temperature, humidity, and pressure metrics from Prometheus.

## Misc

There is a simple telegram bot for this project: https://github.com/RealFatCat/raspiweatherbot
