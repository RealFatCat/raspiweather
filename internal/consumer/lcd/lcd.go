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
	print            func(types.WeatherData)
}

func New(printer Printer, rotationInterval time.Duration) *LCD {
	lcd := &LCD{
		printer:          printer,
		rotationInterval: rotationInterval,
		storage:          newStorage(),
	}

	switch printer.Rows() {
	case 1:
		lcd.print = lcd.printOneRow
	case 4:
		lcd.print = lcd.printFourRows
	default:
		lcd.print = lcd.printTwoRows
	}

	return lcd
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
				l.print(wd)
			}
		}
	}()
}

func (l *LCD) printOneRow(wd types.WeatherData) {
	if err := l.printer.Print(
		fmt.Sprintf("%.*s T%.0f H%.0f P%.0f", 3, wd.SensorID, wd.Temperature, wd.Humidity, wd.Pressure*0.75), 0, 0); err != nil {
		slog.Error("Writing sensor data to LCD", "error", err)
	}
}

func (l *LCD) printTwoRows(wd types.WeatherData) {
	if err := l.printer.Print(fmt.Sprintf("T:%.1fC H:%.1f%%", wd.Temperature, wd.Humidity), 0, 0); err != nil {
		slog.Error("Writing Temperature and Humidity to LCD", "error", err)
	}

	if err := l.printer.Print(fmt.Sprintf("P:%.1fmm [%.*s]", wd.Pressure*0.75, 3, wd.SensorID), 1, 0); err != nil {
		slog.Error("Writing Pressure to LCD", "error", err)
	}
}

func (l *LCD) printFourRows(wd types.WeatherData) {
	idLenLimit := 12
	if l.printer.Columns() == 20 {
		idLenLimit = 16
	}

	if err := l.printer.Print(fmt.Sprintf("ID: %.*s", idLenLimit, wd.SensorID), 0, 0); err != nil {
		slog.Error("Writing SensorID to LCD", "error", err)
	}

	if err := l.printer.Print(fmt.Sprintf("Temp: %.2fC", wd.Temperature), 1, 0); err != nil {
		slog.Error("Writing Temperature to LCD", "error", err)
	}

	if err := l.printer.Print(fmt.Sprintf("Humid: %.2f%%", wd.Humidity), 2, 0); err != nil {
		slog.Error("Writing Humidity to LCD", "error", err)
	}

	if err := l.printer.Print(fmt.Sprintf("Pressure: %.2fmm", wd.Pressure*0.75), 3, 0); err != nil {
		slog.Error("Writing Pressure to LCD", "error", err)
	}
}
