# BlueBubbles MCP

A lightweight MCP gateway for a BlueBubbles REST API. The gateway keeps the upstream credential server-side and exposes a narrow, policy-controlled MCP surface.

## Status

This repository is under active development. The initial service shell supports authenticated Streamable HTTP, `/healthz`, and a read-only default mode. BlueBubbles read tools and interactive operations are added only with tests and explicit policy coverage.

## Configuration

- `BLUEBUBBLES_MODE`: `readonly` (default) or `interactive`.
- `BLUEBUBBLES_URL`: upstream BlueBubbles base URL.
- `BLUEBUBBLES_PASSWORD_FILE`: preferred upstream password source.
- `BLUEBUBBLES_HEADERS_FILE`: optional JSON object of static upstream headers.
- `BLUEBUBBLES_HEADERS_JSON`: optional JSON object of static upstream headers; useful when the deployment supplies configuration through an environment file.
- `MCP_AUTH_TOKEN_FILE`: preferred credential for the gateway's Bearer authentication.
- `MCP_AUTH_TOKEN`: direct gateway Bearer credential; supported for `.env` deployments.
- `MCP_ALLOWED_HOSTS`: optional comma-separated exact Host values accepted for `/mcp`.
- `MCP_ALLOWED_ORIGINS`: optional comma-separated exact `scheme://host[:port]` origins accepted for `/mcp`.
- `MCP_LISTEN_ADDR`: listen address; defaults to loopback.

Example `.env` configuration for a LAN or WAN/reverse-proxy deployment:

```dotenv
BLUEBUBBLES_MODE=readonly
BLUEBUBBLES_URL=http://bluebubbles.lan:1234
BLUEBUBBLES_PASSWORD=replace-me
BLUEBUBBLES_HEADERS_JSON={"CF-Access-Client-Id":"replace-me","CF-Access-Client-Secret":"replace-me"}
MCP_AUTH_TOKEN=replace-me
MCP_LISTEN_ADDR=0.0.0.0:8080
MCP_ALLOWED_HOSTS=mcp.example.test
MCP_ALLOWED_ORIGINS=https://mcp.example.test
```

The `.env` file may provide `BLUEBUBBLES_URL` as either an HTTP or HTTPS LAN/WAN URL. The gateway does not require the URL to be public; it simply needs to be reachable from the container. `BLUEBUBBLES_HEADERS_JSON` is parsed once at startup and supports fixed headers such as Cloudflare Access service-token headers. Never pass arbitrary inbound MCP headers through to the upstream.

Secrets should be supplied by the deployment platform or secret files. Do not place real credentials in this repository or issue them as MCP arguments.

## Endpoints

- `GET /healthz` — local process/configuration health.
- `POST /mcp` — authenticated MCP Streamable HTTP endpoint.

The current service uses stateless JSON Streamable HTTP and a 4 MiB request limit. Put a public deployment behind TLS and an authenticated network boundary.

## Security model

Read-only is fail-closed and is the default. MCP tool annotations are descriptive only; the gateway's operation policy is authoritative. Interactive mode is an explicit startup configuration and does not create a generic upstream request proxy.
