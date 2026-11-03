# Interactive mode

Interactive mode is disabled unless `BLUEBUBBLES_MODE=interactive` is set at process startup. It is not a tool-call argument and cannot be changed remotely.

Interactive operations must have explicit policy coverage, bounded inputs, sanitized errors, and no blind retries for non-idempotent requests. Read receipts are externally visible side effects and must remain separately opt-in. The initial release intentionally does not expose arbitrary URL forwarding, server control, webhook registration, chat deletion, message deletion, group administration, or unsend/edit operations.
