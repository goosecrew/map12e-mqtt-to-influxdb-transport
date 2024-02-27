package app

import "time"

type Cache struct {
	Metrics []*Metric
}

func NewCache() *Cache {
	return &Cache{
		Metrics: make([]*Metric, 0),
	}
}

func (c *Cache) GetOrCreateMetric(topic string) *Metric {
	for _, metric := range c.Metrics {
		if metric.Topic == topic {
			return metric
		}
	}
	m := NewMetric(topic)
	c.Metrics = append(c.Metrics, m)
	return m
}

type Metric struct {
	UpdatedAt time.Time
	Topic     string
	Value     float64
}

func NewMetric(topic string) *Metric {
	return &Metric{Topic: topic}
}

func (m *Metric) Set(v float64) bool {
	mtx.Lock()
	defer mtx.Unlock()
	if time.Now().Sub(m.UpdatedAt) > params.CacheTimeout {
		m.Value = v
		m.UpdatedAt = time.Now()
		return true
	} else {
		return false
	}
}

func (m *Metric) Get() float64 {
	mtx.RLock()
	defer mtx.RUnlock()
	return m.Value
}
