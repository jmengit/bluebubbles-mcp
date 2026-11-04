# Configuration

The service reads configuration at startup. Unknown capability modes fail startup.

| Variable | Required | Default | Description |
|---|---:|---|---|
| `BLUEBUBBLES_MODE` | no | `readonly` | `readonly` or `interactive` |
| `BLUEBUBBLES_URL` | yes for upstream tools | none | BlueBubbles base URL |
| `BLUEBUBBLES_PASSWORD_FILE` | preferred | none | File containing the upstream password |
| `BLUEBUBBLES_HEADERS_FILE` | no | none | JSON object of static upstream headers |
| `BLUEBUBBLES_HEADERS_JSON` | no | none | Direct JSON object of static upstream headers |
| `MCP_AUTH_TOKEN_FILE` | preferred | none | File containing the gateway bearer token |
| `MCP_AUTH_TOKEN` | direct option | none | Gateway bearer token from `.env` |
| `MCP_ALLOWED_HOSTS` | no | unrestricted | Comma-separated exact Host values |
| `MCP_ALLOWED_ORIGINS` | no | unrestricted | Comma-separated exact `scheme://host[:port]` origins |
| `MCP_LISTEN_ADDR` | no | `127.0.0.1:8080` | HTTP listen address |

`BLUEBUBBLES_PASSWORD` and `MCP_AUTH_TOKEN` are supported direct environment alternatives for `.env`-based deployments. When both a file and direct value are provided, the file takes precedence. `BLUEBUBBLES_HEADERS_JSON` is the direct environment alternative for `BLUEBUBBLES_HEADERS_FILE`; it must contain a JSON object. Treat the Arcane environment payload as sensitive because direct environment credentials may be visible to deployment administrators and inspection APIs.


