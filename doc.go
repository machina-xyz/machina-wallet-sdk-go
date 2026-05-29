// Package machinawallet is the MACHINA wallet SDK for Go.
//
// It provides a server-side Go client for the MACHINA wallet management API,
// covering wallet lifecycle, balances, transactions, spend policies, real-time
// streaming, and webhook signature verification.
//
// The SDK is idiomatic Go:
//
//   - All RPC-style methods accept a context.Context as the first parameter and
//     return a (value, error) tuple.
//   - The core package has zero external dependencies — it only uses the Go
//     standard library.
//   - Streaming APIs return Go channels so callers can range over events and
//     stop via context cancellation.
//   - Errors are typed and unwrappable via errors.Is and errors.As.
//
// Quickstart:
//
//	client, err := machinawallet.NewClient(machinawallet.Config{
//	    BaseURL:   "https://api.machina.money",
//	    AppID:     "app_123",
//	    AppSecret: os.Getenv("MACHINA_APP_SECRET"),
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	wallets, err := client.ListWallets(ctx, nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// See the examples directory for end-to-end integrations including a plain
// net/http server, a Gin server with webhook receiver, and a gRPC sidecar
// pattern for Kubernetes-native deployments.
package machinawallet
