/*
 * SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package v3

import (
	"context"
	"log"

	envoy_service_auth_v3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"

	"github.com/gardener/ext-authz-server/pkg/auth"
)

type server struct {
	services auth.Services
}

var _ envoy_service_auth_v3.AuthorizationServer = &server{}

// New creates a new authorization server.
func New(services auth.Services) envoy_service_auth_v3.AuthorizationServer {
	return &server{services}
}

// Check implements authorization's Check interface which performs authorization check based on the
// attributes associated with the incoming request.
func (s *server) Check(
	ctx context.Context,
	req *envoy_service_auth_v3.CheckRequest) (*envoy_service_auth_v3.CheckResponse, error) {
	authorization := req.Attributes.Request.Http.Headers["reversed-vpn"]

	if len(authorization) == 0 {
		log.Printf("request without header denied!\n")
		return &envoy_service_auth_v3.CheckResponse{
			Status: &status.Status{
				Code: int32(code.Code_PERMISSION_DENIED),
			},
		}, nil
	}

	valid, err := s.services.Check(authorization)
	if err != nil || !valid {
		log.Printf("request with header: \"%s\" denied!\n", authorization)
		return &envoy_service_auth_v3.CheckResponse{
			Status: &status.Status{
				Code: int32(code.Code_PERMISSION_DENIED),
			},
		}, err
	}

	log.Printf("request with header: \"%s\" accepted!\n", authorization)
	return &envoy_service_auth_v3.CheckResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
	}, nil

}
