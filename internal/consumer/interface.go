package consumer

import (
	"context"

	"github.com/realfatcat/raspiweather/pkg/types"
)

type Consumer interface {
	Consume(ctx context.Context, ch <-chan types.WeatherData)
}
