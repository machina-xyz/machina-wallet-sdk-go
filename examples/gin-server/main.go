// Package main runs a Gin server that integrates the MACHINA wallet SDK and
// exposes a webhook receiver demonstrating HMAC signature verification.
//
// Run with:
//
//	MACHINA_APP_ID=app_123 \
//	MACHINA_APP_SECRET=... \
//	MACHINA_WEBHOOK_SECRET=whsec_... \
//	go run .
package main

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	machinawallet "github.com/machina-xyz/machina-wallet-sdk-go"
)

func main() {
	cfg := machinawallet.Config{
		BaseURL:   envOr("MACHINA_BASE_URL", machinawallet.DefaultBaseURL),
		AppID:     mustEnv("MACHINA_APP_ID"),
		AppSecret: mustEnv("MACHINA_APP_SECRET"),
		Timeout:   10 * time.Second,
	}
	client, err := machinawallet.NewClient(cfg)
	if err != nil {
		log.Fatalf("init client: %v", err)
	}

	webhookSecret := mustEnv("MACHINA_WEBHOOK_SECRET")

	r := gin.Default()

	r.GET("/wallets/:id", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()
		w, err := client.GetWallet(ctx, c.Param("id"))
		if err != nil {
			respondErr(c, err)
			return
		}
		c.JSON(http.StatusOK, w)
	})

	r.POST("/wallets/:id/send", func(c *gin.Context) {
		var body struct {
			To          string `json:"to"`
			AmountAtoms string `json:"amount_atoms"`
			Chain       string `json:"chain"`
		}
		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
		defer cancel()
		tx, err := client.SubmitTransaction(ctx, c.Param("id"), machinawallet.TxIntent{
			To:          body.To,
			AmountAtoms: body.AmountAtoms,
			Chain:       machinawallet.Chain(body.Chain),
			IntentKind:  "transfer",
		})
		if err != nil {
			respondErr(c, err)
			return
		}
		c.JSON(http.StatusOK, tx)
	})

	r.POST("/webhooks/machina", webhookHandler(webhookSecret))

	addr := envOr("ADDR", ":8080")
	log.Printf("listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}

// webhookHandler verifies the MACHINA signature header and dispatches on the
// CloudEvent type. The raw body is read up-front because the signature
// includes the exact bytes that were transmitted.
func webhookHandler(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read body"})
			return
		}
		sig := c.GetHeader("Machina-Signature")
		if err := machinawallet.VerifySignature(body, sig, secret, 5*time.Minute); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		ev, err := machinawallet.ParseEvent(body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		log.Printf("webhook received: type=%s id=%s", ev.Type, ev.ID)
		c.Status(http.StatusNoContent)
	}
}

func respondErr(c *gin.Context, err error) {
	var pde *machinawallet.PolicyDeniedError
	if errors.As(err, &pde) {
		c.JSON(http.StatusForbidden, gin.H{
			"error":       "policy_denied",
			"policy_id":   pde.PolicyID,
			"policy_name": pde.PolicyName,
		})
		return
	}
	var ar *machinawallet.ApprovalRequiredError
	if errors.As(err, &ar) {
		c.JSON(http.StatusAccepted, gin.H{
			"status":      "approval_required",
			"approval_id": ar.ApprovalID,
		})
		return
	}
	if errors.Is(err, machinawallet.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
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
