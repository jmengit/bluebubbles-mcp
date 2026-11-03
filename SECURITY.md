# Security

Do not report credentials, message contents, private URLs, or other sensitive data in public issues. Use the repository's private security advisory process where available.

The gateway is intended for a trusted deployment boundary. Keep the MCP endpoint behind TLS and an authenticated network path when it is not loopback-only. `BLUEBUBBLES_MODE=readonly` is the default and is enforced by the service policy, not only by tool annotations.
