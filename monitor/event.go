package monitor

import (
	"github.com/ljjjustin/themis/config"
	"errors"
)

type Event struct {
	Hostname   string
	NetworkTag string
	Status     string
}

type Events []*Event

type MonitorInterface interface {
	Start() (chan Events, error)
}

func NewEventMonitor(cfg *config.MonitorConfig) MonitorInterface {

	// TODO: create monitor according to monitor type
	return NewSerfMonitor(cfg.Address)
}

type EventCollector struct {
	Tag       string
	EventChan chan Events
	Monitor   MonitorInterface
}

func NewEventCollector(tag string, cfg *config.MonitorConfig) *EventCollector {
	monitor := NewEventMonitor(cfg)
	return &EventCollector{
		Tag:     tag,
		Monitor: monitor,
	}
}

func (c *EventCollector) Start() error {
	eventCh, err := c.Monitor.Start()
	if err != nil {
		return err
	}
	c.EventChan = eventCh
	return nil
}

func (c *EventCollector) DrainEvents() (Events, error) {
	select {
	case events := <-c.EventChan:
		if judgeSerfDown(events) {
			return nil, errors.New("leader serf down")
		}
		return events, nil
	default:
		return nil, nil
	}
}

func judgeSerfDown(events Events) bool {

	// group by network type
	netType := map[string][]string{}

	for _, e := range events {

		t := netType[e.NetworkTag]

		if t == nil {
			t := []string{}
			netType[e.NetworkTag] = t
		}

		netType[e.NetworkTag] = append(netType[e.NetworkTag], e.Status)

	}

	for _, statusList := range netType {

		failedTimes := 0

		for _, ststus := range statusList {

			if ststus == "failed" {
				failedTimes += 1
			}
		}

		if failedTimes >= len(statusList) - 1 {
			return true
		}
	}

	return false
}