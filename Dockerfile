##########################################################
#Dockerfile
#Copyright (c) 2022 Avesha, Inc. All rights reserved.
#
#SPDX-License-Identifier: Apache-2.0
#
#Licensed under the Apache License, Version 2.0 (the "License");
#you may not use this file except in compliance with the License.
#You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
#Unless required by applicable law or agreed to in writing, software
#distributed under the License is distributed on an "AS IS" BASIS,
#WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#See the License for the specific language governing permissions and
#limitations under the License.
##########################################################

# Build stage
FROM golang:1.24-bookworm AS gobuilder
ARG TARGETOS
ARG TARGETARCH
# Install git and build tools
# Git is required for fetching the dependencies.
RUN apt-get update && apt-get install -y git make build-essential && \
    rm -rf /var/lib/apt/lists/*
# Set the Go source path
WORKDIR /kubeslice/kubeslice-router-sidecar/

# For better caching of layers
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Build the binary.
RUN CGO_ENABLED=1 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -o bin/kubeslice-router-sidecar main.go

# Tools stage - iproute2 tools and dependencies from Debian
FROM debian:12-slim AS tools
RUN apt-get update && \
    apt-get install -y --no-install-recommends iproute2 && \
    cp /sbin/tc /tmp/tc && \
    mkdir -p /tmp/lib && \
    # Use ldd to find shared library dependencies and copy them
    ldd /sbin/tc | awk '/=>/ { print $3 }' | xargs -I {} cp {} /tmp/lib/ && \
    # Clean up
    rm -rf /var/lib/apt/lists/*

# Final stage - distroless with pinned tag
FROM gcr.io/distroless/cc-debian12@sha256:00cc20b928afcc8296b72525fa68f39ab332f758c4f2a9e8d90845d3e06f1dc4
# Copy the tc binary and its dependencies
COPY --from=tools /tmp/tc /usr/sbin/tc
COPY --from=tools /tmp/lib/* /lib/x86_64-linux-gnu/
# Copy static executable
COPY --from=gobuilder /kubeslice/kubeslice-router-sidecar/bin/kubeslice-router-sidecar /kubeslice-router-sidecar
WORKDIR /
EXPOSE 5000 8080
USER nonroot:nonroot
ENTRYPOINT ["/kubeslice-router-sidecar"]
