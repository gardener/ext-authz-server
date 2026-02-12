# External Authorization Server

[![REUSE status](https://api.reuse.software/badge/github.com/gardener/ext-authz-server)](https://api.reuse.software/info/github.com/gardener/ext-authz-server)

The external authorization server implements the [corresponding envoy-proxy API](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/ext_authz_filter) for HTTP basic authentication.
It watches a directory, which can be configured via the `--secrets-dir` flag, for basic authentication files.
The file name reflects the subdomain for which the credentials are valid.
