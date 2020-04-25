package pacman

import (
	"os/exec"
	"strings"

	"github.com/martinohmann/i3-barista/modules/updates"
)

// New creates a new *updates.Module with the pacman provider.
func New() *updates.Module {
	return updates.New(Provider)
}

// Provider is an updates.Provider which checks for pacman updates.
var Provider = updates.ProviderFunc(func() (int, error) {
	out, err := exec.Command("checkupdates").Output()
	if err != nil {
		return 0, err
	}

	lines := strings.Split(strings.TrimRight(string(out), "\n"), "\n")

	return len(lines), nil
})
