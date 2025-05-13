FROM golang:1.24 AS builder
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
COPY --from=builder /w/bin/rr /rr
ENTRYPOINT ["/rr"]
