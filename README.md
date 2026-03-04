# VaultMind

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Stars](https://img.shields.io/github/stars/yourusername/vaultmind-mcp?style=social)](https://github.com/yourusername/vaultmind-mcp)

**VaultMind** is a **local-first, privacy-focused MCP (Model Context Protocol) server** written in Go that lets your AI agents (Claude Desktop, Cursor, Continue.dev, etc.) securely manage credentials — passwords, API keys, seed phrases, tokens — **without ever exposing them to the cloud or LLM provider**.

All data is encrypted at rest (AES-256-GCM + Argon2id-derived keys), stored in an embedded BadgerDB, and decrypted only ephemerally in memory when you explicitly request it. Secrets never enter your chat history or leave your machine.

### Why VaultMind?

Pasting real API keys or seed phrases into Claude/GPT chats risks permanent exposure (provider logs, accidental context leaks, breaches). VaultMind solves this:

- **Zero transmission** — secrets stay local; AI only sees metadata/summaries.
- **AI-powered hygiene** — generate, add, list, retrieve, analyze expiry/strength without plaintext exposure.
- **High-risk warnings** — extra friction & alerts for seed phrases/crypto keys.
- **Offline-first** — no internet needed (optional local breach checks later).

Perfect for developers handling production tokens, crypto users, or anyone who wants AI assistance without trust trade-offs.

### Features

- Generate cryptographically secure passwords/passphrases
- Add & categorize credentials (API keys, passwords, seed phrases)
- List summaries safely (service, type, age, expiry)
- Retrieve with confirmation & ephemeral decryption
- Built-in warnings for high-risk items (e.g., seeds)
- Future: expiry alerts, weak password detection, audit logs

### Quick Start

1. **Prerequisites**
   - Go 1.22+
   - (Optional) BadgerDB-compatible filesystem (works on macOS/Linux/Windows)

2. **Install**
   ```bash
   git clone https://github.com/yourusername/vaultmind-mcp.git
   cd vaultmind-mcp
   go mod tidy
   go build -o vaultmind ./cmd/vaultmind
   ```
   **OR**
3. **Install directly**
   ```bash
   go install github.com/yourusername/vaultmind-mcp/cmd/vaultmind@latest
   ./vaultmind
   ```

Enter your master passphrase when prompted (used to derive encryption key).
Server starts on stdio (for local MCP clients).

Connect to Your AI ClientClaude Desktop / Cursor / Continue.dev:Add as stdio server (command: full path to ./vaultmind).
Example config in claude_desktop_config.json or client settings:json

{
"mcpServers": {
"vaultmind": {
"command": "/path/to/vaultmind",
"args": [],
"type": "stdio"
}
}
}

Prompt examples:"Use VaultMind to generate a 32-char passphrase."
"VaultMind, add my OpenAI API key: sk-..."
"VaultMind, list my API keys added this month."

Security Warnings & Best PracticesLocal-only by design — no remote hosting recommended for sensitive secrets (transit/memory risks).
Not production-audited — treat as experimental for critical items (e.g., main crypto seeds). Use offline backups (metal/paper) for seeds.
Master passphrase — choose a strong one (20+ chars); weak passphrase = whole DB at risk if file stolen.
Run with caution — MCP servers inherit your process privileges.
Backup — Export encrypted DB manually; restore requires same passphrase.

Example Usage in Claude/Cursortext

> Use VaultMind to add a credential:
> service: Stripe
> type: api*key
> secret: sk_test*...
> notes: Test mode key

VaultMind: Credential added securely (ID: abc123...). Warning: Rotate periodically.

text

> VaultMind, get my Stripe API key

VaultMind: Enter master passphrase in server console...
Retrieved: sk*test*... (copy now — will be wiped from memory)

Development & ContributingBuilt with:modelcontextprotocol/go-sdk — official MCP Go SDK
BadgerDB — embedded KV store
golang.org/x/crypto/argon2 — key derivation
golang.org/x/term — secure input

Folder structure follows standard Go layout:

cmd/vaultmind/ # main entry point
internal/
crypto/ # encryption utils
storage/ # DB & credential ops

To contribute:Fork & PR
Add tests (go test ./...)
Follow Go best practices (effective Go, clean architecture)

Ideas welcome: TOTP support, audit viewer, weak password scanner, etc.LicenseMIT License — see LICENSEMade with for privacy-conscious AI users.Star the repo if this solves a pain point for you! Feedback/PRs appreciated.
