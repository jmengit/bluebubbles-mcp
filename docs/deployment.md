# Deployment

Use a pinned image digest in the deployment platform. Supply the BlueBubbles password, MCP bearer token, and optional upstream headers through secret files or the platform's secret mechanism.

BlueBubbles supports HTTP or HTTPS upstream URLs. If the upstream uses a private CA, add a future CA-file configuration rather than disabling certificate verification. Direct `.env` credentials are supported for this deployment shape, but they are visible to Arcane project administrators and may be returned by project inspection APIs; use a protected Arcane variable/secret facility when available.

The container runs as a non-root user with no persistent storage. Publish the HTTP port only to the trusted network that needs MCP access. Put TLS and any identity-aware access control in front of `/mcp` when the service is not strictly private.

Verify after deployment:

1. `GET /healthz` returns HTTP 200.
2. `/mcp` without the bearer returns HTTP 401.
3. An authenticated MCP client can initialize and list tools.
4. Read calls reach the configured upstream.
5. In readonly mode, no mutation tool is advertised or executable.
