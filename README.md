# i3-barista

[![Build Status](https://travis-ci.org/martinohmann/i3-barista.svg?branch=master)](https://travis-ci.org/martinohmann/i3-barista)
[![GoDoc](https://godoc.org/github.com/martinohmann/i3-barista?status.svg)](https://godoc.org/github.com/martinohmann/i3-barista)
![GitHub](https://img.shields.io/github/license/martinohmann/i3-barista?color=orange)

## Installation

```
go get -u github.com/martinohmann/i3-barista
```

## Bar

### Dependencies

The bar requires the following binaries to be available in the path to function correctly:

- `pacman` and `checkupdates` for displaying pacman updates in the bar
- `nmtui-connect` for managing wifi networks
- `urxvt` to open certain click actions in a terminal
- `dmenu_session` for displaying the session picker
- `xset` for querying and toggling DPMS
- `setxkblayout` for querying and toggling keyboard layouts
- `notify send` for displaying the calendar when clicking the clock

## License

The source code of i3-barista is released under the MIT License. See the bundled
LICENSE file for details.


