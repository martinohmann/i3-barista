package main

import (
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"barista.run/bar"
	"barista.run/base/click"
	"barista.run/colors"
	"barista.run/format"
	"barista.run/group/switching"
	"barista.run/modules/battery"
	"barista.run/modules/clock"
	"barista.run/modules/cputemp"
	"barista.run/modules/diskspace"
	"barista.run/modules/meminfo"
	"barista.run/modules/netinfo"
	"barista.run/modules/netspeed"
	"barista.run/modules/static"
	"barista.run/modules/sysinfo"
	"barista.run/modules/volume"
	"barista.run/modules/volume/pulseaudio"
	"barista.run/modules/weather"
	"barista.run/modules/wlan"
	"barista.run/outputs"
	"github.com/kirsle/configdir"
	"github.com/martinlindhe/unit"
	"github.com/martinohmann/i3-barista/internal/notify"
	"github.com/martinohmann/i3-barista/modules"
	"github.com/martinohmann/i3-barista/modules/cpufreq"
	"github.com/martinohmann/i3-barista/modules/cpufreq/sysfs"
	"github.com/martinohmann/i3-barista/modules/dpms"
	"github.com/martinohmann/i3-barista/modules/dpms/xset"
	"github.com/martinohmann/i3-barista/modules/ip"
	"github.com/martinohmann/i3-barista/modules/ip/ipify"
	"github.com/martinohmann/i3-barista/modules/keyboard"
	"github.com/martinohmann/i3-barista/modules/keyboard/xkbmap"
	"github.com/martinohmann/i3-barista/modules/updates"
	"github.com/martinohmann/i3-barista/modules/updates/pacman"
	"github.com/martinohmann/i3-barista/modules/weather/openweathermap"
	psysfs "github.com/prometheus/procfs/sysfs"
)

func init() {
	colors.LoadFromMap(map[string]string{
		"default":  "#cccccc",
		"warning":  "#ffd760",
		"critical": "#ff5454",
		"disabled": "#777777",
		"color0":   "#2e3440",
		"color1":   "#3b4252",
		"color2":   "#434c5e",
		"color3":   "#4c566a",
		"color4":   "#d8dee9",
		"color5":   "#e5e9f0",
		"color6":   "#eceff4",
		"color7":   "#8fbcbb",
		"color8":   "#88c0d0",
		"color9":   "#81a1c1",
		"color10":  "#5e81ac",
		"color11":  "#bf616a",
		"color12":  "#d08770",
		"color13":  "#ebcb8b",
		"color14":  "#a3be8c",
		"color15":  "#b48ead",
	})
}

