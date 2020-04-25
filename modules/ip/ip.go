package ip

import (
	"net"
	"time"

	"barista.run/bar"
	"barista.run/base/notifier"
	"barista.run/base/value"
	"barista.run/outputs"
	"barista.run/timing"
)

type Provider interface {
	GetIP() (net.IP, error)
}

type ProviderFunc func() (net.IP, error)

func (f ProviderFunc) GetIP() (net.IP, error) {
	return f()
}

type Info struct {
	net.IP
}

func (i Info) Connected() bool {
	return i.IP != nil
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
		return outputs.Textf("%s", info.IP)
	})

	m.Every(10 * time.Minute)

	return m
}

func defaultClickHandler(m *Module) func(bar.Event) {
	return func(e bar.Event) {
		if e.Button == bar.ButtonLeft {
			m.Refresh()
		}
	}
}

func (m *Module) Stream(s bar.Sink) {
	ip, err := m.provider.GetIP()
	outputFunc := m.outputFunc.Get().(func(Info) bar.Output)
	for {
		if s.Error(err) {
			continue
		}

		info := Info{ip}

		s.Output(outputs.Group(outputFunc(info)).OnClick(defaultClickHandler(m)))

		select {
		case <-m.outputFunc.Next():
			outputFunc = m.outputFunc.Get().(func(Info) bar.Output)
		case <-m.notifyCh:
			ip, err = m.provider.GetIP()
		case <-m.scheduler.C:
			ip, err = m.provider.GetIP()
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
