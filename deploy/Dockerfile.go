# Build any Go service. Context is the repo root; pass SERVICE + CMD.
#   docker build -f deploy/Dockerfile.go --build-arg SERVICE=services/price --build-arg CMD=./cmd/pricesvc .
FROM golang:1.25 AS build
ARG SERVICE
ARG CMD
WORKDIR /src
COPY ${SERVICE}/ ./
RUN CGO_ENABLED=0 GOFLAGS=-mod=mod go build -trimpath -o /out/app ${CMD}

FROM gcr.io/distroless/base-debian12
COPY --from=build /out/app /app
ENTRYPOINT ["/app"]
