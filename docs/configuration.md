# Configuration

The service reads configuration at startup. Unknown capability modes fail startup.

| Variable | Required | Default | Description |
|---|---:|---|---|
| `BLUEBUBBLES_MODE` | no | `readonly` | `readonly` or `interactive` |
| `BLUEBUBBLES_URL` | yes for upstream tools | none | BlueBubbles base URL |
| `BLUEBUBBLES_PASSWORD_FILE` | preferred | none | File containing the upstream password |
| `BLUEBUBBLES_HEADERS_FILE` | no | none | JSON object of static upstream headers |
| `MCP_AUTH_TOKEN_FILE` | preferred | none | File containing the gateway bearer token |
| `MCP_LISTEN_ADDR` | no | `127.0.0.1:8080` | HTTP listen address |

Direct secret environment variables may be supported for development, but secret files are preferred for deployment. Never pass credentials as MCP tool arguments or commit them to a repository.
