# http-server example

A 60-line `net/http` server that proxies the two most common MACHINA wallet
operations: list wallets and send a transaction. Use it as a reference for
wiring the Go SDK into any existing HTTP backend.

## Run

```
export MACHINA_APP_ID=app_123
export MACHINA_APP_SECRET=...
go run .
```

## Endpoints

```
GET  /list-wallets?owner_user_id=u_1&chain=ethereum
POST /send                {"wallet_id":"wal_x","to":"0xabc","amount_atoms":"1000","chain":"ethereum"}
```

## Error mapping

| SDK error | HTTP response |
| --- | --- |
| `PolicyDeniedError` | 403 with `policy_id` and `policy_name` |
| `ApprovalRequiredError` | 202 with `approval_id` |
| `ErrNotFound` | 404 |
| any other | 502 |
