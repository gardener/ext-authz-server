// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"
	"net"
	"os"

	envoy_service_auth_v3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"github.com/gardener/gardener/cmd/utils/initrun"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	extauthzserver "github.com/gardener/ext-authz-server/pkg/ext-authz-server"
)

// Name is a const for the name of this component.
const Name = "ext-authz-server"

// NewCommand creates a new cobra.Command for running gardenadm.
func NewCommand() *cobra.Command {
	opts := &options{}

	cmd := &cobra.Command{
		Use: Name,

		RunE: func(cmd *cobra.Command, _ []string) error {
			log, err := initrun.InitRun(cmd, opts, Name)
			if err != nil {
				return err
			}
			log.Info("Starting server...")
			return run(cmd.Context(), log, opts)
		},
	}

	opts.AddFlags(cmd.Flags())

	return cmd
}

func run(ctx context.Context, log logr.Logger, o *options) error {
	port := o.port
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen to %d: %v", port, err)
	}

	gs := grpc.NewServer()

	if o.reflection {
		reflection.Register(gs)
	}
	authsrv, err := extauthzserver.New(log, os.DirFS(o.secretsDir))
	if err != nil {
		return fmt.Errorf("failed to set up authorization server: %w", err)
	}
	envoy_service_auth_v3.RegisterAuthorizationServer(gs, authsrv)

	log.Info("Starting gRPC server", "port", port, "reflection", o.reflection)

	errorChannel := make(chan error)
	go func(errc chan error) {
		defer close(errc)
		errc <- gs.Serve(listener)
	}(errorChannel)

	select {
	case <-ctx.Done():
		log.Info("Graceful shutdown")
		gs.GracefulStop()
		return nil
	case err := <-errorChannel:
		return err
	}
}
