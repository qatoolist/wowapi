# Minimal container for GitHub git operations (clone/fetch/pull/push) as a
# specific identity, using ONLY an explicitly-mounted SSH key — no ssh-agent, no
# 1Password, so it can never pick up the wrong account's key. Signing/commits are
# left to the host; this image does network git ops.
FROM alpine:3.20

RUN apk add --no-cache git openssh-client ca-certificates

COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

WORKDIR /repo
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["git", "status"]
