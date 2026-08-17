// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"

	envoy_service_auth_v3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"github.com/gardener/gardener/cmd/utils/initrun"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
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
	var listeners []net.Listener

	if o.unixSocket != "" {
		// Cleanup possible leftovers
		_ = os.Remove(o.unixSocket)

		unixListener, err := net.Listen("unix", o.unixSocket)
		if err != nil {
			return fmt.Errorf("failed to listen on %s: %w", o.unixSocket, err)
		}
		defer os.Remove(o.unixSocket)
		listeners = append(listeners, unixListener)
	}

	port := o.port
	tcpListener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen on %d: %w", port, err)
	}
	listeners = append(listeners, tcpListener)

	// TLS is only supported for TCP-only mode
	var serverOpts []grpc.ServerOption
	if o.unixSocket != "" {
		log.Info("TLS is not supported with unix domain sockets, running in plaintext mode...")
	} else if o.tlsCert != "" && o.tlsKey != "" {
		creds, err := credentials.NewServerTLSFromFile(o.tlsCert, o.tlsKey)
		if err != nil {
			return fmt.Errorf("failed to load TLS credentials: %w", err)
		}
		serverOpts = append(serverOpts, grpc.Creds(creds))
	} else {
		log.Info("TLS certificates not provided, running in plaintext mode...")
	}

	gs := grpc.NewServer(serverOpts...)

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(gs, healthServer)
	healthServer.SetServingStatus("envoy.service.auth.v3.Authorization", grpc_health_v1.HealthCheckResponse_SERVING)

	if o.reflection {
		reflection.Register(gs)
	}
	authsrv, err := extauthzserver.New(log, os.DirFS(o.secretsDir))
	if err != nil {
		return fmt.Errorf("failed to set up authorization server: %w", err)
	}
	envoy_service_auth_v3.RegisterAuthorizationServer(gs, authsrv)

	if o.unixSocket != "" {
		log.Info("Starting gRPC server", "port", port, "unix socket path", o.unixSocket, "reflection", o.reflection)
	} else {
		log.Info("Starting gRPC server", "port", port, "reflection", o.reflection)
	}

	eg, egCtx := errgroup.WithContext(ctx)
	for _, listener := range listeners {
		eg.Go(func() error {
			if err := gs.Serve(listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
				return err
			}
			return nil
		})
	}

	eg.Go(func() error {
		<-egCtx.Done()
		if errors.Is(ctx.Err(), context.Canceled) {
			log.Info("Graceful shutdown")
		}
		healthServer.Shutdown()
		gs.GracefulStop()
		return nil
	})

	return eg.Wait()
}
