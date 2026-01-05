package consumer

import (
	"context"

	"github.com/realfatcat/raspiweather/internal/types"
)

type Consumer interface {
	Consume(ctx context.Context, ch <-chan types.WeatherData)
}
