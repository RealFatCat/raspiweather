package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/realfatcat/fanoutsub"
	lcd1602 "github.com/realfatcat/lcd1602/pkg/lcd"
	"github.com/realfatcat/raspiweather/internal/consumer/lcd"
	"github.com/realfatcat/raspiweather/internal/consumer/meter"
	bme280 "github.com/realfatcat/raspiweather/internal/devices/bme280"
	"github.com/realfatcat/raspiweather/internal/metrics"
	"github.com/realfatcat/raspiweather/internal/producer"
	"github.com/realfatcat/raspiweather/internal/server"
	"github.com/realfatcat/raspiweather/internal/types"
)

var Version string

var (
	interval      = flag.Duration("interval", 1*time.Minute, "Interval of collecting sensors data")
	address       = flag.String("address", ":9111", "Address for HTTP Server")
	i2cBMEDevPath = flag.String("bmeDevPath", bme280.DefaultDevPath, "Path to i2c bme device")
	i2cLCDDevPath = flag.String("lcdDevPath", lcd1602.DefaultDevice, "Path to i2c lcd device")
	bmeAddr       = flag.Int("bme280Addr", bme280.DefaultI2CAddress, "Address of bme280")
	lcdAddr       = flag.Int("lcdAddr", lcd1602.DefaultAddress, "Address of lcd1602")
	lcdColumns    = flag.Int("lcdCols", 16, "Number of LCD columns")
	lcdRows       = flag.Int("lcdRows", 2, "Number of LCD rows")
	lcdBacklight  = flag.Bool("lcdBacklight", false, "Turn on LCD backlight")
	lcdEnabled    = flag.Bool("lcd", false, "Enable LCD1602")
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

	bmeBus, err := bme280.New(*i2cBMEDevPath, *bmeAddr)
	if err != nil {
		slog.Error("Starting BME280", "error", err)
		os.Exit(1)
	}
	defer func() { bmeBus.Close() }()

	dataProducer := producer.New(bmeBus, *interval)
	ch := dataProducer.Produce(ctx)

	fan := fanoutsub.New(ch)

	meterCh := make(chan types.WeatherData, 1)
	fan.Subscribe(meterCh)
	m := meter.New(sensorMetrics)
	m.Consume(ctx, meterCh)

	if *lcdEnabled {
		lcdDev, err := lcd1602.New(*i2cLCDDevPath, *lcdAddr, *lcdColumns, *lcdRows, *lcdBacklight)
		if err != nil {
			slog.Error("Starting LCD1602", "error", err)
			os.Exit(1)
		}

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

	httpSrv := server.NewHTTP(*address, dataProducer)
	if err := httpSrv.ListenAndServe(ctx); err != nil {
		slog.Error("HTTP server", "error", err)
		os.Exit(1)
	}
}
