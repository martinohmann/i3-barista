package keyboard

import (
	"sync"
	"time"

	"barista.run/bar"
	"barista.run/base/notifier"
	"barista.run/base/value"
	"barista.run/outputs"
	"barista.run/timing"
	"golang.org/x/time/rate"
)

type Provider interface {
	GetLayout() (string, error)
	SetLayout(layout string) error
}

type Controller interface {
	Next()
	Previous()
	Current() string
	SetLayout(layout string)
	AllLayouts() []string
}

type Layout struct {
	Controller

	Name string
}

func (l Layout) String() string {
	return l.Name
}

type controller struct {
	sync.Mutex
	layoutMap map[string]int
	layouts   []string
	current   int
	provider  Provider
	update    func(string)
}

func NewController(provider Provider, layouts []string, updateFn func(string)) Controller {
	c := &controller{
		layouts:   layouts,
		layoutMap: make(map[string]int),
		provider:  provider,
		update:    updateFn,
	}

	for i, layout := range layouts {
		c.layoutMap[layout] = i
	}

	currentLayout, _ := provider.GetLayout()

	// Set the current layout as active, add it to the list of layouts if not
	// present yet.
	i, ok := c.layoutMap[currentLayout]
	if ok {
		c.current = i
	} else {
		c.current = len(c.layouts)
		c.layoutMap[currentLayout] = c.current
		c.layouts = append(c.layouts, currentLayout)
	}

	return c
}

func (c *controller) AllLayouts() []string {
	c.Lock()
	defer c.Unlock()
	return c.layouts
}

func (c *controller) Current() string {
	c.Lock()
	defer c.Unlock()
	return c.layouts[c.current]
}

func (c *controller) Next() {
	c.Lock()
	defer c.Unlock()

	c.current++
	if c.current >= len(c.layouts) {
		c.current = 0
	}

	c.setLayout()
}

func (c *controller) Previous() {
	c.Lock()
	defer c.Unlock()

	c.current--
	if c.current < 0 {
		c.current = len(c.layouts) - 1
	}

	c.setLayout()
}

func (c *controller) SetLayout(layout string) {
	c.Lock()
	defer c.Unlock()

	idx, ok := c.layoutMap[layout]
	if !ok {
		return
	}

	c.current = idx
	c.setLayout()
}

func (c *controller) setLayout() {
	current := c.layouts[c.current]
	c.provider.SetLayout(current)

	if c.update != nil {
		c.update(current)
	}
}

type Module struct {
	controller Controller
	provider   Provider
	outputFunc value.Value // of func(Info) bar.Output
	notifyCh   <-chan struct{}
	notifyFn   func()
	scheduler  *timing.Scheduler
}

func New(provider Provider, layouts ...string) *Module {
	m := &Module{
		provider:  provider,
		scheduler: timing.NewScheduler(),
	}

	m.controller = NewController(provider, layouts, func(string) {
		m.Refresh()
	})
	m.notifyFn, m.notifyCh = notifier.New()
	m.outputFunc.Set(func(layout Layout) bar.Output {
		return outputs.Text(layout.Name)
	})

	m.Every(10 * time.Second)

	return m
}

// RateLimiter throttles layout updates to once every ~20ms to avoid unexpected
// behaviour.
var RateLimiter = rate.NewLimiter(rate.Every(20*time.Millisecond), 1)

func defaultClickHandler(l Layout) func(bar.Event) {
	return func(e bar.Event) {
		if !RateLimiter.Allow() {
			return
		}

		switch {
		case e.Button == bar.ButtonLeft || e.Button == bar.ScrollUp:
			l.Next()
		case e.Button == bar.ButtonRight || e.Button == bar.ScrollDown:
			l.Previous()
		}
	}
}

func (m *Module) Stream(s bar.Sink) {
	layout, err := m.provider.GetLayout()
	outputFunc := m.outputFunc.Get().(func(Layout) bar.Output)
	for {
		if s.Error(err) {
			continue
		}

		l := Layout{
			Controller: m.controller,
			Name:       layout,
		}

		s.Output(outputs.Group(outputFunc(l)).OnClick(defaultClickHandler(l)))

		select {
		case <-m.outputFunc.Next():
			outputFunc = m.outputFunc.Get().(func(Layout) bar.Output)
		case <-m.notifyCh:
			layout, err = m.provider.GetLayout()
		case <-m.scheduler.C:
			layout, err = m.provider.GetLayout()
		}
	}
}

func (m *Module) Output(format func(Layout) bar.Output) *Module {
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
