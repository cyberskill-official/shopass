# Build any Go service. Context is the repo root; pass SERVICE + CMD.
# Keeping the source hierarchy is intentional: several service modules use
# local Go-module replacements such as ../../obs and ../price.
#   docker build -f deploy/Dockerfile.go --build-arg SERVICE=services/price --build-arg CMD=./cmd/pricesvc .
FROM golang:1.25 AS build
ARG SERVICE
ARG CMD
WORKDIR /src
COPY obs/ ./obs/
COPY region/ ./region/
COPY services/ ./services/
RUN cd "${SERVICE}" && CGO_ENABLED=0 GOFLAGS=-mod=readonly go build -trimpath -o /out/app ${CMD}

FROM gcr.io/distroless/base-debian12
COPY --from=build /out/app /app
ENTRYPOINT ["/app"]
