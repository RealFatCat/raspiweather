package meter

import "context"

// Metrics is an interface that defines the methods for recording metrics.
// In future we can use otel metrics library to record metrics,
// so it is better to keep context in the interface.
type Metrics interface {
	RecordTemperature(ctx context.Context, sensorID string, num float64)
	RecordHumidity(ctx context.Context, sensorID string, num float64)
	RecordPressure(ctx context.Context, sensorID string, num float64)
}
