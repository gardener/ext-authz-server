// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package extauthzserver

import (
	"context"
	"fmt"
	"io/fs"
	"strings"

	envoycorev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoyserviceauthv3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	envoytypev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/go-logr/logr"
	"google.golang.org/genproto/googleapis/rpc/code"
	googlestatus "google.golang.org/genproto/googleapis/rpc/status"
)

type server struct {
	log   logr.Logger
	store *store
}

var _ envoyserviceauthv3.AuthorizationServer = &server{}

// New creates a new authorization server.
func New(log logr.Logger, dir fs.FS) (envoyserviceauthv3.AuthorizationServer, error) {
	store, err := newStore(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to set up store: %w", err)
	}
	return &server{
		log:   log,
		store: store,
	}, nil
}

// Check implements authorization's Check interface which performs authorization check based on the
// attributes associated with the incoming request.
func (s *server) Check(
	_ context.Context,
	req *envoyserviceauthv3.CheckRequest,
) (*envoyserviceauthv3.CheckResponse, error) {
	if req.Attributes == nil || req.Attributes.Request == nil || req.Attributes.Request.Http == nil {
		return denyResponse(s.log, "invalid request"), nil
	}
	http := req.Attributes.Request.Http

	log := s.log.WithValues("id", http.Id, "path", http.Path, "host", http.Host)

	auth := http.Headers["authorization"]
	if auth == "" {
		return denyResponse(log, "missing authorization header"), nil
	}

	host := http.Host
	hostParts := strings.Split(host, ".")
	if len(hostParts) == 0 || host == "" {
		return denyResponse(log, "missing host header"), nil
	}

	if err := s.store.IsValid(hostParts[0], auth); err != nil {
		return denyResponse(log, fmt.Sprintf("invalid authorization: %s", err.Error())), nil
	}

	log.Info("auth request allowed")

	return &envoyserviceauthv3.CheckResponse{
		Status: &googlestatus.Status{
			Code: int32(code.Code_OK),
		},
		HttpResponse: &envoyserviceauthv3.CheckResponse_OkResponse{
			OkResponse: &envoyserviceauthv3.OkHttpResponse{
				HeadersToRemove: []string{"authorization"},
			},
		},
	}, nil
}

func denyResponse(log logr.Logger, message string) *envoyserviceauthv3.CheckResponse {
	log.Info("auth request denied", "message", message)

	return &envoyserviceauthv3.CheckResponse{
		Status: &googlestatus.Status{
			Code:    int32(code.Code_PERMISSION_DENIED),
			Message: message,
		},
		HttpResponse: &envoyserviceauthv3.CheckResponse_DeniedResponse{
			DeniedResponse: &envoyserviceauthv3.DeniedHttpResponse{
				Headers: []*envoycorev3.HeaderValueOption{{
					Header: &envoycorev3.HeaderValue{Key: "WWW-Authenticate", Value: "Basic realm=\"Authentication Required\""},
				}},
				Status: &envoytypev3.HttpStatus{Code: envoytypev3.StatusCode_Unauthorized},
			},
		},
	}
}
