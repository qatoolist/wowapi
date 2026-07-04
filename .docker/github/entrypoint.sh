#!/bin/sh
# Configure a hermetic git identity + auth for GitHub, then exec the command.
# Auth uses ONLY the key mounted at /keys/id (IdentitiesOnly), so this container
# always acts as the intended account regardless of any host agent/1Password.
set -e

# The repo is bind-mounted from the host (different uid) — trust it.
git config --global --add safe.directory /repo
git config --global --add safe.directory '*'

# Identity (overridable via env).
git config --global user.name  "${GIT_USER_NAME:-qatoolist}"
git config --global user.email "${GIT_USER_EMAIL:-qatoolist@gmail.com}"

# Force the mounted key for every SSH connection; no agent, no other identities.
mkdir -p /root/.ssh
git config --global core.sshCommand \
  "ssh -i /keys/id -o IdentitiesOnly=yes -o StrictHostKeyChecking=accept-new -o UserKnownHostsFile=/root/.ssh/known_hosts"

# The host repo may use the 'github_personal' SSH host alias (defined only in the
# host's ~/.ssh/config). Inside here that alias does not exist, so rewrite it to
# the real host — the forced key above still selects the right account.
git config --global url."git@github.com:".insteadOf "git@github_personal:"

exec "$@"
