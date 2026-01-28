package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/realfatcat/fanoutsub"
	"github.com/realfatcat/lcd1602"
	"github.com/realfatcat/raspiweather/internal/consumer/lcd"
	"github.com/realfatcat/raspiweather/internal/consumer/meter"
	bme280 "github.com/realfatcat/raspiweather/internal/devices/bme280"
	"github.com/realfatcat/raspiweather/internal/metrics"
	"github.com/realfatcat/raspiweather/internal/producer"
	"github.com/realfatcat/raspiweather/internal/server"
	"github.com/realfatcat/raspiweather/pkg/types"
)

var Version string

var (
	interval      = flag.Duration("interval", 1*time.Minute, "Interval of collecting sensors data")
	httpAddress   = flag.String("httpAddress", ":9111", "Address for HTTP Server")
	i2cLCDDevPath = flag.String("lcdDevPath", lcd1602.DefaultDevice, "Path to i2c lcd device")
	lcdAddr       = flag.Int("lcdAddr", lcd1602.DefaultAddress, "Address of lcd1602")
	lcdColumns    = flag.Int("lcdCols", 16, "Number of LCD columns")
	lcdRows       = flag.Int("lcdRows", 2, "Number of LCD rows")
	lcdBacklight  = flag.Bool("lcdBacklight", false, "Turn on LCD backlight")
	lcdEnabled    = flag.Bool("lcd", false, "Enable LCD1602")
	bmeSensors    = flag.String("bmeSensors", "out:/dev/i2c-1:0x76", "Comma-separated list of BME280 sensors in format id:devPath:address (e.g., 'sensor1:/dev/i2c-1:0x76,sensor2:/dev/i2c-1:0x77')")
	font          = flag.Uint("font", lcd1602.Font5x8, "lcd font, possible values 0 for 5x8 and 4 for 5x10")
	showVersion   = flag.Bool("v", false, "Show version and exit")
)

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("%s\n", Version)
		os.Exit(0)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	sensorMetrics := metrics.NewSensorsMetrics()

	var sensorConfigs []producer.SensorConfig

	// Parse multiple sensors from flag, init buses.
	sensorConfigs, err := initSensorConfigs(bmeSensors)
	if err != nil {
		slog.Error("Init Sensor Config", "error", err)
		os.Exit(1)
	}

	// Create multi producer
	dataProducer := producer.NewMultiProducer(sensorConfigs, *interval)

	ch := dataProducer.Produce(ctx)

	fan := fanoutsub.New(ch)

	meterCh := make(chan types.WeatherData, 1)
	fan.Subscribe(meterCh)
	m := meter.New(sensorMetrics)
	m.Consume(ctx, meterCh)

	var backlightToggler server.BacklightToggler
	if *lcdEnabled {
		if *font != lcd1602.Font5x8 && *font != lcd1602.Font5x10 {
			slog.Error("Invalid font", "value", *font, "acceptable", []int{lcd1602.Font5x8, lcd1602.Font5x10})
			os.Exit(1)
		}
		font := byte(*font)
		lcdDev, err := lcd1602.New(*i2cLCDDevPath, *lcdAddr, *lcdColumns, *lcdRows, font, *lcdBacklight)
		if err != nil {
			slog.Error("Starting LCD1602", "error", err)
			os.Exit(1)
		}
		backlightToggler = lcdDev

		lcdCh := make(chan types.WeatherData, 1)
		fan.Subscribe(lcdCh)
		lcdConsumer := lcd.New(lcdDev)
		lcdConsumer.Consume(ctx, lcdCh)

		defer func() {
			_ = lcdDev.Clear()
			_ = lcdDev.Close()
		}()
	}

	fan.Start(ctx)

	httpSrv := server.NewHTTP(*httpAddress, dataProducer, backlightToggler)
	if err := httpSrv.ListenAndServe(ctx); err != nil {
		slog.Error("HTTP server", "error", err)
		os.Exit(1)
	}
}

// initSensorConfigs parses sensor configurations, create buses.
func initSensorConfigs(bmeSensors *string) ([]producer.SensorConfig, error) {
	if *bmeSensors == "" {
		return nil, fmt.Errorf("bmeSensors flag is not set")
	}

	var sensorConfigs []producer.SensorConfig

	for part := range strings.SplitSeq(*bmeSensors, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Format: id:devPath:address
		fields := strings.Split(part, ":")
		if len(fields) != 3 {
			slog.Error("Invalid sensor format", "sensor", part, "expected_format", "id:devPath:address")
			os.Exit(1)
		}

		sensorID := strings.TrimSpace(fields[0])
		devPath := strings.TrimSpace(fields[1])
		addrStr := strings.TrimSpace(fields[2])

		addr, err := parseAddr(addrStr)
		if err != nil {
			return nil, err
		}

		bus, err := bme280.New(devPath, addr)
		if err != nil {
			return nil, fmt.Errorf("starting BME280, sensor_id %s: %w", sensorID, err)
		}

		sensorConfigs = append(sensorConfigs, producer.SensorConfig{
			ID:  sensorID,
			Bus: bus,
		})
	}
	return sensorConfigs, nil
}

// parseAddress parses strings address to int, support both decimal and hex
func parseAddr(inAddr string) (int, error) {
	if strings.HasPrefix(inAddr, "0x") || strings.HasPrefix(inAddr, "0X") {
		addr64, err := strconv.ParseInt(inAddr[2:], 16, 64)
		if err != nil {
			return 0, err
		}
		return int(addr64), nil
	}

	addr64, err := strconv.ParseInt(inAddr, 10, 64)
	if err != nil {
		return 0, err
	}
	return int(addr64), nil
}
