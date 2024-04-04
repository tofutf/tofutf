// Package fxlib contains providers, and invokers for uber/fx.
package fxlib

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/tofutf/tofutf/internal/logr"
)

// ProvideLogger provides a logger constructor.
func ProvideLogger() func(flagset *pflag.FlagSet) (logr.Logger, error) {
	return func(flagset *pflag.FlagSet) (logr.Logger, error) {
		loggerConfig := logr.NewConfigFromFlags(flagset)

		logger, err := logr.New(loggerConfig)
		if err != nil {
			return logr.NewNoopLogger(), err
		}

		return logger, nil
	}
}

func ProvideFlags(cmd *cobra.Command) func() *pflag.FlagSet {
	return func() *pflag.FlagSet {
		return cmd.Flags()
	}
}
