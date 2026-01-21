package lcd

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/realfatcat/raspiweather/pkg/types"
)

type LCD struct {
	printer          Printer
	rotationInterval time.Duration
	storage          *storage
}

func New(printer Printer, rotationInterval time.Duration) *LCD {
	return &LCD{
		printer:          printer,
		rotationInterval: rotationInterval,
		storage:          newStorage(),
	}
}

func (l *LCD) Consume(ctx context.Context, ch <-chan types.WeatherData) {
	exitCh := make(chan struct{}, 1)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case sd, ok := <-ch:
				if !ok {
					// Channel closed, exit gracefully
					exitCh <- struct{}{}
					return
				}
				l.storage.put(sd.SensorID, sd)
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(l.rotationInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-exitCh:
				return
			case <-ticker.C:
				wd, ok := l.storage.getNext()
				if !ok {
					break
				}

				if err := l.printer.Print(fmt.Sprintf("T:%.1fC H:%.1f%%", wd.Temperature, wd.Humidity), 0, 0); err != nil {
					slog.Error("Writing Temperature and Humidity to LCD", "error", err)
				}

				if err := l.printer.Print(fmt.Sprintf("P:%.1fmmHg %.*s", wd.Pressure*0.75, 3, wd.SensorID), 1, 0); err != nil {
					slog.Error("Writing Pressure to LCD", "error", err)
				}
			}
		}
	}()
}
