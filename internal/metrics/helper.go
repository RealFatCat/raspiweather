package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

func gaugeVec(reg prometheus.Registerer, name, help string, labels []string) (*prometheus.GaugeVec, error) {
	vec := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: name,
			Help: help,
		},
		labels,
	)
	if err := reg.Register(vec); err != nil {
		return nil, fmt.Errorf("registering %s: %w", name, err)
	}
	return vec, nil
}
