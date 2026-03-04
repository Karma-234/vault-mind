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
