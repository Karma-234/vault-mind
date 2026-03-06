# VaultMind

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**VaultMind** is a **local-first, privacy-focused MCP (Model Context Protocol) server** written in Go that lets AI agents (Claude Desktop, Cursor, Continue.dev, GitHub Copilot in VS Code, etc.) securely manage credentials — passwords, API keys, seed phrases, tokens — **without ever exposing them to cloud providers or including them in chat context**.

All secrets are encrypted at rest using AES-256-GCM with keys derived from your master passphrase via Argon2id. Data is stored in an embedded Pebble database. Decryption happens only ephemerally in memory when explicitly requested — zero trust leakage to LLMs.

## Why VaultMind?

Pasting real API keys or seed phrases into AI chats risks permanent exposure (provider logs, shared conversations, context leaks, breaches). VaultMind solves this:

- **Zero transmission** to LLM providers — secrets never enter chat history
- **AI-powered** credential operations without plaintext exposure
- **Offline-first** — no internet required
- **Shared vault** across multiple clients/sessions (via HTTP transport)
- **Strong warnings** and friction for high-risk items (seed phrases)

Perfect for developers, crypto users, and privacy-conscious individuals who want AI assistance without compromising security.

## Features

- Generate cryptographically secure passwords/passphrases
- Add & categorize credentials (API keys, passwords, seed phrases)
- Retrieve credentials with ephemeral decryption & immediate memory wiping
- Delete credentials with explicit confirmation
- Fully local storage (Pebble embedded KV database)
- HTTP transport for multi-client / multi-session access
- High-risk warnings for seed phrases and sensitive data

## Quick Start

### Prerequisites

- Go 1.22+
- macOS / Linux / Windows

### Installation

````bash
git clone https://github.com/yourusername/vaultmind.git
cd vaultmind
go mod tidy
go build -o vaultmind ./cmd/vaultmind

## Running the Server

### 1. Set your strong master passphrase

This passphrase is used to derive the encryption key. It must be set every time you start the server.

```bash
export VAULTMIND_PASSPHRASE="your-very-strong-passphrase-here"
````

Use a **unique, long passphrase** (20+ characters recommended).  
Never commit it to git or share it.

---

### Start the server (HTTP transport – recommended)

## Demo
<p align="center">
  <img src="assets/gif/vaultmin-mcp.gif" width="800"/>
</p>

```bash
./vaultmind
```

Expected output:

```
VaultMind starting with passphrase from VAULTMIND_PASSPHRASE env var
VaultMind HTTP server listening at: http://127.0.0.1:49152
Press Ctrl+C to stop
```

Keep this terminal open — the server must remain running.

Note the port (e.g. `49152`) — you'll use this URL to connect clients.

The server binds only to `127.0.0.1` (localhost) for security.

**Tip:**  
If you want a fixed port (e.g. always `8765`), change:

```
127.0.0.1:0
```

to:

```
127.0.0.1:8765
```

in `main.go` and rebuild.

---

# Connecting Your AI Client

## Claude Desktop / Cursor / Continue.dev

1. Open the client's **MCP server settings**
2. Add a new server
3. Select **HTTP transport**
4. Enter the URL from the server log, for example:

```
http://127.0.0.1:49152
```

5. Save and reload/restart the client if required.

---

## GitHub Copilot in VS Code

1. Open **VS Code**
2. Press:

```
Ctrl+Shift+P
```

(macOS: `Cmd+Shift+P`)

3. Type:

```
MCP: Open User Configuration
```

4. Select it.

This opens (or creates) **`mcp.json`**.

Add your server:

```json
{
  "servers": {
    "vaultmind": {
      "name": "VaultMind",
      "type": "http",
      "url": "http://127.0.0.1:49152"
    }
  }
}
```

5. Save the file.

6. Reload the VS Code window:

```
Ctrl+R
```

(macOS: `Cmd+R`)

---

# Testing the Integration

Open a chat or agent window in your connected client and try these example prompts.

### Generate a password

```
Use VaultMind to generate a 24-character password with symbols
```

### Add a credential

```
VaultMind, add a credential:
service: GitHub
type: api_key
secret: ghp_test1234567890abcdef
notes: Personal access token - rotate soon
```

### Retrieve a credential

```
Use VaultMind to get credential [paste the returned ID here]
```

### Delete a credential

```
Use VaultMind to delete credential [ID] and confirm it
```

### Expected behavior

- The AI shows **thinking steps → tool call → response from VaultMind**
- No connection or authentication errors (if the server is running)

---

# Security & Best Practices

## Critical warnings

### Local-only

Do **not expose the HTTP port remotely** without strong authentication  
(mTLS, API keys, VPN, etc.).

### Master passphrase

Must be **strong (20+ characters, unique)**.

A weak passphrase means the entire database can be decrypted if the file is stolen.

### Seed phrases

Extremely high risk even when encrypted.

Always maintain **secure offline backups**  
(metal or paper in a safe location).

VaultMind is for **convenience only**.

### Not production-audited

Treat as **experimental for critical or high-value secrets**.

### Backups

Periodically copy the directory:

```
./vaultmind.pebble
```

The database is encrypted.

Restoring requires **the exact same passphrase**.

---

# Project Structure

```
cmd/
  vaultmind/         # entry point & main logic

internal/
  server/            # environment variable & config loading
  crypto/            # encryption, decryption, key derivation
  mtools/            # Tool handlers
  storage/           # Pebble database operations & Credential model
```

---

# Contributing

1. Fork the repository and create a pull request
2. Add tests

```bash
go test ./...
```

3. Follow Go best practices

- formatting
- error handling
- documentation

### Ideas welcome

- `listCredentials` (metadata-only overview)
- `bulkAddCredentials` (faster multi-add)
- Audit logging for access & changes
- Expiry / rotation reminders
- TOTP / 2FA support
- Backup / restore tools

---

# License

MIT License

See **LICENSE** for full text.
