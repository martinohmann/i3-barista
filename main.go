package main

import (
	"fmt"
	"log"

	barista "barista.run"
	"barista.run/colors"
	"github.com/martinohmann/i3-barista/modules"
	"github.com/spf13/cobra"
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

func main() {
	rootCmd := newRootCommand()

	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

func newRootCommand() *cobra.Command {
	o := &options{bar: "top"}

	cmd := &cobra.Command{
		Use: "i3-barista",
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.Run()
		},
	}

	cmd.Flags().StringVar(&o.bar, "bar", o.bar, "bar to output the status for")

	return cmd
}

type options struct {
	bar string
}

func (o *options) Run() error {
	factory, ok := barFactoryFuncs[o.bar]
	if !ok {
		return fmt.Errorf("unsupported bar %q", o.bar)
	}

	registry := modules.NewRegistry()

	factory(registry)

	if err := registry.Err(); err != nil {
		return err
	}

	return barista.Run(registry.Modules()...)
}
