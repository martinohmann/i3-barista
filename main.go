package main

import (
	"fmt"
	"log"

	barista "barista.run"
	"github.com/martinohmann/i3-barista/modules"
	"github.com/spf13/cobra"
)

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
	registerModules, ok := barFactoryFuncs[o.bar]
	if !ok {
		return fmt.Errorf("unsupported bar %q", o.bar)
	}

	registry := modules.NewRegistry()

	err := registerModules(registry)
	if err != nil {
		return err
	}

	return barista.Run(registry.Modules()...)
}
