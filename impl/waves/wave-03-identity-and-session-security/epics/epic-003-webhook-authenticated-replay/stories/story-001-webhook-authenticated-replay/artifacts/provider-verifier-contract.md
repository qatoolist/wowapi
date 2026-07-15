---
id: ART-W03-E03-S001-005
type: contract-document
parent_story: W03-E03-S001
status: produced
created_at: 2026-07-13
updated_at: 2026-07-13
---

# Provider-verifier contract

This document defines the contract every `kernel/webhook.Verifier` implementation must satisfy. It
applies to all provider-specific verifiers registered with `Service.RegisterVerifier`.

## Core guarantee: authenticated fields only

`Verifier.Verify` returns an `Envelope`:

```go
type Envelope struct {
    CanonicalBody    []byte
    EventID          string
    OccurredAt       time.Time
    SignatureVersion string
    KeyID            string
}
```

Every field in `Envelope` MUST be derived from data covered by the provider's signature. No
caller-supplied request field — such as `InboundIn.Timestamp`, `InboundIn.ExternalEventID`, or any
unauthenticated header — may influence an `Envelope` field. If a field cannot be authenticated by
the signature scheme, it must be synthesized from authenticated data (e.g., a hash of the canonical
body) or from local, trusted state (e.g., the server's receipt time), or left empty.

## Failure contract

`Verify` returns `(Envelope, error)`. On failure the returned `error` MUST be non-nil and the
`Envelope` is undefined. Callers MUST check the error first and MUST NOT read any `Envelope` field
when `error != nil`. Implementations SHOULD return a zero-valued `Envelope` on failure.

## Field semantics

| Field | Meaning | Source rule |
|---|---|---|
| `CanonicalBody` | The exact byte sequence the signature authenticated. | MUST be the authenticated byte sequence. For body-only HMAC this is the raw request body; for timestamped schemes it may be a canonicalized string. |
| `EventID` | Stable identifier used for replay/dedup. | MUST be derived from authenticated data. If the provider scheme authenticates an event id, use it; otherwise synthesize a stable id from `CanonicalBody`. |
| `OccurredAt` | Timestamp used for the replay window check. | MUST be a timestamp the signature covers, or a locally-trusted receipt time when the scheme does not authenticate a timestamp. |
| `SignatureVersion` | Scheme version (e.g., `sha256`). | From authenticated data if the scheme versions signatures; otherwise may be empty. |
| `KeyID` | Identifier of the key used to verify. | From authenticated data if the scheme authenticates a key id; otherwise empty. |

## Security consequences

`HandleInbound` sources every replay-window and dedup decision exclusively from `Envelope`. A
verifier that surfaces an unauthenticated field in `Envelope` reintroduces the successful-signature
replay/dedup gap: an attacker with a validly signed body can manipulate the unauthenticated field
to bypass replay protection or occupy another event's dedup slot.

## Reference example: `HMACVerifier`

`kernel/webhook.HMACVerifier` implements the contract for the common external-provider scheme:
HMAC-SHA256 over the raw request body.

```go
func (v HMACVerifier) Verify(secret string, body []byte, headers map[string]string) (Envelope, error) {
    // 1. Locate and validate the signature header.
    got, err := v.extractSignature(headers)
    if err != nil {
        return Envelope{}, err
    }

    // 2. Recompute the HMAC over the body (the only authenticated datum).
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(body)
    want := hex.EncodeToString(mac.Sum(nil))
    if !hmac.Equal([]byte(got), []byte(want)) {
        return Envelope{}, signatureMismatchError()
    }

    // 3. Populate Envelope from authenticated or trusted local data only.
    sum := sha256.Sum256(body)
    return Envelope{
        CanonicalBody:    body,
        EventID:          "sha256:" + hex.EncodeToString(sum[:]),
        OccurredAt:       time.Now(), // receipt time: scheme does not authenticate a timestamp
        SignatureVersion: "sha256",
        KeyID:            "", // scheme does not authenticate a key id
    }, nil
}
```

## Limitation example: timestamped-provider protocols

`HMACVerifier`'s `OccurredAt` is the local receipt time because the body-only signature does not
cover a provider-asserted timestamp. This makes it unsuitable for providers whose security model
requires the provider's own timestamp (e.g., a provider that signs `"<timestamp>.<body>"`). Such
providers need a dedicated `Verifier` that extracts the authenticated timestamp from the signed
data and returns it in `Envelope.OccurredAt`.

## Implementation checklist

When adding a new `Verifier`:

1. Identify every field the provider's signature actually covers.
2. Derive each `Envelope` field only from those covered fields or from trusted local state.
3. Return a non-nil error on any signature failure; do not populate a partial `Envelope`.
4. Document any provider-specific limitation (e.g., receipt-time `OccurredAt`) in the verifier's
   godoc and, if security-relevant, in this contract.
5. Add unit tests proving each `Envelope` field is independent of caller-supplied, unauthenticated
   request fields.
