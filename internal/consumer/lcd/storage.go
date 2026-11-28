package lcd

import (
	"sync"

	"github.com/realfatcat/raspiweather/pkg/types"
)

type storage struct {
	mu    sync.Mutex
	data  map[string]types.WeatherData
	order []string
	head  int
}

func newStorage() *storage {
	return &storage{
		data:  make(map[string]types.WeatherData),
		order: make([]string, 0, 8),
	}
}

func (s *storage) put(sensorID string, wd types.WeatherData) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.data[sensorID]; !ok {
		s.order = append(s.order, sensorID)
	}
	s.data[sensorID] = wd
}

func (s *storage) isEmpty() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.order) == 0
}

func (s *storage) getNext() (types.WeatherData, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.order) == 0 {
		return types.WeatherData{}, false
	}

	id := s.order[s.head]
	s.head = (s.head + 1) % len(s.order)

	wd, ok := s.data[id]
	return wd, ok
}
