// Package main runs a minimal net/http server that proxies MACHINA wallet
// reads and a send endpoint via the Go SDK.
//
// Run with:
//
//	MACHINA_APP_ID=app_123 MACHINA_APP_SECRET=... go run .
package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	machinawallet "github.com/machina-xyz/machina-wallet-sdk-go"
)

func main() {
	cfg := machinawallet.Config{
		BaseURL:   envOr("MACHINA_BASE_URL", machinawallet.DefaultBaseURL),
		AppID:     mustEnv("MACHINA_APP_ID"),
		AppSecret: mustEnv("MACHINA_APP_SECRET"),
		Timeout:   10 * time.Second,
		Logger:    slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}
	client, err := machinawallet.NewClient(cfg)
	if err != nil {
		log.Fatalf("init client: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/list-wallets", listWalletsHandler(client))
	mux.HandleFunc("/send", sendHandler(client))

	addr := envOr("ADDR", ":8080")
	log.Printf("listening on %s", addr)
	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func listWalletsHandler(c *machinawallet.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()

		page, err := c.ListWallets(ctx, &machinawallet.ListWalletsOptions{
			OwnerUserID: r.URL.Query().Get("owner_user_id"),
			Chain:       machinawallet.Chain(r.URL.Query().Get("chain")),
			Limit:       50,
		})
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, page)
	}
}

type sendRequest struct {
	WalletID    string `json:"wallet_id"`
	To          string `json:"to"`
	AmountAtoms string `json:"amount_atoms"`
	Chain       string `json:"chain"`
}

func sendHandler(c *machinawallet.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req sendRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		tx, err := c.SubmitTransaction(ctx, req.WalletID, machinawallet.TxIntent{
			To:          req.To,
			AmountAtoms: req.AmountAtoms,
			Chain:       machinawallet.Chain(req.Chain),
			IntentKind:  "transfer",
		})
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, tx)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, err error) {
	var pde *machinawallet.PolicyDeniedError
	if errors.As(err, &pde) {
		writeJSON(w, http.StatusForbidden, map[string]any{
			"error":       "policy_denied",
			"policy_id":   pde.PolicyID,
			"policy_name": pde.PolicyName,
		})
		return
	}
	var ar *machinawallet.ApprovalRequiredError
	if errors.As(err, &ar) {
		writeJSON(w, http.StatusAccepted, map[string]any{
			"status":      "approval_required",
			"approval_id": ar.ApprovalID,
		})
		return
	}
	if errors.Is(err, machinawallet.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not_found"})
		return
	}
	writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("missing %s", key)
	}
	return v
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
