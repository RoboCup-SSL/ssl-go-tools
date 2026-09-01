FROM --platform=$BUILDPLATFORM golang:1.27-alpine@sha256:4c9fe60190a2a3350ddc51de80d0224b8a6698d12bdfc999fee45ea9d6c46dbc AS build_go
ARG TARGETOS
ARG TARGETARCH
ARG BINARY_NAME
WORKDIR /work
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-w -s" -o /go/bin/${BINARY_NAME} ./cmd/${BINARY_NAME}

# Start fresh from a smaller image
FROM alpine:3@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b
ARG BINARY_NAME
COPY --from=build_go /go/bin/${BINARY_NAME} /app
WORKDIR /data
RUN chown 1000: /data
USER 1000
ENTRYPOINT ["/app"]
CMD []
