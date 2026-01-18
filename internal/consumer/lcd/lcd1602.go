package lcd

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/realfatcat/raspiweather/pkg/types"
)

type LCD struct {
	p Printer
}

func New(p Printer) *LCD {
	return &LCD{
		p: p,
	}
}

func (l *LCD) Consume(ctx context.Context, ch <-chan types.WeatherData) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case sd, ok := <-ch:
				if !ok {
					// Channel closed, exit gracefully
					return
				}
				if err := l.p.Print(fmt.Sprintf("T:%.1fC H:%.1f%%", sd.Temperature, sd.Humidity), 0, 0); err != nil {
					slog.Error("Writing Temperature and Humidity to LCD", "error", err)
				}

				if err := l.p.Print(fmt.Sprintf("P:%.1fmmHg  ^_^", sd.Pressure*0.75), 1, 0); err != nil {
					slog.Error("Writing Pressure to LCD", "error", err)
				}
			}
		}
	}()
}
