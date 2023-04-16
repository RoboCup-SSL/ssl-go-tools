FROM golang:1.20-alpine AS build_go
WORKDIR work
COPY . .
RUN go install ./cmd/ssl-auto-recorder

# Start fresh from a smaller image
FROM alpine:3
COPY --from=build_go /go/bin/ssl-auto-recorder /app/ssl-auto-recorder
WORKDIR /data
RUN chown 1000: /data
USER 1000
ENTRYPOINT ["/app/ssl-auto-recorder"]
CMD []
