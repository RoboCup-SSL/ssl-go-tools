FROM golang:1.23-alpine AS build_go
ARG cmd
WORKDIR work
COPY . .
RUN go install ./cmd/${cmd}

# Start fresh from a smaller image
FROM alpine:3
ARG cmd
COPY --from=build_go /go/bin/${cmd} /app
WORKDIR /data
RUN chown 1000: /data
USER 1000
ENTRYPOINT ["/app"]
CMD []
