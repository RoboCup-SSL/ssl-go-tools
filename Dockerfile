FROM golang:1.24-alpine@sha256:7772cb5322baa875edd74705556d08f0eeca7b9c4b5367754ce3f2f00041ccee AS build_go
ARG cmd
WORKDIR /work
COPY . .
RUN go install ./cmd/${cmd}

# Start fresh from a smaller image
FROM alpine:3@sha256:8a1f59ffb675680d47db6337b49d22281a139e9d709335b492be023728e11715
ARG cmd
COPY --from=build_go /go/bin/${cmd} /app
WORKDIR /data
RUN chown 1000: /data
USER 1000
ENTRYPOINT ["/app"]
CMD []
