// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"errors"

	"github.com/spf13/pflag"
)

type options struct {
	logLevel   string
	logFormat  string
	port       int
	secretsDir string
	reflection bool
	tlsCert    string
	tlsKey     string
	socketPath string
}

func (o *options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.logLevel, "log-level", "info", "Log level")
	fs.StringVar(&o.logFormat, "log-format", "text", "Log format")
	fs.IntVar(&o.port, "port", 10000, "Port of grpc server")
	fs.StringVar(&o.secretsDir, "secrets-dir", "/secrets", "directory holding basic authentication data")
	fs.BoolVar(&o.reflection, "grpc-reflection", false, "enable grpc reflection")
	fs.StringVar(&o.tlsCert, "tls-cert", "/tls/tls.crt", "server certificate to use for tls communication (requires also tls-key)")
	fs.StringVar(&o.tlsKey, "tls-key", "/tls/tls.key", "private key to use for tls communication (requires also tls-cert)")
	fs.StringVar(&o.socketPath, "socket-path", "", "unix domain socket path to listen on (mutually exclusive with --port, --tls-cert, and --tls-key)")
}

func (o *options) Complete() error {
	if o.socketPath != "" {
		o.tlsCert = ""
		o.tlsKey = ""
	}
	return nil
}

func (o *options) Validate() error {
	if o.socketPath != "" && o.port != 10000 {
		return errors.New("--socket-path and --port are mutually exclusive")
	}
	if o.socketPath != "" {
		if o.tlsCert != "" || o.tlsKey != "" {
			return errors.New("--tls-cert and --tls-key cannot be used with --socket-path")
		}
		return nil
	}
	// TCP mode: tls-cert and tls-key must be provided together or not at all
	if (o.tlsCert != "") != (o.tlsKey != "") {
		return errors.New("--tls-cert and --tls-key must be provided together")
	}
	return nil
}

func (o *options) LogConfig() (string, string) {
	return o.logLevel, o.logFormat
}
