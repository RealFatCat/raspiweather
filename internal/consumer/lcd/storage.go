package lcd

import (
	"sync"

	"github.com/realfatcat/raspiweather/pkg/types"
)

type storage struct {
	mu     sync.RWMutex
	data   map[string]types.WeatherData
	idxes  []string
	curIdx int
}

func newStorage() *storage {
	return &storage{
		data:  make(map[string]types.WeatherData),
		idxes: make([]string, 0, 8),
	}
}

func (s *storage) put(sensorID string, wd types.WeatherData) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[sensorID] = wd
	s.idxes = append(s.idxes, sensorID)
}

func (s *storage) getNext() (types.WeatherData, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.idxes) == 0 {
		return types.WeatherData{}, false
	}

	sd := s.idxes[s.curIdx]
	wd, ok := s.data[sd]

	s.curIdx++
	if s.curIdx >= len(s.idxes) {
		s.curIdx = 0
	}
	return wd, ok
}
