FROM golang:1.20-alpine AS build_go
WORKDIR work
COPY . .
RUN go install ./cmd/ssl-log-player

# Start fresh from a smaller image
FROM alpine:3
COPY --from=build_go /go/bin/ssl-log-player /app/ssl-log-player
WORKDIR /data
RUN chown 1000: /data
USER 1000
ENTRYPOINT ["/app/ssl-log-player"]
CMD []
