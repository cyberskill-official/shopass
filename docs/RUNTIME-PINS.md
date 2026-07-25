# Runtime pins

One toolchain version per stack. Align local, Docker, and CI to these pins.

| Stack | Pin | Source of truth |
| --- | --- | --- |
| Go | **1.25** (`go 1.25.0` in every `go.mod`) | `.github/workflows/ci.yml` `go-version: "1.25"`; `deploy/Dockerfile.go` `golang:1.25` |
| Node | **24** (`>=24 <25` where `engines` is set) | CI `node-version: "24.18.0"`; `deploy/Dockerfile.node` / `Dockerfile.web` `node:24-slim` |
| Python (ML) | **3.11** | CI `python-version: "3.11"`; `deploy/Dockerfile.ml` `python:3.11-slim`; `services/ml/.python-version` |

Do not bump dependency trees when changing these pins — only the toolchain directive / base image / `engines` field.

**CI note:** `gofmt` enforcement in `.github/workflows/ci.yml` is owned by the H5 hardening item; formatting fixes land separately (L1).
