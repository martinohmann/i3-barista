package keyboard

import (
	"errors"
	"sync"
	"testing"

	"barista.run/bar"
	"barista.run/outputs"
	testBar "barista.run/testing/bar"
	"golang.org/x/time/rate"
)

type testProvider struct {
	sync.Mutex
	err    error
	layout string
}

func (p *testProvider) GetLayout() (string, error) {
	p.Lock()
	defer p.Unlock()
	if p.err != nil {
		return "", p.err
	}

	return p.layout, nil
}

func (p *testProvider) SetLayout(layout string) error {
	p.Lock()
	defer p.Unlock()
	if p.err != nil {
		return p.err
	}

	p.layout = layout
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
		layout: "us",
	}

	k := New(testProvider, "us", "de", "fr")
	testBar.Run(k)

	out := testBar.NextOutput("on start")
	out.AssertText([]string{"us"})
	testProvider.SetLayout("de")
	k.Refresh()
	out = testBar.NextOutput("layout changed")
	out.AssertText([]string{"de"})

	oldRateLimiter := RateLimiter
	defer func() { RateLimiter = oldRateLimiter }()
	// To speed up the tests.
	RateLimiter = rate.NewLimiter(rate.Inf, 0)

	out.At(0).Click(bar.Event{Button: bar.ScrollUp})
	out = testBar.NextOutput("next layout")
	out.AssertText([]string{"de"}, "change us -> de")

	out.At(0).Click(bar.Event{Button: bar.ButtonLeft})
	out = testBar.NextOutput("next layout")
	out.AssertText([]string{"fr"}, "change de -> fr")

	out.At(0).Click(bar.Event{Button: bar.ButtonLeft})
	out = testBar.NextOutput("layout wrap around")
	out.AssertText([]string{"us"}, "change fr -> us")

	out.At(0).Click(bar.Event{Button: bar.ButtonRight})
	out = testBar.NextOutput("layout wrap around - reverse")
	out.AssertText([]string{"fr"}, "change us -> fr")

	out.At(0).Click(bar.Event{Button: bar.ScrollDown})
	out = testBar.NextOutput("prev layout")
	out.AssertText([]string{"de"}, "change fr -> de")

	testProvider.setError(errors.New("foo"))

	out.At(0).Click(bar.Event{Button: bar.ScrollDown})
	testBar.AssertNoOutput("error during volume change")

	out.At(0).Click(bar.Event{Button: bar.ButtonLeft})
	testBar.AssertNoOutput("error during mute")

	testProvider.setError(nil)

	k.Output(func(layout Layout) bar.Output {
		return outputs.Textf("keyboard: %s", layout.Name).
			OnClick(func(e bar.Event) {
				switch e.Button {
				case bar.ButtonLeft:
					layout.SetLayout("de")
				case bar.ButtonRight:
					layout.SetLayout("fr")
				case bar.ScrollDown:
					layouts := layout.AllLayouts()
					if len(layouts) > 0 {
						layout.SetLayout(layouts[0])
					}
				}
			})
	})

	out = testBar.NextOutput("on output format change")

	out.At(0).Click(bar.Event{Button: bar.ButtonRight})
	out = testBar.NextOutput("switch to fr")
	out.AssertText([]string{"keyboard: fr"}, "change to fr")

	out.At(0).Click(bar.Event{Button: bar.ButtonLeft})
	out = testBar.NextOutput("switch to de")
	out.AssertText([]string{"keyboard: de"}, "change to de")

	out.At(0).Click(bar.Event{Button: bar.ScrollDown})
	out = testBar.NextOutput("switch to us")
	out.AssertText([]string{"keyboard: us"}, "change to us")
}