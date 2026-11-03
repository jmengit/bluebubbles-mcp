# Upstream compatibility

BlueBubbles authenticates its REST API with a server password in a query parameter. The gateway must construct that query itself and redact it from logs. BlueBubbles route names and response envelopes can vary by server version; compatibility tests should use a fake upstream that models the documented envelope and a versioned live smoke test before adding new operations.

Do not infer safety from HTTP verbs alone. Some read-like POST query operations are semantically reads, while some GET routes can restart or otherwise control the server.
