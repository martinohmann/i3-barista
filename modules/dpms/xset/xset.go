package xset

import (
	"github.com/martinohmann/i3-barista/modules/dpms"
	"github.com/martinohmann/i3-barista/xset"
)

func New() *dpms.Module {
	return dpms.New(&provider{})
}

type provider struct{}

func (*provider) Set(enabled bool) error {
	return xset.SetDPMS(enabled)
}

func (*provider) Get() (bool, error) {
	return xset.GetDPMS()
}
