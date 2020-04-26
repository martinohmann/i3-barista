package xkbmap

import (
	"os/exec"
	"regexp"
	"strings"
)

// query can be replaced in tests.
var query = func() ([]byte, error) {
	return exec.Command("setxkbmap", "-query").Output()
}

type Info struct {
	Rules  string
	Model  string
	Layout string
}

var xkbInfoRegexp = regexp.MustCompile(`([^:]*?)\s*:\s*(.*)$`)

func Query() (Info, error) {
	raw, err := query()
	if err != nil {
		return Info{}, err
	}

	lines := strings.Split(string(raw), "\n")

	info := Info{}

	for _, line := range lines {
		submatches := xkbInfoRegexp.FindStringSubmatch(line)
		if submatches == nil {
			continue
		}

		key := submatches[1]
		value := submatches[2]

		switch key {
		case "rules":
			info.Rules = value
		case "model":
			info.Model = value
		case "layout":
			info.Layout = value
		}
	}

	return info, nil
}

func SetLayout(layout string) error {
	return exec.Command("setxkbmap", layout).Run()
}
