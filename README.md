# machina-wallet-sdk-go

MACHINA wallet SDK for Go — server-side integration.

The Go SDK is the third member of the MACHINA wallet SDK family, alongside the
TypeScript and Python SDKs. It targets Kubernetes operators, cloud
infrastructure, and crypto infra workloads where Go is the lingua franca and
the alternatives are absent: no other production-grade agent-wallet platform
ships a first-class Go SDK. This SDK exists to close that gap.

- Module: `github.com/machina-xyz/machina-wallet-sdk-go`
- License: MIT
- Status: phase 1 scaffold

## Why Go

Go is unavoidable in three of MACHINA's target customer segments:

- Kubernetes operators and cloud infrastructure
- Crypto infra (validators, RPC providers, MEV operators, indexers)
- Enterprise fintech backends where performance and tail-latency rule

These customers will not adopt a wallet provider that forces them off Go.

## Design

- **Idiomatic Go.** Every RPC takes `context.Context` first and returns
  `(value, error)`. No callbacks, no panics, no clever wrappers.
- **Zero-dep core.** The wallet client is implemented on `net/http`,
  `crypto/ed25519`, `crypto/hmac`, `crypto/sha256`, `encoding/json`, and
  `log/slog`. Examples may pull in third-party libraries (Gin, gRPC).
- **Channels for streams.** Transaction streams are `<-chan Transaction` with
  a sibling `<-chan error`. Cancellation flows through `context.Context`.
- **Typed errors.** `errors.Is(err, ErrPolicyDenied)` and
  `errors.As(err, &PolicyDeniedError{})` both work as expected.
- **Signed requests.** Ed25519 (recommended) or HMAC-SHA256. Mode is
  auto-detected from the secret format and matches the TypeScript and Python
  SDKs byte-for-byte.

## 30-second quickstart

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    machinawallet "github.com/machina-xyz/machina-wallet-sdk-go"
)

func main() {
    client, err := machinawallet.NewClient(machinawallet.Config{
        AppID:     "app_123",
        AppSecret: os.Getenv("MACHINA_APP_SECRET"),
    })
    if err != nil {
        log.Fatal(err)
    }

    page, err := client.ListWallets(context.Background(), nil)
    if err != nil {
        log.Fatal(err)
    }
    for _, w := range page.Items {
        fmt.Println(w.ID, w.Chain, w.Address)
    }
}
```

## Webhook integration in 60 seconds

```go
http.HandleFunc("/webhooks/machina", func(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)
    if err := machinawallet.VerifySignature(body, r.Header.Get("Machina-Signature"), webhookSecret, 5*time.Minute); err != nil {
        http.Error(w, err.Error(), http.StatusUnauthorized)
        return
    }
    ev, err := machinawallet.ParseEvent(body)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    log.Printf("event %s type=%s", ev.ID, ev.Type)
    w.WriteHeader(http.StatusNoContent)
})
```

## gRPC sidecar pattern (Kubernetes-native)

The `examples/grpc-microservice/` example demonstrates the deployment shape
crypto infra teams reach for first: a sidecar container owns the app secret
and exposes a typed `WalletGateway` over loopback gRPC. Every other container
in the Pod calls the gateway instead of talking to the wallet management API
directly. This consolidates signing, retries, observability, and approval
handling into one place per Pod — without any container needing access to the
app secret.

```
Pod
+---------------------------+      +-------------------------+
| your service (any lang)   | ---> | wallet-gateway sidecar  | --> MACHINA
+---------------------------+      | (this Go SDK)           |
                                   +-------------------------+
```

## Streaming transactions

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

events, errs, err := client.StreamTransactions(ctx, walletID)
if err != nil {
    log.Fatal(err)
}
for {
    select {
    case tx, ok := <-events:
        if !ok {
            return
        }
        handle(tx)
    case err := <-errs:
        log.Println("stream error:", err)
        return
    case <-ctx.Done():
        return
    }
}
```

## Error handling

```go
_, err := client.SubmitTransaction(ctx, walletID, intent)
switch {
case errors.Is(err, machinawallet.ErrPolicyDenied):
    var pde *machinawallet.PolicyDeniedError
    errors.As(err, &pde)
    log.Printf("blocked by policy %s (%s)", pde.PolicyID, pde.PolicyName)

case errors.Is(err, machinawallet.ErrApprovalRequired):
    var ar *machinawallet.ApprovalRequiredError
    errors.As(err, &ar)
    log.Printf("queued for approval: %s", ar.ApprovalID)

case errors.Is(err, machinawallet.ErrNotFound):
    return http.StatusNotFound

case err != nil:
    return http.StatusBadGateway
}
```

## Feature matrix

| Capability | Status |
| --- | --- |
| List, get, create wallets | yes |
| Balances | yes |
| Submit transaction (custodial) | yes |
| Prepare and broadcast (self-custody) | yes |
| Spend policies (CRUD + typed builders) | yes |
| Policy check (dry-run) | yes |
| SSE transaction stream | yes |
| Webhook signature verification | yes |
| Ed25519 and HMAC-SHA256 signing | yes |
| Retry with exponential backoff and `Retry-After` | yes |
| Context-based cancellation everywhere | yes |
| Zero external dependencies in core | yes |
| gRPC sidecar example | yes |

## Companion SDKs

- TypeScript and React: `machina-xyz/machina-wallet-sdk-ts`
- Python (sync, async, LangChain): `machina-xyz/machina-wallet-sdk-python`

## License

MIT. See [LICENSE](LICENSE).
