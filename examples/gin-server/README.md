# gin-server example

A Gin server that integrates the MACHINA wallet SDK and includes a webhook
receiver that performs HMAC signature verification on every incoming event.

## Run

```
export MACHINA_APP_ID=app_123
export MACHINA_APP_SECRET=...
export MACHINA_WEBHOOK_SECRET=whsec_...
go run .
```

## Endpoints

```
GET  /wallets/:id
POST /wallets/:id/send         {"to":"0xabc","amount_atoms":"1000","chain":"ethereum"}
POST /webhooks/machina         # HMAC-verified CloudEvents receiver
```

## Webhook verification

The receiver:

1. Reads the raw body (do not parse before verifying — the signature covers the
   exact bytes that were transmitted).
2. Calls `machinawallet.VerifySignature(body, header, secret, tolerance)`.
3. Parses the body into a `CloudEvent` and dispatches on `ev.Type`.

The signature header is Stripe-compatible: `t=<unix>,v1=<hex>`.
