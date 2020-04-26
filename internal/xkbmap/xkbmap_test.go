package xkbmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)
	assert.Equal(t, expected, info)
}
