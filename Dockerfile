# SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.22.0 AS builder

COPY . /app
WORKDIR /app
ENV CGO_ENABLED=0

# Copy go mod and sum files
COPY go.mod go.sum ./
# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

RUN go install ./...

FROM gcr.io/distroless/static-debian11:nonroot
WORKDIR /

COPY --from=builder /go/bin/ext-authz-server /app/server
CMD ["/app/server"]
