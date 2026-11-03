# BlueBubbles MCP

A lightweight MCP gateway for a BlueBubbles REST API. The gateway keeps the upstream credential server-side and exposes a narrow, policy-controlled MCP surface.

## Status

This repository is under active development. The initial service shell supports authenticated Streamable HTTP, `/healthz`, and a read-only default mode. BlueBubbles read tools and interactive operations are added only with tests and explicit policy coverage.

## Configuration

- `BLUEBUBBLES_MODE`: `readonly` (default) or `interactive`.
- `BLUEBUBBLES_URL`: upstream BlueBubbles base URL.
- `BLUEBUBBLES_PASSWORD_FILE`: preferred upstream password source.
- `BLUEBUBBLES_HEADERS_FILE`: optional JSON object of static upstream headers.
- `MCP_AUTH_TOKEN_FILE`: preferred credential for the gateway's Bearer authentication.
- `MCP_LISTEN_ADDR`: listen address; defaults to loopback.

Secrets should be supplied by the deployment platform or secret files. Do not place real credentials in this repository or issue them as MCP arguments.

## Endpoints

- `GET /healthz` — local process/configuration health.
- `POST /mcp` — authenticated MCP Streamable HTTP endpoint.

The current service uses stateless JSON Streamable HTTP and a 4 MiB request limit. Put a public deployment behind TLS and an authenticated network boundary.

## Security model

Read-only is fail-closed and is the default. MCP tool annotations are descriptive only; the gateway's operation policy is authoritative. Interactive mode is an explicit startup configuration and does not create a generic upstream request proxy.
