FROM golang:1.25-alpine@sha256:b6ed3fd0452c0e9bcdef5597f29cc1418f61672e9d3a2f55bf02e7222c014abd AS build_go
ARG cmd
WORKDIR /work
COPY . .
RUN go install ./cmd/${cmd}

# Start fresh from a smaller image
FROM alpine:3@sha256:4bcff63911fcb4448bd4fdacec207030997caf25e9bea4045fa6c8c44de311d1
ARG cmd
COPY --from=build_go /go/bin/${cmd} /app
WORKDIR /data
RUN chown 1000: /data
USER 1000
ENTRYPOINT ["/app"]
CMD []
