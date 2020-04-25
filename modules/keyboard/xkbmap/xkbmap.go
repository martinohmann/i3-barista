package xkbmap

import (
	"github.com/martinohmann/i3-barista/modules/keyboard"
	"github.com/martinohmann/i3-barista/xkbmap"
)

func New(layouts ...string) *keyboard.Module {
	return keyboard.New(&provider{}, layouts...)
}

type provider struct{}

func (p *provider) SetLayout(layout string) error {
	return xkbmap.SetLayout(layout)
}

func (p *provider) GetLayout() (string, error) {
	info, err := xkbmap.Query()
	if err != nil {
		return "", err
	}

	return info.Layout, nil
}
