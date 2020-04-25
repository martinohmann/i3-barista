package dpms

import (
	"time"

	"barista.run/bar"
	"barista.run/base/notifier"
	"barista.run/base/value"
	l "barista.run/logging"
	"barista.run/outputs"
	"barista.run/timing"
)

type Provider interface {
	Get() (bool, error)
	Set(enabled bool) error
}

type Info struct {
	Enabled bool

	provider Provider
	update   func()
}

func (i Info) String() string {
	if i.Enabled {
		return "dpms enabled"
	}

	return "dpms disabled"
}

func (i Info) Enable() {
	i.setEnabled(true)
}

func (i Info) Disable() {
	i.setEnabled(false)
}

func (i Info) Toggle() {
	enabled, err := i.provider.Get()
	if err != nil {
		l.Log("Error obtaining DPMS status: %v", err)
		return
	}

	i.setEnabled(!enabled)
}

func (i Info) setEnabled(enabled bool) {
	if err := i.provider.Set(enabled); err != nil {
		l.Log("Error updating DPMS status: %v", err)
		return
	}

	i.update()
}

type Module struct {
	provider   Provider
	outputFunc value.Value // of func(Info) bar.Output
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
	m.outputFunc.Set(func(info Info) bar.Output {
		return outputs.Textf("%s", info)
	})

	m.Every(1 * time.Minute)

	return m
}

func defaultClickHandler(i Info) func(bar.Event) {
	return func(e bar.Event) {
		if e.Button == bar.ButtonLeft {
			i.Toggle()
		}
	}
}

func (m *Module) Stream(s bar.Sink) {
	enabled, err := m.provider.Get()
	outputFunc := m.outputFunc.Get().(func(Info) bar.Output)
	for {
		if s.Error(err) {
			continue
		}

		info := Info{
			Enabled:  enabled,
			update:   func() { m.Refresh() },
			provider: m.provider,
		}

		s.Output(outputs.Group(outputFunc(info)).OnClick(defaultClickHandler(info)))

		select {
		case <-m.outputFunc.Next():
			outputFunc = m.outputFunc.Get().(func(Info) bar.Output)
		case <-m.notifyCh:
			enabled, err = m.provider.Get()
		case <-m.scheduler.C:
			enabled, err = m.provider.Get()
		}
	}
}

func (m *Module) Output(format func(Info) bar.Output) *Module {
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
