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

	"github.com/realfatcat/raspiweather/internal/consumer/meter"
	bme280 "github.com/realfatcat/raspiweather/internal/devices/bme280"
	"github.com/realfatcat/raspiweather/internal/metrics"
	"github.com/realfatcat/raspiweather/internal/producer"
	"github.com/realfatcat/raspiweather/internal/server"
)

var Version string

var (
	interval    = flag.Duration("interval", 1*time.Minute, "Interval of collecting sensors data")
	address     = flag.String("address", ":9111", "Address for HTTP Server")
	i2cDevPath  = flag.String("devPath", bme280.DefaultDevPath, "Path to i2c device")
	bmeAddr     = flag.Int("bme280Addr", bme280.DefaultI2CAddress, "Address of bme280")
	showVersion = flag.Bool("v", false, "Show version and exit")
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

	bus, err := bme280.New(*i2cDevPath, *bmeAddr)
	if err != nil {
		slog.Error("Starting BME280", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := bus.Close(); err != nil {
			slog.Error("Closing BME280", "error", err)
		}
	}()

	dataProducer := producer.New(bus, *interval)
	ch := dataProducer.Produce(ctx)

	m := meter.New(sensorMetrics)
	m.Consume(ctx, ch)

	httpSrv := server.NewHTTP(*address, dataProducer)
	if err := httpSrv.ListenAndServe(ctx); err != nil {
		slog.Error("HTTP server", "error", err)
		os.Exit(1)
	}
}
