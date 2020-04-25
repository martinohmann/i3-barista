package dpms

import (
	"errors"
	"sync"
	"testing"

	"barista.run/bar"
	testBar "barista.run/testing/bar"
)

type testProvider struct {
	sync.Mutex
	err     error
	enabled bool
}

func (p *testProvider) Get() (bool, error) {
	p.Lock()
	defer p.Unlock()
	if p.err != nil {
		return false, p.err
	}

	return p.enabled, nil
}

func (p *testProvider) Set(enabled bool) error {
	p.Lock()
	defer p.Unlock()
	if p.err != nil {
		return p.err
	}

	p.enabled = enabled
	return nil
}

func (p *testProvider) setError(err error) {
	p.Lock()
	defer p.Unlock()
	p.err = err
}

func TestModule(t *testing.T) {
	testBar.New(t)

	testProvider := &testProvider{
		enabled: true,
	}

	k := New(testProvider)
	testBar.Run(k)

	out := testBar.NextOutput("on start")
	out.AssertText([]string{"dpms enabled"})
	testProvider.Set(false)
	k.Refresh()
	out = testBar.NextOutput("dpms disabled")
	out.AssertText([]string{"dpms disabled"})
	testProvider.Set(true)
	k.Refresh()
	out = testBar.NextOutput("reenabled")
	out.AssertText([]string{"dpms enabled"})

	out.At(0).Click(bar.Event{Button: bar.ButtonLeft})
	out = testBar.NextOutput("disabled via default click handler")
	out.AssertText([]string{"dpms disabled"})

	testProvider.setError(errors.New("whoops"))

	k.Refresh()
	out = testBar.NextOutput("error")
	out.AssertError()

	// @FIXME: for some reason the following shows weird behaviour
	//
	// testProvider.setError(nil)

	// k.Output(func(info Info) bar.Output {
	// 	return outputs.Textf("dpms: %v", info.Enabled).
	// 		OnClick(func(e bar.Event) {
	// 			switch e.Button {
	// 			case bar.ButtonLeft:
	// 				info.Enable()
	// 			case bar.ButtonRight:
	// 				info.Disable()
	// 			case bar.ScrollUp:
	// 				info.Disable()
	// 			}
	// 		})
	// })

	// out = testBar.NextOutput("on output format change")

	// out.At(0).Click(bar.Event{Button: bar.ButtonLeft})
	// out = testBar.NextOutput("enable")
	// out.AssertText([]string{"dpms: true"})

	// out.At(0).Click(bar.Event{Button: bar.ButtonRight})
	// out = testBar.NextOutput("disable")
	// out.AssertText([]string{"dpms: false"})

	// out.At(0).Click(bar.Event{Button: bar.ScrollUp})
	// out = testBar.NextOutput("toggle")
	// out.AssertText([]string{"dpms: true"})
}
