package testmetrics

import (
	"sync"
	"time"
)

type Statter struct {
	lock                *sync.RWMutex
	incCalls            []IncCall
	gaugeCalls          []GaugeCall
	timingDurationCalls []TimingDurationCall
}

func NewStatter() *Statter {
	return &Statter{
		lock: &sync.RWMutex{},
	}
}

func (s *Statter) IncCalls() []IncCall {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.incCalls
}

func (s *Statter) GaugeCalls() []GaugeCall {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.gaugeCalls
}

func (s *Statter) TimingDurationCalls() []TimingDurationCall {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.timingDurationCalls
}

func (s *Statter) Inc(metric string, value int64, rate float32) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.incCalls = append(s.incCalls, IncCall{
		Metric: metric,
		Value:  value,
		Rate:   rate,
	})

	return nil
}

func (s *Statter) Gauge(metric string, value int64, rate float32) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.gaugeCalls = append(s.gaugeCalls, GaugeCall{
		Metric: metric,
		Value:  value,
		Rate:   rate,
	})

	return nil
}

func (s *Statter) TimingDuration(metric string, value time.Duration, rate float32) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.timingDurationCalls = append(s.timingDurationCalls, TimingDurationCall{
		Metric: metric,
		Value:  value,
		Rate:   rate,
	})

	return nil
}

type IncCall struct {
	Metric string
	Value  int64
	Rate   float32
}

type GaugeCall struct {
	Metric string
	Value  int64
	Rate   float32
}

type TimingDurationCall struct {
	Metric string
	Value  time.Duration
	Rate   float32
}
