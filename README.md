This URL is embedded into your target file. When the file is accessed and the URL is fetched, the listener verifies the secret using constant-time comparison, logs the event to `~/.gobaitr/tokens.db`, prints an alert to the terminal, and optionally fires a webhook to any HTTP endpoint.

All data stays on your machine. Nothing is sent anywhere unless you configure a webhook.

---

## Why Not canarytokens.org

| | canarytokens.org | gobaitr |
|--|--|--|
| Self-hosted | No | Yes |
| Air-gapped environments | No | Yes |
| Audit log you own | No | Yes |
| Zero third-party trust | No | Yes |
| Works offline | No | Yes |
| Single binary install | No | Yes |

---

## Requirements

- Go 1.22 or later — [download](https://go.dev/dl/)
- No CGO required
- No external dependencies beyond Go modules
- SQLite database is handled automatically — nothing to install

Tested on Linux, macOS, and Windows.

## Token Types

| Type | How It Works | Use Case |
|------|-------------|----------|
| `url` | Prints raw callback URL to stdout | Markdown files, documents, anywhere you control the text |
| `file` | Embeds URL into target file (.env, .json, .txt) | Config files, credential files, JSON secrets |
| `env` | Outputs `export API_KEY="<callback-url>"` | Fake API key canaries in shell environments |

---

## Commands Reference

### generate

```bash
gobaitr generate <type> [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--note` | string | | Optional label for this token |
| `--port` | int | 8080 | Listener port used in callback URL |
| `--expires-in` | string | | Token TTL e.g. 24h, 168h |

### embed

```bash
gobaitr embed --token <id> --target <file> [flags]
```

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--token` | string | Yes | Token ID to embed |
| `--target` | string | Yes | Target file path (.env, .json, .txt) |
| `--dry-run` | bool | No | Preview change without writing to disk |

### listen

```bash
gobaitr listen [flags]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--port` | int | 8080 | Local port to bind |
| `--webhook` | string | | URL to POST alert JSON on trigger |
| `--quiet` | bool | false | Suppress terminal output |
| `--tls` | bool | false | Enable TLS |
| `--cert` | string | | Path to TLS certificate file |
| `--key` | string | | Path to TLS private key file |

### verify

```bash
gobaitr verify <token-id> [flags]
```

| Flag | Type | Description |
|------|------|-------------|
| `--json` | bool | Output as JSON instead of table |
| `--all` | bool | Show full request headers for each event |

### list

```bash
gobaitr list [flags]
```

| Flag | Type | Description |
|------|------|-------------|
| `--triggered` | bool | Show only triggered tokens |
| `--type` | string | Filter by type: url, file, env |

### delete

```bash
gobaitr delete <token-id> [flags]
```

| Flag | Type | Description |
|------|------|-------------|
| `--force` | bool | Skip confirmation prompt |

---

## Webhook Format

On trigger, gobaitr POSTs the following JSON to your configured webhook URL:

```json
{
  "event": "token_triggered",
  "token_id": "abc-123-...",
  "token_type": "file",
  "token_note": "prod credentials",
  "triggered_at": "2026-04-10T10:30:00Z",
  "remote_ip": "203.0.113.42",
  "user_agent": "curl/8.18.0",
  "headers": {
    "User-Agent": "curl/8.18.0",
    "Accept": "*/*"
  },
  "gobaitr_version": "1.0.0"
}
```

Compatible with Slack incoming webhooks, PagerDuty, and any HTTP endpoint.

```bash
gobaitr listen --webhook https://hooks.slack.com/services/xxx/yyy/zzz
```

---

## Security Notes

### Secret-in-URL

The callback secret is embedded in the URL path. On plaintext HTTP, this secret is visible in server access logs, browser history, and any network proxy on the path. For local and trusted-network use, this is acceptable. For internet-facing deployments, always use TLS:

```bash
gobaitr listen --tls --cert cert.pem --key key.pem
```

### Timing-safe verification

Secret comparison uses `crypto/subtle.ConstantTimeCompare`. A probing attacker cannot distinguish valid from invalid token IDs based on response timing.

### Rate limiting

The callback endpoint limits each source IP to 5 requests per 10 seconds. Excess requests receive a silent 200 OK. No error codes are returned that could reveal the endpoint's purpose.

### Local-only storage

All token data lives in `~/.gobaitr/tokens.db`. Nothing leaves your machine unless you configure a webhook. SQLite WAL mode prevents concurrent-access corruption when the listener and CLI access the database simultaneously.

### Token TTL

Use `--expires-in` to create time-bounded canaries. Expired token hits receive a silent 200 OK and are not logged.

### Known limitation

All tokens share the `/t/` route prefix. An attacker who discovers the listener address could attempt to enumerate token IDs. Route randomization is planned for V2.

---

## Contributing

Open an issue before submitting a large change so the direction can be agreed on first. Bug fixes and small improvements are welcome without prior discussion.

[Open an issue](https://github.com/sudesh856/gobaitr/issues)

---

## License

MIT. See [LICENSE](LICENSE).