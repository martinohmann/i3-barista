package xset

import (
	"github.com/martinohmann/i3-barista/internal/xset"
	"github.com/martinohmann/i3-barista/modules/dpms"
)

// New creates a new *dpms.Module using xset as a DPMS provider.
func New() *dpms.Module {
	return dpms.New(&provider{})
}

type provider struct{}

// Set implements dpms.Provider.
func (*provider) Set(enabled bool) error {
	return xset.SetDPMS(enabled)
}

// Get implements dpms.Provider.
func (*provider) Get() (bool, error) {
	return xset.GetDPMS()
}
