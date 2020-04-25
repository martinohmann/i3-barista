package pacman

import (
	"os/exec"
	"strings"

	"github.com/martinohmann/i3-barista/modules/updates"
)

func New() *updates.Module {
	return updates.New(Provider)
}

var Provider = updates.ProviderFunc(func() (int, error) {
	out, err := exec.Command("checkupdates").Output()
	if err != nil {
		return 0, err
	}

	lines := strings.Split(strings.TrimRight(string(out), "\n"), "\n")

	return len(lines), nil
})
