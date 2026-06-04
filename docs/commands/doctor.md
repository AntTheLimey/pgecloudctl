# doctor

Run environment diagnostics to verify that `pgecloudctl` and its
environment are configured correctly.

`doctor` does not require authentication — it can be run when auth is
broken to help diagnose the problem.

## Usage

```
pgecloudctl doctor [flags]
```

## Flags

| Flag | Required | Description |
|------|----------|-------------|
| `-h, --help` | No | help for doctor |

## Checks

`doctor` runs 9 checks and prints a status table. Each check reports
`ok`, `warning`, or `error`.

| Check | What it verifies |
|-------|-----------------|
| **Version** | Reports the current `pgecloudctl` version, Go runtime version, OS, and architecture. Always `ok`. |
| **Latest version** | Queries the GitHub Releases API to compare the installed version against the latest published release. Reports `ok` if up to date, `warning` if a newer version is available or the check could not reach GitHub. |
| **Auth** | Resolves credentials from the config file or environment variables and checks whether the cached token exists and has not expired. Reports `ok` if authenticated with a valid token, `warning` if the token is expired, and `error` if no credentials are found. |
| **API connectivity** | Makes an HTTP GET to the configured API base URL and records the HTTP status code and round-trip latency in milliseconds. Reports `ok` if reachable, `error` if the host is unreachable. |
| **Config** | Checks whether `config.json` and `token.json` exist in the config directory (default `~/.config/pgecloudctl/`). Reports `ok` if `config.json` is present, `warning` if it is missing. |
| **Environment** | Reports the OS/arch and whether the `PGEDGE_CLIENT_ID`, `PGEDGE_API_URL`, and `NO_COLOR` environment variables are set. Always `ok` — informational only. |
| **Shell** | Detects the current shell from `$SHELL` and checks whether `pgecloudctl` is on `$PATH` via `which`. Reports `ok` if found in PATH, `warning` if not. |
| **Install method** | Infers how `pgecloudctl` was installed by inspecting the executable path. Reports `homebrew`, `install-script`, `go-install`, or `unknown`. Always `ok` — informational only. |
| **Skill** | Checks whether the pgecloudctl Claude Code skill is installed in `~/.claude/plugins/pgecloudctl/`. Reports `ok` with the skill version if found, `warning` if not installed. |

## Example

```bash
pgecloudctl doctor
```

**Example output (table):**

```
CHECK             STATUS    DETAILS
Version           ok        v0.3.0 (go1.22.3, darwin/arm64)
Latest version    ok        v0.3.0 (up to date)
Auth              ok        authenticated via config (expires 2026-12-31T23:59:59Z)
API connectivity  ok        https://api.pgedge.com (200, 87ms)
Config            ok        /Users/alice/.config/pgecloudctl (config.json: yes, token.json: yes)
Environment       ok        darwin/arm64, PGEDGE_CLIENT_ID: not set
Shell             ok        /bin/zsh, in PATH
Install method    ok        homebrew
Skill             ok        installed (v0.3.0)
```

## Output formats

Use `-o json` or `-o yaml` to get machine-readable output covering all
check fields, useful for CI diagnostics.

```bash
pgecloudctl doctor -o json
```
