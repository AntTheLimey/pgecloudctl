# Team Onboarding

Invite a new member to the pgEdge Cloud organization and confirm membership.

## Prerequisites

- Authenticated session with organization admin permissions
- The new member's email address

## Decision Points

- **Custom expiration** — invites default to a standard TTL; pass
  `--expiration <hours>` to set a non-default window (e.g. `72` for three days)
- **Member accepts via CLI vs UI** — the invite URL works in a browser; Step 3
  covers CLI acceptance for automated or scripted onboarding

## Steps

### Step 1: Create the invite

```bash
pgecloudctl invites create --email user@example.com -o json
```

To set a custom expiration (in hours):

```bash
pgecloudctl invites create --email user@example.com --expiration 72 -o json
```

Expected output: a JSON object containing `id` and a URL or `token` field.
Capture:

- `id` as `<invite-id>`
- `token` as `<token>` (used by the recipient if accepting via CLI)
- The invite URL for sharing via email or Slack

### Step 2: Share the invite URL

Send the invite URL from Step 1 to the new member. The URL is valid until the
expiration window passes.

If the member will accept via the pgEdge Cloud UI, no further CLI steps are
needed on their behalf — skip to Step 4 to confirm membership after they
accept.

### Step 3: Member accepts via CLI (optional)

If the new member is onboarding programmatically:

```bash
pgecloudctl invites accept <invite-id> --token <token>
```

Expected output: confirmation message or JSON with `"status": "accepted"`.

### Step 4: Verify membership

```bash
pgecloudctl memberships list
```

Expected output: the new member's email appears in the membership list with an
`active` status.

## Verification

- `pgecloudctl memberships list` — new member present with `"status": "active"`

## Error Handling

| Exit Code | Meaning | Recovery |
|-----------|---------|----------|
| 0 | Success | Continue to next step |
| 1 | General error (validation, expired invite, token issue) | Check command output for details; re-create invite if needed |
| 2 | Authentication/authorization failure | Confirm the account has org admin role; run `auth login` if needed |
| 3 | Request timeout | Retry the command; check network connectivity |
| 4 | Invite not found | Verify invite ID with `pgecloudctl invites list` |
