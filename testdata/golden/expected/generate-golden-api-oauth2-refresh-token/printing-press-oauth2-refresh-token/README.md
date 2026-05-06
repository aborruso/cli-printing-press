# Printing Press Oauth2 Refresh Token CLI

Purpose-built fixture for the OAuth2 refresh_token auth shape.

## Install

### Binary

Download a pre-built binary for your platform from the [latest release](https://github.com/mvanhorn/printing-press-library/releases/tag/printing-press-oauth2-refresh-token-current). On macOS, clear the Gatekeeper quarantine: `xattr -d com.apple.quarantine <binary>`. On Unix, mark it executable: `chmod +x <binary>`.

### Go

```
go install github.com/mvanhorn/printing-press-library/library/other/printing-press-oauth2-refresh-token/cmd/printing-press-oauth2-refresh-token-pp-cli@latest
```

## Quick Start

### 1. Install

See [Install](#install) above.

### 2. Set Up Credentials

Self-authorize your private application in the provider console, export the OAuth client ID, OAuth client secret, and refresh token, then run doctor:

```bash
export TEST_API_LWA_CLIENT_ID="<client-id>"
export TEST_API_LWA_CLIENT_SECRET="<client-secret>"
export TEST_API_REFRESH_TOKEN="<refresh-token>"
printing-press-oauth2-refresh-token-pp-cli doctor
```

The CLI exchanges the refresh token for an access token on the first live request and caches the access token locally.

### 3. Verify Setup

```bash
printing-press-oauth2-refresh-token-pp-cli doctor
```

This checks your configuration and credentials.

### 4. Try Your First Command

```bash
printing-press-oauth2-refresh-token-pp-cli items
```

## Usage

Run `printing-press-oauth2-refresh-token-pp-cli --help` for the full command reference and flag list.

## Commands

### items

Manage items.

- **`printing-press-oauth2-refresh-token-pp-cli items list`** - List items.


## Output Formats

```bash
# Human-readable table (default in terminal, JSON when piped)
printing-press-oauth2-refresh-token-pp-cli items

# JSON for scripting and agents
printing-press-oauth2-refresh-token-pp-cli items --json

# Filter to specific fields
printing-press-oauth2-refresh-token-pp-cli items --json --select id,name,status

# Dry run — show the request without sending
printing-press-oauth2-refresh-token-pp-cli items --dry-run

# Agent mode — JSON + compact + no prompts in one flag
printing-press-oauth2-refresh-token-pp-cli items --agent
```

## Agent Usage

This CLI is designed for AI agent consumption:

- **Non-interactive** - never prompts, every input is a flag
- **Pipeable** - `--json` output to stdout, errors to stderr
- **Filterable** - `--select id,name` returns only fields you need
- **Previewable** - `--dry-run` shows the request without sending
- **Read-only by default** - this CLI does not create, update, delete, publish, send, or mutate remote resources
- **Offline-friendly** - sync/search commands can use the local SQLite store when available
- **Agent-safe by default** - no colors or formatting unless `--human-friendly` is set

Exit codes: `0` success, `2` usage error, `3` not found, `4` auth error, `5` API error, `7` rate limited, `10` config error.

## Use with Claude Code

Install the focused skill — it auto-installs the CLI on first invocation:

```bash
npx skills add mvanhorn/printing-press-library/cli-skills/pp-printing-press-oauth2-refresh-token -g
```

Then invoke `/pp-printing-press-oauth2-refresh-token <query>` in Claude Code. The skill is the most efficient path — Claude Code drives the CLI directly without an MCP server in the middle.

<details>
<summary>Use as an MCP server in Claude Code (advanced)</summary>

If you'd rather register this CLI as an MCP server in Claude Code, install the MCP binary first:

```bash
go install github.com/mvanhorn/printing-press-library/library/other/printing-press-oauth2-refresh-token/cmd/printing-press-oauth2-refresh-token-pp-mcp@latest
```

Then register it:

```bash
# Self-authorize your private application, then provide the OAuth setup env vars:
claude mcp add printing-press-oauth2-refresh-token printing-press-oauth2-refresh-token-pp-mcp -e TEST_API_LWA_CLIENT_ID=<client-id> -e TEST_API_LWA_CLIENT_SECRET=<client-secret> -e TEST_API_REFRESH_TOKEN=<refresh-token>
```

</details>

## Use with Claude Desktop

This CLI ships an [MCPB](https://github.com/modelcontextprotocol/mcpb) bundle — Claude Desktop's standard format for one-click MCP extension installs (no JSON config required).

The bundle uses OAuth setup environment variables from your MCP host. Self-authorize your private application, provide the OAuth client ID, OAuth client secret, and refresh token, then run doctor.

To install:

1. Download the `.mcpb` for your platform from the [latest release](https://github.com/mvanhorn/printing-press-library/releases/tag/printing-press-oauth2-refresh-token-current).
2. Double-click the `.mcpb` file. Claude Desktop opens and walks you through the install.
3. Fill in `TEST_API_LWA_CLIENT_ID`, `TEST_API_LWA_CLIENT_SECRET`, and `TEST_API_REFRESH_TOKEN` when Claude Desktop prompts you.

Requires Claude Desktop 1.0.0 or later. Pre-built bundles ship for macOS Apple Silicon (`darwin-arm64`) and Windows (`amd64`, `arm64`); for other platforms, use the manual config below.

<details>
<summary>Manual JSON config (advanced)</summary>

If you can't use the MCPB bundle (older Claude Desktop, unsupported platform), install the MCP binary and configure it manually.

```bash
go install github.com/mvanhorn/printing-press-library/library/other/printing-press-oauth2-refresh-token/cmd/printing-press-oauth2-refresh-token-pp-mcp@latest
```

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "printing-press-oauth2-refresh-token": {
      "command": "printing-press-oauth2-refresh-token-pp-mcp",
      "env": {
        "TEST_API_LWA_CLIENT_ID": "<client-id>",
        "TEST_API_LWA_CLIENT_SECRET": "<client-secret>",
        "TEST_API_REFRESH_TOKEN": "<refresh-token>"
      }
    }
  }
}
```

</details>

## Health Check

```bash
printing-press-oauth2-refresh-token-pp-cli doctor
```

Verifies configuration, credentials, and connectivity to the API.

## Configuration

Config file: `~/.config/printing-press-oauth2-refresh-token-pp-cli/config.toml`

Environment variables:

| Name | Kind | Required | Description |
| --- | --- | --- | --- |
| `TEST_API_LWA_CLIENT_ID` | auth_flow_input | Yes | Set during initial auth setup. |
| `TEST_API_LWA_CLIENT_SECRET` | auth_flow_input | Yes | Set during initial auth setup. |
| `TEST_API_REFRESH_TOKEN` | auth_flow_input | Yes | Set during initial auth setup. |

## Troubleshooting
**Authentication errors (exit code 4)**
- Run `printing-press-oauth2-refresh-token-pp-cli doctor` to check credentials
**Not found errors (exit code 3)**
- Check the resource ID is correct
- Run the `list` command to see available items

---

Generated by [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press)
