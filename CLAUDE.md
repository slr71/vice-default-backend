# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is **vice-default-backend**, a Go service that acts as the catch-all default backend for VICE (Visual Interactive Computing Environment) analysis subdomains. When a request hits a `*.cyverse.run` subdomain before the analysis-specific HTTPRoute is active, this service serves a waiting page that periodically refreshes until the vice-operator loading page or the analysis itself takes over.

## Build and Run Commands

```bash
# Build
go build ./...

# Run locally
./vice-default-backend --listen 0.0.0.0:60000

# Run with custom refresh interval
./vice-default-backend --refresh-seconds 10

# Lint (uses shared workflow with golangci-lint)
golangci-lint run
```

## CLI Flags

- `--listen` - Listen address (default: `0.0.0.0:60000`)
- `--refresh-seconds` - Seconds between page reloads while waiting (default: `5`)
- `--log-level` - One of: trace, debug, info, warn, error, fatal, or panic

## Architecture

Single-file Go service (`main.go`) with an embedded HTML template (`templates/waiting.html`):
- **HandleWaiting** - Serves the waiting page for all requests; sets `X-Vice-Default-Backend: true` header
- **Health endpoint** - `/healthz` for Kubernetes probes
- **Waiting page** - Self-contained HTML/CSS/JS that counts down and reloads the page on the configured interval

## Deployment

Uses Skaffold for Kubernetes deployment. Image is built and pushed to `harbor.cyverse.org/de/vice-default-backend`. Kubernetes manifests are in `k8s/vice-default-backend.yml`.
