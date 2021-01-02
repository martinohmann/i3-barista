package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/user"

	barista "barista.run"
	"barista.run/oauth"
	"github.com/martinohmann/barista-contrib/modules"
	"github.com/martinohmann/i3-barista/internal/keyring"
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

	err = setupOAuthEncryption()
	if err != nil {
		return err
	}

	return barista.Run(registry.Modules()...)
}

func setupOAuthEncryption() error {
	var username string
	if u, err := user.Current(); err == nil {
		username = u.Username
	} else {
		username = fmt.Sprintf("user-%d", os.Getuid())
	}

	var secretBytes []byte
	// IMPORTANT: The oauth tokens used by some modules are very sensitive, so
	// we encrypt them with a random key and store that random key using
	// libsecret (gnome-keyring or equivalent). If no secret provider is
	// available, there is no way to store tokens (since the version of
	// sample-bar used for setup-oauth will have a different key from the one
	// running in i3bar). See also https://github.com/zalando/go-keyring#linux.
	secret, err := keyring.Get(username)
	if err == nil {
		secretBytes, err = base64.RawURLEncoding.DecodeString(secret)
	}

	if err != nil {
		secretBytes = make([]byte, 64)
		_, err := rand.Read(secretBytes)
		if err != nil {
			return err
		}

		secret = base64.RawURLEncoding.EncodeToString(secretBytes)
		err = keyring.Set(username, secret)
		if err != nil {
			return err
		}
	}

	oauth.SetEncryptionKey(secretBytes)
	return nil
}
