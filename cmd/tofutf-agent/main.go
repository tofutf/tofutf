package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	cmdutil "github.com/tofutf/tofutf/cmd"
	"github.com/tofutf/tofutf/internal"
	"github.com/tofutf/tofutf/internal/agent"
	otfapi "github.com/tofutf/tofutf/internal/api"
	"github.com/tofutf/tofutf/internal/xslog"
)

func main() {
	// Configure ^C to terminate program
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-ctx.Done()
		// Stop handling ^C; another ^C will exit the program.
		cancel()
	}()

	if err := run(ctx, os.Args[1:]); err != nil {
		cmdutil.PrintError(err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	var (
		loggerConfig *xslog.Config
		clientConfig otfapi.Config
		agentConfig  *agent.Config
	)

	cmd := &cobra.Command{
		Use:           "otf-agent",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       internal.Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger, err := xslog.New(loggerConfig)
			if err != nil {
				return err
			}

			agent, err := agent.NewPoolDaemon(logger, *agentConfig, clientConfig)
			if err != nil {
				return fmt.Errorf("initializing agent: %w", err)
			}

			// blocks
			return agent.Start(cmd.Context())
		},
	}

	cmd.Flags().StringVar(&clientConfig.Address, "address", otfapi.DefaultAddress, "Address of OTF server")
	cmd.Flags().StringVar(&clientConfig.Token, "token", "", "Agent token for authentication")
	cmd.MarkFlagRequired("token") //nolint:errcheck
	cmd.SetArgs(args)

	loggerConfig = xslog.NewConfigFromFlags(cmd.Flags())
	agentConfig = agent.NewConfigFromFlags(cmd.Flags())

	if err := cmdutil.SetFlagsFromEnvVariables(cmd.Flags()); err != nil {
		return errors.Wrap(err, "failed to populate config from environment vars")
	}

	return cmd.ExecuteContext(ctx)
}
