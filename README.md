# External Authorization Server

[![REUSE status](https://api.reuse.software/badge/github.com/gardener/ext-authz-server)](https://api.reuse.software/info/github.com/gardener/ext-authz-server)

The external authorization server implements the [corresponding envoy-proxy API](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/ext_authz_filter) for HTTP basic authentication.
It watches a directory, which can be configured via the `--secrets-dir` flag, for basic authentication files.
The file name reflects the subdomain for which the credentials are valid.

## Usage

The external authorization server uses  [google.golang.org/grpc](https://pkg.go.dev/google.golang.org/grpc#section-readme) for implementing the gRPC server.
Its default port is `10000`, which can be changed via the `--port` flag.

You can configure the logging level of the gRPC server via the following environment variables:

```
$ export GRPC_GO_LOG_VERBOSITY_LEVEL=99
$ export GRPC_GO_LOG_SEVERITY_LEVEL=info
```
