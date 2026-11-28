package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/realfatcat/fanoutsub"
	"github.com/realfatcat/lcd1602"
	consumer_lcd "github.com/realfatcat/raspiweather/internal/consumer/lcd"
	"github.com/realfatcat/raspiweather/internal/consumer/meter"
	bme280 "github.com/realfatcat/raspiweather/internal/devices/bme280"
	"github.com/realfatcat/raspiweather/internal/devices/lcd"
	"github.com/realfatcat/raspiweather/internal/metrics"
	"github.com/realfatcat/raspiweather/internal/producer"
	"github.com/realfatcat/raspiweather/internal/server"
	"github.com/realfatcat/raspiweather/pkg/types"
)

// Version is injected at build time via -ldflags (see Makefile).
// Falls back to "dev" for plain `go build`.
var Version = "dev"

var (
	interval         = flag.Duration("interval", 1*time.Minute, "Interval at which all sensors are read; the first read happens immediately at startup")
	httpAddress      = flag.String("httpAddress", ":9111", "Address for HTTP Server")
	i2cLCDDevPath    = flag.String("lcdDevPath", lcd1602.DefaultDevice, "Path to i2c lcd device")
	lcdAddr          = flag.Int("lcdAddr", lcd1602.DefaultAddress, "Address of lcd1602")
	lcdColumns       = flag.Int("lcdCols", 16, "Number of LCD columns")
	lcdRows          = flag.Int("lcdRows", 2, "Number of LCD rows")
	lcdBacklight     = flag.Bool("lcdBacklight", false, "Turn on LCD backlight")
	lcdEnabled       = flag.Bool("lcd", false, "Enable LCD1602")
	bmeSensors       = flag.String("bmeSensors", "out:/dev/i2c-1:0x76", "Comma-separated list of BME280 sensors in format id:devPath:address (e.g., 'sensor1:/dev/i2c-1:0x76,sensor2:/dev/i2c-1:0x77')")
	font             = flag.Uint("font", lcd1602.Font5x8, "lcd font, possible values 0 for 5x8 and 4 for 5x10")
	showVersion      = flag.Bool("v", false, "Show version and exit")
	rotationInterval = flag.Duration("rotationInterval", 10*time.Second, "Rotation interval of sensors data displayed on screen")
	lcdStaleWindow   = flag.Duration("lcdStaleWindow", 1*time.Minute, "Data older than this window is shown as \"No data\" on the LCD")
)

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("%s\n", Version)
		return
	}

	if err := run(); err != nil {
		slog.Error("Failed to start", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if *interval <= 0 {
		return fmt.Errorf("invalid -interval %v: must be greater than 0", *interval)
	}
	if *rotationInterval <= 0 {
		return fmt.Errorf("invalid -rotationInterval %v: must be greater than 0", *rotationInterval)
	}
	if *lcdStaleWindow <= 0 {
		return fmt.Errorf("invalid -lcdStaleWindow %v: must be greater than 0", *lcdStaleWindow)
	}

	fontByte, err := parseFont(*font)
	if err != nil {
		return fmt.Errorf("invalid -font: %w", err)
	}

	sensorMetrics, err := metrics.NewSensorsMetrics(prometheus.DefaultRegisterer)
	if err != nil {
		return fmt.Errorf("init metrics: %w", err)
	}

	// Parse sensors and open buses.
	sensorConfigs, err := initSensorConfigs(bmeSensors)
	if err != nil {
		return fmt.Errorf("init sensor configs: %w", err)
	}
	defer func() {
		for i := range sensorConfigs {
			_ = sensorConfigs[i].Bus.Close()
		}
	}()

	// Create multi producer
	dataProducer := producer.NewMultiProducer(sensorConfigs, *interval)

	ch := dataProducer.Produce(ctx)

	fan := fanoutsub.New(ch)

	var backlightToggler server.BacklightToggler
	var lcdDone <-chan struct{}
	var meterDone <-chan struct{}
	var started bool

	meterCh := make(chan types.WeatherData, 1)
	if err := fan.Subscribe(meterCh); err != nil {
		return fmt.Errorf("subscribe meter consumer: %w", err)
	}
	m := meter.New(sensorMetrics)
	meterDone = m.Consume(ctx, meterCh)

	if *lcdEnabled {
		lcdDev, err := lcd.New(*i2cLCDDevPath, *lcdAddr, *lcdColumns, *lcdRows, fontByte, *lcdBacklight)
		if err != nil {
			return fmt.Errorf("starting LCD1602: %w", err)
		}
		defer func() {
			_ = lcdDev.Clear()
			_ = lcdDev.Close()
		}()
		backlightToggler = lcdDev

		lcdCh := make(chan types.WeatherData, 1)
		if err := fan.Subscribe(lcdCh); err != nil {
			return fmt.Errorf("subscribe LCD consumer: %w", err)
		}
		lcdConsumer := consumer_lcd.New(lcdDev, *rotationInterval, *lcdStaleWindow)
		lcdDone = lcdConsumer.Consume(ctx, lcdCh)
	}

	// Registered last so it runs first.
	// Cancel the context and let the fanout drain and the consumers (meter, LCD),
	// finish before any resource is closed.
	// Done is only awaited if the fanout actually started,
	// otherwise fanout never ran and Done is never closed.
	defer func() {
		stop()
		if started {
			<-fan.Done()
		}
		<-meterDone
		if lcdDone != nil {
			<-lcdDone
		}
	}()

	if err := fan.Start(ctx); err != nil {
		return fmt.Errorf("starting fanout: %w", err)
	}
	started = true

	httpSrv := server.NewHTTP(*httpAddress, dataProducer, backlightToggler)
	if err := httpSrv.ListenAndServe(ctx); err != nil {
		return fmt.Errorf("HTTP server: %w", err)
	}
	return nil
}

func parseFont(font uint) (byte, error) {
	if slices.Contains(lcd.AcceptableFonts, font) {
		return byte(font), nil
	}

	return 0, fmt.Errorf("invalid font value %d, acceptable: %v", font, lcd.AcceptableFonts)
}

// initSensorConfigs parses sensor configurations and opens buses.
// On a failure after some buses were opened, already-opened buses are closed.
func initSensorConfigs(bmeSensors *string) ([]producer.SensorConfig, error) {
	if *bmeSensors == "" {
		return nil, fmt.Errorf("bmeSensors flag is not set")
	}

	type sensorSpec struct {
		id      string
		devPath string
		addr    int
	}

	var specs []sensorSpec
	seenIDs := make(map[string]bool)

	for part := range strings.SplitSeq(*bmeSensors, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Format: id:devPath:address
		fields := strings.Split(part, ":")
		if len(fields) != 3 {
			return nil, fmt.Errorf("invalid sensor format %q, expected id:devPath:address", part)
		}

		sensorID := strings.TrimSpace(fields[0])
		devPath := strings.TrimSpace(fields[1])
		addrStr := strings.TrimSpace(fields[2])

		if sensorID == "" {
			return nil, fmt.Errorf("sensor id is empty, expected id:devPath:address")
		}
		if devPath == "" {
			return nil, fmt.Errorf("devPath is empty for sensor_id %s, expected id:devPath:address", sensorID)
		}

		if seenIDs[sensorID] {
			return nil, fmt.Errorf("duplicate sensor id %q", sensorID)
		}
		seenIDs[sensorID] = true

		addr, err := parseAddr(addrStr)
		if err != nil {
			return nil, fmt.Errorf("parsing address for sensor_id %s: %w", sensorID, err)
		}

		specs = append(specs, sensorSpec{id: sensorID, devPath: devPath, addr: addr})
	}

	if len(specs) == 0 {
		return nil, fmt.Errorf("no sensors configured")
	}

	var sensorConfigs []producer.SensorConfig
	for _, spec := range specs {
		bus, err := bme280.New(spec.devPath, spec.addr)
		if err != nil {
			for _, cfg := range sensorConfigs {
				_ = cfg.Bus.Close()
			}
			return nil, fmt.Errorf("starting BME280, sensor_id %s: %w", spec.id, err)
		}

		sensorConfigs = append(sensorConfigs, producer.SensorConfig{
			ID:  spec.id,
			Bus: bus,
		})
	}
	return sensorConfigs, nil
}

// parseAddr parses I2C address (decimal, or hex with 0x/0X prefix).
func parseAddr(inAddr string) (int, error) {
	s := strings.TrimSpace(inAddr)
	if s == "" {
		return 0, fmt.Errorf("address is empty")
	}

	base := 10
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		base = 16
		s = s[2:]
	}
	if s == "" {
		return 0, fmt.Errorf("invalid address %q: no digits", inAddr)
	}
	if strings.ContainsAny(s, "+-_ ") {
		return 0, fmt.Errorf("invalid address %q: only digits are allowed", inAddr)
	}

	addr64, err := strconv.ParseInt(s, base, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid address %q: %w", inAddr, err)
	}
	if addr64 < 0 || addr64 > 0x7F {
		return 0, fmt.Errorf("address 0x%02x out of range, expected 0x00-0x7F", addr64)
	}
	return int(addr64), nil
}
