# Design

## Purpose

This service is a narrow policy boundary between an MCP client and a BlueBubbles REST API. It owns the upstream credential and exposes typed operations rather than arbitrary upstream URLs.

## Capability modes

`BLUEBUBBLES_MODE=readonly` is the default. Only read operations are registered and the policy layer rejects classified mutations. `interactive` is explicit process-start configuration; it does not grant arbitrary BlueBubbles API access. Operations such as sending, read receipts, deletion, server control, webhook registration, and group administration require separate implementation and tests.

## Trust boundaries

- The upstream BlueBubbles URL, password, and static headers are deployment configuration.
- The MCP bearer token authenticates callers to this gateway.
- The gateway never forwards arbitrary inbound headers or paths.
- Logs must not contain credentials, full URLs, message text, or contact identifiers.

## Transport

The service uses the official Go MCP SDK's stateless Streamable HTTP transport with JSON responses at `/mcp`. `/healthz` is separate and does not contact BlueBubbles. A TLS-terminating reverse proxy or private network boundary is required for non-loopback deployments.
