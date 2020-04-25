package updates

import (
	"testing"

	"barista.run/bar"
	"barista.run/outputs"
	testBar "barista.run/testing/bar"
)

func TestModule(t *testing.T) {
	testBar.New(t)

	p := ProviderFunc(func() func() (int, error) {
		var i int
		return func() (int, error) {
			i++
			return i, nil
		}
	}())

	u := New(p)
	testBar.Run(u)

	testBar.LatestOutput().AssertText([]string{"1 update"})
	u.Refresh()
	testBar.LatestOutput().AssertText([]string{"2 updates"})
	u.Output(func(updates int) bar.Output {
		return outputs.Textf("%d", updates)
	})
	testBar.LatestOutput().AssertText([]string{"2"})
	u.Refresh()
	testBar.LatestOutput().AssertText([]string{"3"})
}
