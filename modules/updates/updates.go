package updates

import (
	"time"

	"barista.run/bar"
	"barista.run/base/notifier"
	"barista.run/base/value"
	"barista.run/outputs"
	"barista.run/timing"
)

// Provider provides the count of currently available updates for the bar.
type Provider interface {
	Updates() (int, error)
}

// ProviderFunc is a func that satisfies the Provider interface.
type ProviderFunc func() (int, error)

// Updates implements Provider.
func (f ProviderFunc) Updates() (int, error) {
	return f()
}

// Module is a module for displaying currently available updates in the bar.
type Module struct {
	outputFunc value.Value // of func(int) bar.Output
	provider   Provider
	notifyCh   <-chan struct{}
	notifyFn   func()
	scheduler  *timing.Scheduler
}

// New creates a new *Module with the given update count provider. By default,
// the module will refresh the update counts every hour. The refresh interval
// can be configured using `Every`.
func New(provider Provider) *Module {
	m := &Module{
		provider:  provider,
		scheduler: timing.NewScheduler(),
	}

	m.notifyFn, m.notifyCh = notifier.New()
	m.outputFunc.Set(func(updates int) bar.Output {
		if updates == 1 {
			return outputs.Text("1 update")
		}
		return outputs.Textf("%d updates", updates)
	})

	m.Every(time.Hour)

	return m
}

// Stream implements bar.Module.
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

// Output updates the output format func.
func (m *Module) Output(format func(int) bar.Output) *Module {
	m.outputFunc.Set(format)
	return m
}

// Every configures the refresh interval for the module. Passing a zero
// interval will disable refreshing.
func (m *Module) Every(interval time.Duration) *Module {
	if interval == 0 {
		m.scheduler.Stop()
	} else {
		m.scheduler.Every(interval)
	}
	return m
}

// Refresh forces a refresh of the module output.
func (m *Module) Refresh() {
	m.notifyFn()
}
