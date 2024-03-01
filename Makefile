# SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

ENSURE_GARDENER_MOD                    := $(shell go get github.com/gardener/gardener@$$(go list -m -f "{{.Version}}" github.com/gardener/gardener))
GARDENER_HACK_DIR                      := $(shell go list -m -f "{{.Dir}}" github.com/gardener/gardener)/hack
VERSION                                := $(shell cat VERSION)
REGISTRY                               := europe-docker.pkg.dev/gardener-project/public/gardener
PREFIX                                 := ext-authz-server
EXTERNAL_AUTHZ_SERVER_IMAGE_REPOSITORY := $(REGISTRY)/$(PREFIX)
EXTERNAL_AUTHZ_SERVER_IMAGE_TAG        := $(VERSION)
REPO_ROOT                              := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
HACK_DIR                               := $(REPO_ROOT)/hack

#########################################
# Tools                                 #
#########################################

TOOLS_DIR := $(HACK_DIR)/tools
include $(GARDENER_HACK_DIR)/tools.mk

#################################################################
# Rules related to binary build, Docker image build and release #
#################################################################

.PHONY: ext-authz-server-docker-image
ext-authz-server-docker-image:
	@docker build -t $(EXTERNAL_AUTHZ_SERVER_IMAGE_REPOSITORY):$(EXTERNAL_AUTHZ_SERVER_IMAGE_TAG) -f Dockerfile --rm .
.PHONY: docker-images
docker-images: ext-authz-server-docker-image

.PHONY: release
release: docker-images docker-login docker-push

.PHONY: docker-login
docker-login:
	@gcloud auth activate-service-account --key-file .kube-secrets/gcr/gcr-readwrite.json

.PHONY: docker-push
docker-push:
	@if ! docker images $(EXTERNAL_AUTHZ_SERVER_IMAGE_REPOSITORY) | awk '{ print $$2 }' | grep -q -F $(EXTERNAL_AUTHZ_SERVER_IMAGE_TAG); then echo "$(EXTERNAL_AUTHZ_SERVER_IMAGE_REPOSITORY) version $(EXTERNAL_AUTHZ_SERVER_IMAGE_TAG) is not yet built. Please run 'ext-authz-server-docker-image'"; false; fi
	@docker -- push $(EXTERNAL_AUTHZ_SERVER_IMAGE_REPOSITORY):$(EXTERNAL_AUTHZ_SERVER_IMAGE_TAG)

#####################################################################
# Rules for verification, formatting, linting, testing and cleaning #
#####################################################################

.PHONY: tidy
tidy:
	@GO111MODULE=on go mod tidy
	@mkdir -p $(REPO_ROOT)/.ci/hack && cp $(GARDENER_HACK_DIR)/.ci/* $(REPO_ROOT)/.ci/hack/ && chmod +xw $(REPO_ROOT)/.ci/hack/*

.PHONY: check
check: $(GOIMPORTS) $(GOLANGCI_LINT)
	@bash $(GARDENER_HACK_DIR)/check.sh ./cmd/... ./pkg/...

.PHONY: format
format: $(GOIMPORTS) $(GOIMPORTSREVISER)
	@bash $(GARDENER_HACK_DIR)/format.sh ./cmd ./pkg

.PHONY: test
test:
	@bash $(GARDENER_HACK_DIR)/test.sh ./cmd/... ./pkg/...

.PHONY: test-cov
test-cov:
	@bash $(GARDENER_HACK_DIR)/test-cover.sh ./cmd/... ./pkg/...

.PHONY: test-cov-clean
test-cov-clean:
	@bash $(GARDENER_HACK_DIR)/test-cover-clean.sh

.PHONY: verify
verify: check format test

.PHONY: verify-extended
verify-extended: check format test-cov test-cov-clean
