ARG WINDOWS_BASE_IMAGE=mcr.microsoft.com/windows/nanoserver:ltcs2022

FROM --platform=$BUILDPLATFORM golang:1.24 AS builder
ARG TARGETARCH
ARG TARGETOS

# Get dependencies
WORKDIR /w
COPY go.mod go.sum ./
RUN go mod download

# Build binary
COPY . ./
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH make bin/rr

# Build the asset container, copy over rr
FROM gcr.io/distroless/static:nonroot AS simple
COPY --from=builder /w/bin/ /rr
ENTRYPOINT ["/rr"]

FROM $WINDOWS_BASE_IMAGE AS windows
COPY --from=builder /w/bin/ /rr.exe
ENTRYPOINT ["/rr.exe"]