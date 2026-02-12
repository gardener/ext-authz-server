# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

############# builder
FROM golang:1.26.0 AS builder

WORKDIR /build

# Copy go mod and sum files
COPY go.mod go.sum ./
# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

COPY . .
ARG TARGETARCH
RUN make release GOARCH=$TARGETARCH

############# ext-authz-server
FROM gcr.io/distroless/static-debian13:nonroot AS ext-authz-server

COPY --from=builder /build/ext-authz-server /ext-authz-server
ENTRYPOINT ["/ext-authz-server"]
