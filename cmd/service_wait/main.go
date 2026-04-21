package main

import (
	"context"
	servicewait "github.com/mafigit/service-wait/pkg/service-wait"
	"github.com/urfave/cli/v3"
	"os"
)

func main() {
	cmd := &cli.Command{
		Name:        "service-wait",
		Description: "Wait for http, mongo or psql endpoints to be available",
		Flags:       servicewait.GetFlags(),
		Version:     "0.1.0",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			config, err := servicewait.New(cmd)
			if err != nil {
				return err
			}

			if servicewait.IsHTTPProbeActive(config) {
				err := servicewait.ProbeHttpEndpoint(ctx, config)
				if err != nil {
					return err
				}
			}
			if servicewait.IsMongoProbeActive(config) {
				err := servicewait.ProbeMongoEndpoint(ctx, config)
				if err != nil {
					return err
				}
			}
			if servicewait.IsPostgresProbeActive(config) {
				err := servicewait.ProbePSQLEndpoint(ctx, config)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		servicewait.Log.Error(err.Error())
		os.Exit(1)
	}
}
