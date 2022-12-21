package mongo

import (
	"context"
	"sync"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/event"
)

type connectionCounters struct {
	total  int64
	inuse  int64
	failed int64
}

type monitorState struct {
	mu sync.Mutex

	appName  string
	ctx      context.Context
	counters map[string]*connectionCounters
}

func newMonitor(ctx context.Context, appName string) *monitorState {
	return &monitorState{
		appName:  appName,
		ctx:      ctx,
		counters: make(map[string]*connectionCounters),
	}
}

func (m *monitorState) PoolMonitor() *event.PoolMonitor {
	return &event.PoolMonitor{
		Event: m.Event,
	}
}

func (m *monitorState) Event(ev *event.PoolEvent) {
	if ev == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	l := log(m.ctx).WithField("mongodb.client.appname", m.appName).WithFields(formatEventFields(ev))

	if _, ok := m.counters[ev.Address]; !ok {
		m.counters[ev.Address] = new(connectionCounters)
	}

	switch ev.Type {
	case event.ConnectionCreated:
		m.counters[ev.Address].total++
	case event.ConnectionClosed:
		if m.counters[ev.Address].total > 0 {
			m.counters[ev.Address].total--
		}
	case event.GetSucceeded:
		m.counters[ev.Address].inuse++
	case event.GetFailed:
		m.counters[ev.Address].failed++
	case event.ConnectionReturned:
		if m.counters[ev.Address].inuse > 0 {
			m.counters[ev.Address].inuse--
		}
	case event.PoolCleared, event.PoolClosedEvent:
		m.counters[ev.Address] = new(connectionCounters)
	}

	l.WithFields(logrus.Fields{
		"mongodb.connections.total":  m.counters[ev.Address].total,
		"mongodb.connections.inuse":  m.counters[ev.Address].inuse,
		"mongodb.connections.failed": m.counters[ev.Address].failed,
	}).Info()
}

func formatEventFields(ev *event.PoolEvent) logrus.Fields {
	fields := logrus.Fields{
		"mongodb.pool.event.type":         ev.Type,
		"mongodb.pool.event.address":      ev.Address,
		"mongodb.pool.event.connectionId": ev.ConnectionID,
	}

	if ev.Reason != "" {
		fields["mongodb.pool.event.reason"] = ev.Reason
	}

	if ev.PoolOptions != nil {
		fields["mongodb.pool.event.options"] = *ev.PoolOptions
	}

	return fields
}