// barFactoryFuncs contains factory functions that populate the module registry
// for every configured bar.
var barFactoryFuncs = map[string]func(registry *modules.Registry) error{
	"top": func(registry *modules.Registry) error {
		return registry.
			Add(
				battery.All().Output(func(i battery.Info) bar.Output {
					var sep string
					switch {
					case i.Status == battery.Full:
						return outputs.Textf(" %d%%", i.RemainingPct())
					case i.Status == battery.Disconnected:
						return outputs.Text(" not present").Color(colors.Scheme("disabled"))
					case i.Status == battery.Charging:
						sep = " "
					case i.Status == battery.Discharging:
						sep = " "
					}

					out := outputs.Textf(" %d%% %s%s", i.RemainingPct(), sep, format.Duration(i.RemainingTime()))

					switch {
					case i.RemainingPct() < 5:
						out = out.Color(colors.Scheme("critical"))
					case i.RemainingPct() < 10:
						out = out.Color(colors.Scheme("color11"))
					case i.RemainingPct() < 15:
						out = out.Color(colors.Scheme("color12"))
					case i.RemainingPct() < 20:
						out = out.Color(colors.Scheme("color13"))
					}

					return out
				}),
				volume.New(pulseaudio.DefaultSink()).Output(func(v volume.Volume) bar.Output {
					out := outputs.Textf("婢 %d%%", v.Pct())
					if v.Mute {
						out = out.Color(colors.Scheme("color11"))
					}

					return out
				}),
				cputemp.OfType("acpitz").Output(func(t unit.Temperature) bar.Output {
					out := outputs.Textf(" %.0f°C", t.Celsius())
					switch {
					case t.Celsius() > 85:
						out = out.Color(colors.Scheme("critical"))
					case t.Celsius() > 80:
						out = out.Color(colors.Scheme("color11"))
					case t.Celsius() > 75:
						out = out.Color(colors.Scheme("color12"))
					case t.Celsius() > 70:
						out = out.Color(colors.Scheme("color13"))
					}

					return out
				}),
				pacman.New().Output(func(info updates.Info) bar.Output {
					if info.Updates == 0 {
						return nil
					}

					return outputs.Textf(" %d", info.Updates).
						OnClick(click.Left(func() {
							notify.Send("Available Pacman Updates", info.PackageDetails.String())
						}))
				}),
				wlan.Any().Output(func(info wlan.Info) bar.Output {
					onClick := click.RunLeft("urxvt", "-name", "nmtui", "-geometry", "100x40", "-e", "nmtui-connect")

					switch {
					case !info.Enabled():
						return nil
					case info.Connecting():
						return outputs.Text(" ...").Color(colors.Scheme("disabled")).OnClick(onClick)
					case !info.Connected():
						return outputs.Text(" disconnected").Color(colors.Scheme("disabled")).OnClick(onClick)
					default:
						return outputs.Textf(" %s", info.SSID).OnClick(onClick)
					}
				}),
				xkbmap.New("us", "de").Output(func(layout keyboard.Layout) bar.Output {
					return outputs.Textf("⌨ %s", strings.ToUpper(layout.Name))
				}),
				static.New(outputs.Text("").OnClick(click.RunLeft("dmenu_session"))),
			).
			Addf(func() (bar.Module, error) {
				replacer := strings.NewReplacer(
					"\u001b[7m", `<span foreground="#000000" background="#ffffff"><b>`,
					"\u001b[27m", `</b></span>`,
				)

				calenderFn := func() string {
					out, _ := exec.Command("cal", "--months", "6", "--color=always").Output()
					return string(out)
				}

				mod := clock.Local().Output(time.Second, func(now time.Time) bar.Output {
					return outputs.Textf(" %s ", now.Format("Mon Jan 02 2006 15:04")).
						OnClick(click.Left(func() {
							notify.Send("Calendar", replacer.Replace(calenderFn()))
						}))
				})
				return mod, nil
			}).
			Err()
	},
	"bottom": func(registry *modules.Registry) error {
		return registry.
			Addf(func() (bar.Module, error) {
				// Prefix of the interface that should be active initially.
				activePrefix := "wlp"

				ifaces, err := net.Interfaces()
				if err != nil {
					return nil, err
				}

				mods := make([]bar.Module, len(ifaces))

				// Collect modules.
				for i, iface := range ifaces {
					mods[i] = netspeed.New(iface.Name)
				}

				group, ctrl := switching.Group(mods...)

				// Don't need no buttons, click handlers will be set on all bar segments.
				ctrl.ButtonFunc(func(c switching.Controller) (start, end bar.Output) {
					return nil, nil
				})

				clickHandler := func(e bar.Event) {
					switch e.Button {
					case bar.ButtonLeft:
						ctrl.Next()
					case bar.ButtonRight:
						ctrl.Previous()
					}
				}

				// Setup module output and click handlers.
				for i, iface := range ifaces {
					mod := mods[i].(*netspeed.Module)

					mod.Output(func(s netspeed.Speeds) bar.Output {
						out := outputs.Textf("異 %s %s   %s ", iface.Name, format.IByterate(s.Tx), format.IByterate(s.Rx)).
							OnClick(clickHandler)

						if s.Connected() {
							return out.Color(colors.Scheme("color4"))
						}

						return out.Color(colors.Scheme("disabled"))
					})

					if strings.HasPrefix(iface.Name, activePrefix) {
						ctrl.Show(i)
					}
				}

				return group, nil
			}).
			Add(
				ipify.New().Output(func(i ip.Info) bar.Output {
					if i.Connected() {
						return outputs.Textf("爵 %s", i).Color(colors.Scheme("color5"))
					}

					return outputs.Text("爵 offline").Color(colors.Scheme("disabled"))
				}),
				netinfo.Prefix("tun").Output(func(s netinfo.State) bar.Output {
					if len(s.Name) == 0 {
						return nil
					}

					if len(s.IPs) < 1 {
						return outputs.Textf(" %s", s.Name).Color(colors.Scheme("disabled"))
					}

					return outputs.Textf(" %s %v", s.Name, s.IPs[0]).
						Color(colors.Scheme("color6"))
				}),
				netinfo.Prefix("wlp").Output(func(s netinfo.State) bar.Output {
					if len(s.Name) == 0 {
						return nil
					}

					if len(s.IPs) < 1 {
						return outputs.Textf(" %s", s.Name).Color(colors.Scheme("disabled"))
					}
					return outputs.Textf(" %s %v", s.Name, s.IPs[0]).
						Color(colors.Scheme("color7"))
				}),
				netinfo.Prefix("enp").Output(func(s netinfo.State) bar.Output {
					if len(s.Name) == 0 {
						return nil
					}

					if len(s.IPs) < 1 {
						return outputs.Textf(" %s", s.Name).Color(colors.Scheme("disabled"))
					}
					return outputs.Textf(" %s %v", s.Name, s.IPs[0]).
						Color(colors.Scheme("color8"))
				}),
				sysinfo.New().Output(func(i sysinfo.Info) bar.Output {
					return outputs.Textf("祥 up %v", format.Duration(i.Uptime)).
						Color(colors.Scheme("color9"))
				}),
			).
			Addf(func() (bar.Module, error) {
				fs, err := psysfs.NewDefaultFS()
				if err != nil {
					return nil, err
				}

				return sysfs.New(fs).Output(func(info cpufreq.Info) bar.Output {
					return outputs.Textf(" %.2fGHz", info.AverageFreq().Gigahertz()).
						Color(colors.Scheme("color10"))
				}), nil
			}).
			Add(
				sysinfo.New().Output(func(i sysinfo.Info) bar.Output {
					return outputs.Textf("溜 %.2f %.2f %.2f (%d)", i.Loads[0], i.Loads[1], i.Loads[2], i.Procs).
						Color(colors.Scheme("color11"))
				}),
				meminfo.New().Output(func(i meminfo.Info) bar.Output {
					used := (i["MemTotal"] - i.Available()).Gigabytes()
					total := i["MemTotal"].Gigabytes()

					return outputs.Textf(" %.1f/%.1fG", used, total).
						Color(colors.Scheme("color12"))
				}),
				diskspace.New("/").Output(func(i diskspace.Info) bar.Output {
					return outputs.Textf(" / %.2f/%.2fG", i.Used().Gigabytes(), i.Total.Gigabytes()).
						Color(colors.Scheme("color13")).
						OnClick(click.RunLeft("thunar", "/"))
				}),
			).
			Addf(func() (bar.Module, error) {
				configFile := configdir.LocalConfig("i3/barista/openweathermap.json")

				owm, err := openweathermap.NewFromConfig(configFile)
				switch {
				case os.IsNotExist(err):
					return nil, nil
				case err == openweathermap.ErrAPIKeyMissing:
					return static.New(outputs.Text(" apiKey missing").
						Color(colors.Scheme("disabled"))), nil
				case err != nil:
					return static.New(outputs.Errorf("failed to load openweathermap config: %v", err)), nil
				}

				return weather.New(owm).Output(func(info weather.Weather) bar.Output {
					return outputs.Textf(" %.0f°C, %s", info.Temperature.Celsius(), info.Description).
						Color(colors.Scheme("color14"))
				}), nil
			}).
			Add(
				xset.New().Output(func(info dpms.Info) bar.Output {
					out := outputs.Text("⏾ dpms ")

					if info.Enabled {
						return out.Color(colors.Scheme("color15"))
					}

					return out.Color(colors.Scheme("disabled"))
				}),
			).
			Err()
	},
}
