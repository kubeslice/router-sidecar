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

FROM golang:1.24-alpine3.21 AS gobuilder

ARG TARGETOS
ARG TARGETARCH

# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git make build-base

# Set the Go source path
WORKDIR /kubeslice/kubeslice-router-sidecar/
COPY . .
# Build the binary.
RUN go mod download && \
    CGO_ENABLED=1 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -o bin/kubeslice-router-sidecar main.go

# Build reduced image from distroless-base image
FROM gcr.io/distroless/static

# Add the necessary pakages:
# tc - is needed for traffic control and shaping on the sidecar.  it is part of the iproute2
COPY --from=gobuilder /kubeslice/kubeslice-router-sidecar/bin/kubeslice-router-sidecar /router-sidecar
ENTRYPOINT ["/router-sidecar"]

# Run the sidecar binary.
WORKDIR /kubeslice

# Copy our static executable.
COPY --from=gobuilder /kubeslice/kubeslice-router-sidecar/bin/kubeslice-router-sidecar .
EXPOSE 5000
EXPOSE 8080
# Or could be CMD
ENTRYPOINT ["./kubeslice-router-sidecar"]
