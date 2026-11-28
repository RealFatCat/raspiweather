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
	stalenessWindow  time.Duration
	storage          *storage
	print            func(types.WeatherData)
}

// New creates an LCD consumer. stalenessWindow must be greater than zero:
// data older than this window is considered stale and the screen displays
// "No data" for that sensor. It is independent of the producer's collection
// interval.
func New(printer Printer, rotationInterval time.Duration, stalenessWindow time.Duration) *LCD {
	lcd := &LCD{
		printer:          printer,
		rotationInterval: rotationInterval,
		stalenessWindow:  stalenessWindow,
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

func (l *LCD) Consume(ctx context.Context, ch <-chan types.WeatherData) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(l.rotationInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case sd, ok := <-ch:
				if !ok {
					return
				}
				isFirst := l.storage.isEmpty()
				l.storage.put(sd.SensorID, sd)
				if isFirst {
					l.render(sd)
				}
			case <-ticker.C:
				wd, ok := l.storage.getNext()
				if !ok {
					continue
				}
				if time.Since(wd.Timestamp) > l.stalenessWindow {
					l.showNoData(wd.SensorID)
					continue
				}
				l.render(wd)
			}
		}
	}()
	return done
}

func (l *LCD) render(wd types.WeatherData) {
	if err := l.printer.Clear(); err != nil {
		slog.Error("Clearing LCD", "error", err)
	}
	l.print(wd)
}

func (l *LCD) showNoData(sensorID string) {
	if err := l.printer.Clear(); err != nil {
		slog.Error("Clearing LCD", "error", err)
	}
	const suffix = ": No data"
	// Truncate the id so the whole line fits the LCD and the suffix stays visible.
	// idLen is not printed; it is only a width budget.
	idLen := max(l.printer.Columns()-len(suffix), 1)
	if err := l.printer.Print(truncateRunes(sensorID, idLen)+suffix, 0, 0); err != nil {
		slog.Error("Writing no-data message to LCD", "error", err)
	}
}

func (l *LCD) printOneRow(wd types.WeatherData) {
	if err := l.printer.Print(
		fmt.Sprintf("%s T%.0f H%.0f P%.0f", truncateRunes(wd.SensorID, 3), wd.Temperature, wd.Humidity, wd.Pressure*0.75), 0, 0); err != nil {
		slog.Error("Writing sensor data to LCD", "error", err)
	}
}

func (l *LCD) printTwoRows(wd types.WeatherData) {
	if err := l.printer.Print(fmt.Sprintf("T:%.1fC H:%.1f%%", wd.Temperature, wd.Humidity), 0, 0); err != nil {
		slog.Error("Writing Temperature and Humidity to LCD", "error", err)
	}

	if err := l.printer.Print(fmt.Sprintf("P:%.1fmm [%s]", wd.Pressure*0.75, truncateRunes(wd.SensorID, 3)), 1, 0); err != nil {
		slog.Error("Writing Pressure to LCD", "error", err)
	}
}

func (l *LCD) printFourRows(wd types.WeatherData) {
	idLenLimit := 12
	if l.printer.Columns() == 20 {
		idLenLimit = 16
	}

	if err := l.printer.Print(fmt.Sprintf("ID: %s", truncateRunes(wd.SensorID, idLenLimit)), 0, 0); err != nil {
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

// truncateRunes returns s shortened to at most n runes. Byte-wise
// truncation (`fmt.Sprintf("%.*s", n, s)`) may split a multi-byte rune and
// produce invalid UTF-8 for non-ASCII sensor IDs.
func truncateRunes(s string, n int) string {
	if n <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n])
}
