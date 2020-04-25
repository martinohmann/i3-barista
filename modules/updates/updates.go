package updates

import (
	"time"

	"barista.run/bar"
	"barista.run/base/notifier"
	"barista.run/base/value"
	"barista.run/outputs"
	"barista.run/timing"
)

type Provider interface {
	Updates() (int, error)
}

type ProviderFunc func() (int, error)

func (f ProviderFunc) Updates() (int, error) {
	return f()
}

type Module struct {
	outputFunc value.Value // of func(int) bar.Output
	provider   Provider
	notifyCh   <-chan struct{}
	notifyFn   func()
	scheduler  *timing.Scheduler
}

func New(provider Provider) *Module {
	m := &Module{
		provider:  provider,
		scheduler: timing.NewScheduler(),
	}

	m.notifyFn, m.notifyCh = notifier.New()
	m.outputFunc.Set(func(updates int) bar.Output {
		return outputs.Textf("%d updates", updates)
	})

	m.Every(time.Hour)

	return m
}

func (m *Module) Stream(s bar.Sink) {
	updates, err := m.provider.Updates()
	outputFunc := m.outputFunc.Get().(func(int) bar.Output)
	for {
		if s.Error(err) {
			continue
		}
		s.Output(outputFunc(updates))
		select {
		case <-m.outputFunc.Next():
			outputFunc = m.outputFunc.Get().(func(int) bar.Output)
		case <-m.notifyCh:
			updates, err = m.provider.Updates()
		case <-m.scheduler.C:
			updates, err = m.provider.Updates()
		}
	}
}

func (m *Module) Output(format func(int) bar.Output) *Module {
	m.outputFunc.Set(format)
	return m
}

func (m *Module) Every(interval time.Duration) *Module {
	if interval == 0 {
		m.scheduler.Stop()
	} else {
		m.scheduler.Every(interval)
	}
	return m
}

func (m *Module) Refresh() {
	m.notifyFn()
}
