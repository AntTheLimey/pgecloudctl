# pgecloudctl — Roadmap

Items ranked by weighted score. Higher = do first.

**Score = (Impact × 2 + Urgency) / Effort**

### Impact

| Score | Meaning |
|-------|---------|
| 5 | Closes competitive gap or unlocks new user segment |
| 4 | Significant UX improvement for existing users |
| 3 | Useful but not blocking adoption |
| 2 | Nice-to-have, quality of life |
| 1 | Cosmetic or speculative |

### Urgency

| Score | Meaning |
|-------|---------|
| 5 | Blocks other work or missing from every competitor |
| 4 | Needed for first public release |
| 3 | Expected before recommending to customers |
| 2 | Can ship without it |
| 1 | No time pressure |

### Effort

| Size | Hours | Score |
|------|-------|-------|
| S | 1–4 | 1 |
| M | 4–16 | 2 |
| L | 16–40 | 3 |
| XL | 40+ | 4 |

---

| Item | Impact | Urgency | Effort | Score | Status | Notes |
|------|--------|---------|--------|-------|--------|-------|
| Auth (login/status/logout) | 5 | 5 | S (1) | 15.0 | Done | v0.1 — foundation for all commands |
| Clusters CRUD | 5 | 5 | M (2) | 7.5 | Done | v0.1 — core resource |
| Databases CRUD | 5 | 5 | M (2) | 7.5 | Done | v0.1 — core resource |
| Tasks list/get/wait | 5 | 5 | M (2) | 7.5 | Done | v0.1 — async bridge for AI agents |
| OpenAPI client generation | 4 | 5 | M (2) | 6.5 | Done | v0.1 — foundation for all API calls |
| Cloud accounts CRUD | 4 | 4 | M (2) | 6.0 | Done | v0.1 — needed to create clusters |
| MCP server deploy/update | 4 | 4 | M (2) | 6.0 | Done | v0.1 — key differentiator |
| RAG server deploy/update | 4 | 4 | M (2) | 6.0 | Done | v0.1 — key differentiator |
| Database services list/get/remove | 4 | 4 | S (1) | 12.0 | Done | v0.1 — service management |
| Table + JSON output | 5 | 5 | S (1) | 15.0 | Done | v0.1 — AI agents need JSON |
| CI + linting + goreleaser | 4 | 4 | M (2) | 6.0 | Done | v0.1 — pgEdge repo standards |
| Backups + backup stores | 3 | 3 | M (2) | 4.5 | Done | v0.2 |
| SSH keys | 3 | 2 | S (1) | 8.0 | Done | v0.2 |
| Ingresses + service registration | 4 | 3 | M (2) | 5.5 | Done | v0.2 — enterprise service exposure |
| Invites + memberships | 3 | 2 | M (2) | 4.0 | Done | v0.2 |
| Shares + CloudFormation template | 2 | 2 | S (1) | 6.0 | Done | v0.2 |
| YAML output format | 2 | 2 | S (1) | 6.0 | Done | v0.2 |
| --verbose HTTP tracing | 3 | 2 | S (1) | 8.0 | Done | v0.2 — debugging |
| Install script | 2 | 2 | S (1) | 6.0 | Done | v0.2 |
| UUID prefix matching | 3 | 2 | S (1) | 8.0 | Idea | v0.2 — accept short IDs like Docker does |
| Command-level tests | 3 | 3 | M (2) | 4.5 | Done | v0.2 — checkResponse, buildServiceList, wait loop |
| --no-color wiring + color output | 2 | 2 | S (1) | 6.0 | Done | v0.2 — flags declared but not yet functional |
| llms.txt | 4 | 3 | M (2) | 5.5 | Done | v0.3 — AI discoverability |
| Claude Code skill | 4 | 3 | M (2) | 5.5 | Done | v0.3 — Claude Code integration |
| AI workflow recipes | 3 | 2 | M (2) | 4.0 | Done | v0.3 — docs/guides for AI agents |
| pgecloudctl doctor | 3 | 2 | S (1) | 8.0 | Done | v0.3 — AI self-diagnostics |
| Multi-tenancy support | 4 | 3 | L (3) | 3.7 | Idea | Blocked on API changes |
| PostgREST service commands | 3 | 2 | S (1) | 8.0 | Idea | Blocked on API availability |
| MCP server for CLI | 3 | 2 | L (3) | 2.7 | Idea | Revisit if market demands it |

---

### Completed

**v0.1**

- Auth (login/status/logout)
- Clusters CRUD
- Databases CRUD
- Tasks list/get/wait
- OpenAPI client generation
- Cloud accounts CRUD
- MCP server deploy/update
- RAG server deploy/update
- Database services list/get/remove
- Table + JSON output
- CI + linting + goreleaser

**v0.2**

- Backups + backup stores
- SSH keys
- Ingresses + service registration
- Invites + memberships
- Shares + CloudFormation template
- YAML output format
- --verbose HTTP tracing
- Install script
- --no-color wiring + color output
- Command-level tests

**v0.3**

- llms.txt
- Claude Code skill
- AI workflow recipes
- pgecloudctl doctor

---

Post-v0.3 ideas tracked in the Idea rows above.
Multi-tenancy and PostgREST blocked on upstream API work.
