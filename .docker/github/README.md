# Hermetic GitHub git container (`ghgit`)

Runs GitHub **network git operations** (clone / fetch / pull / push / ls-remote) as a fixed identity
inside a throwaway container that uses **only an explicitly-mounted SSH key** — no `ssh-agent`, no
1Password, no host keychain. This guarantees the operation authenticates as the intended account and can
never silently pick up a different account's key (the problem this solves: the host 1Password agent was
serving the `qualitycoe` key instead of `qatoolist`).

## What it does / doesn't

- **Does:** authenticate as `qatoolist` using `~/.ssh/qatoolist_personal` (mounted read-only), rewrite the
  `github_personal` host alias to `github.com`, and run any git network command against the repo.
- **Doesn't:** sign commits. Signing stays on the host (1Password `op-ssh-sign`). Make + sign commits on
  the host, then push them with `ghgit`. (In-container `commit` is forced unsigned so it can't fail on the
  missing 1Password signer.)

## Usage

```bash
# from anywhere inside the repo:
.docker/github/ghgit ls-remote origin        # read-only auth check
.docker/github/ghgit push -u origin main      # push as qatoolist
.docker/github/ghgit fetch --all
```

Put it on your PATH for convenience:

```bash
ln -s "$PWD/.docker/github/ghgit" ~/bin/ghgit   # then: ghgit push -u origin main
```

The image builds automatically on first run (`wowapi-ghgit:latest`).

## Configuration (env overrides)

| Var | Default | Purpose |
|---|---|---|
| `GHGIT_KEY` | `~/.ssh/qatoolist_personal` | private key to auth with (its `.pub` must sit beside it) |
| `GHGIT_IMAGE` | `wowapi-ghgit:latest` | image tag |
| `GIT_USER_NAME` | `qatoolist` | commit author name (for in-container commits) |
| `GIT_USER_EMAIL` | `qatoolist@gmail.com` | commit author email |

To use a different key (e.g. a dedicated deploy key you provide instead of the personal one):

```bash
GHGIT_KEY=~/.ssh/some_other_key .docker/github/ghgit push -u origin main
```

## How it works

- `Dockerfile` — alpine + git + openssh-client.
- `entrypoint.sh` — sets identity, `core.sshCommand = ssh -i /keys/id -o IdentitiesOnly=yes …` (forces the
  mounted key), and `url."git@github.com:".insteadOf "git@github_personal:"`.
- `ghgit` — host wrapper: builds the image if needed, bind-mounts the repo (`/repo`) and the key
  (`/keys/id[.pub]` read-only), and runs `git -c commit.gpgsign=false …`.

## Rebuild after editing the Dockerfile/entrypoint

```bash
docker build -t wowapi-ghgit:latest .docker/github
```
