# i3-barista

[![Build Status](https://travis-ci.com/martinohmann/i3-barista.svg?branch=master)](https://travis-ci.com/martinohmann/i3-barista)
[![GoDoc](https://godoc.org/github.com/martinohmann/i3-barista?status.svg)](https://godoc.org/github.com/martinohmann/i3-barista)
![GitHub](https://img.shields.io/github/license/martinohmann/i3-barista?color=orange)

Additional modules for i3 [barista](https://github.com/soumya92/barista). This
repository also contains the configuration for the bars I use together with i3
status.

## Module installation

If you just want you use the [modules](https://godoc.org/github.com/martinohmann/i3-barista/modules):

```
go get -u github.com/martinohmann/i3-barista
```

## Bar

Screenshot

[![i3-barista](assets/screenshot.png)](assets/screenshot.png)

### Installation

To install the bar, run the following:

```
make install
```

This will build and place the `i3-barista` executable in `$GOPATH/bin`.

Update the `status_command` in the i3 configuration:

```conf
# top bar
bar {
  id bar0
  # You need one of the nerd fonts to correctly display the glyphs used in the
  # bar. See https://github.com/ryanoasis/nerd-fonts for more information.
  font "xft:Hack Nerd Font Mono Bold 9"
  status_command i3-barista --bar top
  position top
  ...
}

# bottom bar
bar {
  id bar1
  font "xft:Hack Nerd Font Mono Bold 9"
  status_command i3-barista --bar bottom
  position bottom
  ...
}
```

### Dependencies

The bar requires the following binaries to be available in the path to function correctly:

- `checkupdates` from
  [pacman-contrib](https://www.archlinux.org/packages/community/x86_64/pacman-contrib/)
  for displaying pacman updates in the bar
- `nmtui-connect` for managing wifi networks
- `urxvt` to open certain click actions in a terminal
- [`dmenu_session`](https://github.com/martinohmann/bin-pub/blob/master/dmenu_session)
  for displaying the session picker when clicking on the session bar segment
- `xset` for querying and toggling DPMS
- `setxkblayout` for querying and toggling keyboard layouts
- `notify-send` for displaying additional information when clicking certain bar segments

### OpenWeatherMap configuration

The OpenWeatherMap configuration is read from
`~/.config/i3/barista/openweathermap.json` and has to include at least the
`apiKey` field. See
[`modules/weather/openweathermap/owm.go`](modules/weather/openweathermap/owm.go)
for all configuration values.

## License

The source code of i3-barista is released under the MIT License. See the bundled
LICENSE file for details.
