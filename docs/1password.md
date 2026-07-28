# 1Password CLI Access

## Decision

Use `op-agent` for non-interactive agent access, regardless of whether the
caller is Codex, Claude Code, Pi, or another agent. Use plain `op` for personal
account access that can require desktop approval, biometrics, SSO, or MFA.

`bin/op-agent` is a transparent launcher for the 1Password CLI. On macOS it
loads a service-account token from Keychain and exports it only to the `op`
child process. On other systems it accepts a token injected into that process
by the host's secret manager. It does not store or print the token.

Keep authentication local to each machine. Do not commit tokens or add them to
chezmoi, shell startup files, agent instructions, command arguments, or logs.

## Access Model

Create a dedicated 1Password service account for unattended agent work. Grant
it access only to the required vaults and permissions:

- `read_items` for reading and injecting existing secrets.
- `write_items` only where agents are allowed to run `op-agent item create` or
  update items.

Prefer a separate service account per host so one machine can be audited,
rotated, or revoked without affecting the others. Personal vault access stays
behind plain `op` and interactive user authentication. Service accounts cannot
access built-in Personal, Private, or Employee vaults, so agent-accessible
items must live in a separately granted vault.

## macOS Host Setup

1. Install the 1Password CLI:

   ```bash
   brew install 1password-cli
   ```

2. Clone this repository and ensure its `bin/` directory is on `PATH`. Confirm
   that both commands resolve:

   ```bash
   command -v op
   command -v op-agent
   ```

3. Create a service account in 1Password for this host and grant the minimum
   vault permissions above. The creation flow shows its token only once; save
   the recovery copy in an appropriately restricted 1Password vault, then copy
   the token for the next step.

4. Store the token in macOS Keychain. This form prompts for the token instead
   of placing it in the command line or shell history:

   ```bash
   /usr/bin/security add-generic-password \
     -a 'service-account-token' \
     -s 'com.sjunepark.op-agent' \
     -U \
     -w
   ```

5. Verify identity and visible vaults without reading secret values:

   ```bash
   op-agent whoami
   op-agent vault list
   ```

The first Keychain read may require local approval. Do not weaken the Keychain
item's access controls merely to suppress that approval.

## Personal Account Setup

The `op-agent` launcher keeps its service-account token scoped to its child
process. For personal access, ensure `OP_SERVICE_ACCOUNT_TOKEN` is not already
set, enable CLI integration in the 1Password desktop app's Developer settings,
then verify access. Install and sign in to the desktop app first if it is not
already present:

```bash
op signin
op whoami
```

Use plain `op` only when the task requires personal-account access or explicit
user authorization. Do not copy personal session tokens between machines.

## Other Operating Systems

`op-agent` uses macOS Keychain when no token is already present. On another
operating system, use that host's secret manager to inject
`OP_SERVICE_ACCOUNT_TOKEN` into the `op-agent` process. Do not persist it in a
shell profile or repository file.

For example, the secret manager should perform the equivalent of:

```bash
OP_SERVICE_ACCOUNT_TOKEN='<injected at runtime>' op-agent whoami
```

The placeholder is explanatory only. Do not type a real token directly into a
command, because it can be retained in shell history or process metadata.

## Rotation And Removal

When rotating access, create or obtain the replacement token and rerun the
Keychain command from the macOS setup. Verify `op-agent whoami`, then revoke the
old service account or token in 1Password.

To deprovision a macOS host, revoke its service account first, then remove the
local Keychain item:

```bash
/usr/bin/security delete-generic-password \
  -a 'service-account-token' \
  -s 'com.sjunepark.op-agent'
```

## `op-codex` Migration

`op-agent` replaces the machine-local `op-codex` name because the credential
boundary is shared by multiple agent harnesses. During migration, the launcher
also recognizes the former Keychain identifiers so existing access continues
to work:

```text
account: codex-remote-host
service: com.openai.codex.1password-service-account
```

Provision the generic Keychain item using the macOS setup above and verify it
before deleting the old item or local `~/.local/bin/op-codex` wrapper. The
legacy lookup can be removed from `bin/op-agent` after all hosts migrate.

For current CLI installation, authentication, and service-account behavior,
refer to the official [1Password CLI documentation](https://www.1password.dev/cli/get-started/).
