# grpc-microservice example

A gRPC microservice that wraps MACHINA wallet operations behind a small
`WalletGateway` proto API. This is the **Kubernetes-native sidecar pattern**
that crypto infra teams will recognize: a single co-located process owns the
app secret, centralizes retries and observability, and exposes a typed local
interface to every other container in the Pod.

## Why this pattern

- A single point that owns the MACHINA app secret — no other Pod container
  ever sees the credential.
- A loopback gRPC hop replaces signed-HTTP overhead in chatty workloads.
- The gateway is the natural place for tracing, rate limiting, and approval-
  required handling shared across services.
- Trivial to deploy as a daemonset or a sidecar container.

## Layout

```
.
├── go.mod
├── main.go
├── proto/
│   └── wallet_gateway.proto    # source of truth
└── pb/                         # generated stubs (committed)
    ├── wallet_gateway.pb.go
    └── wallet_gateway_grpc.pb.go
```

## Run

```
export MACHINA_APP_ID=app_123
export MACHINA_APP_SECRET=...
go run .
```

Then call from any client:

```
grpcurl -plaintext -d '{"wallet_id":"wal_x"}' localhost:50051 \
    machina.wallet.gateway.v1.WalletGateway/GetWallet
```

## Regenerating stubs

```
protoc --go_out=pb --go_opt=paths=source_relative \
       --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
       -I proto wallet_gateway.proto
```
