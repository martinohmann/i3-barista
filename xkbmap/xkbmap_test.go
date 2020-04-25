package xkbmap

import (
	"reflect"
	"testing"
)

func TestQuery(t *testing.T) {
	oldQuery := query
	defer func() { query = oldQuery }()

	query = func() ([]byte, error) {
		return []byte(`rules:      evdev
model:      pc105
layout:     us
`), nil
	}

	expected := Info{
		Rules:  "evdev",
		Model:  "pc105",
		Layout: "us",
	}

	info, err := Query()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(expected, info) {
		t.Errorf("expected %#v, got %#v", info, expected)
	}
}
