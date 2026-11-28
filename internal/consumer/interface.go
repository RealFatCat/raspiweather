package consumer

import (
	"context"

	"github.com/realfatcat/raspiweather/pkg/types"
)

// Consumer processes weather data produced during the fanout run.
// Consume returns a channel that is closed when the consumer goroutine
// exits, letting the caller wait for shutdown before releasing resources.
type Consumer interface {
	Consume(ctx context.Context, ch <-chan types.WeatherData) <-chan struct{}
}
