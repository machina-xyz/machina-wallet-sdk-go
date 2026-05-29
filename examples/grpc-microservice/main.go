// Package main runs a gRPC microservice that wraps the MACHINA wallet SDK
// behind a small wallet-gateway proto API.
//
// Why this pattern matters: in Kubernetes-native deployments it is common to
// run a sidecar that owns a credential and exposes domain APIs to colocated
// containers over loopback. This sidecar centralizes signing, retries, and
// observability for every wallet call in the Pod.
//
// Run with:
//
//	MACHINA_APP_ID=app_123 MACHINA_APP_SECRET=... go run .
package main

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	machinawallet "github.com/machina-xyz/machina-wallet-sdk-go"
	pb "github.com/machina-xyz/machina-wallet-sdk-go/examples/grpc-microservice/pb"
)

type gateway struct {
	pb.UnimplementedWalletGatewayServer
	client *machinawallet.Client
}

func (g *gateway) GetWallet(ctx context.Context, req *pb.GetWalletRequest) (*pb.Wallet, error) {
	w, err := g.client.GetWallet(ctx, req.GetWalletId())
	if err != nil {
		return nil, toGRPCStatus(err)
	}
	return walletToProto(w), nil
}

func (g *gateway) ListWallets(ctx context.Context, req *pb.ListWalletsRequest) (*pb.WalletPage, error) {
	page, err := g.client.ListWallets(ctx, &machinawallet.ListWalletsOptions{
		OwnerUserID: req.GetOwnerUserId(),
		Chain:       machinawallet.Chain(req.GetChain()),
		Status:      machinawallet.WalletStatus(req.GetStatus()),
		Limit:       int(req.GetLimit()),
		Cursor:      req.GetCursor(),
	})
	if err != nil {
		return nil, toGRPCStatus(err)
	}
	out := &pb.WalletPage{Items: make([]*pb.Wallet, 0, len(page.Items))}
	for i := range page.Items {
		out.Items = append(out.Items, walletToProto(&page.Items[i]))
	}
	if page.NextCursor != nil {
		out.NextCursor = *page.NextCursor
	}
	return out, nil
}

func (g *gateway) SubmitTransaction(ctx context.Context, req *pb.SubmitTransactionRequest) (*pb.SubmittedTransaction, error) {
	intent := machinawallet.TxIntent{
		To:          req.GetTo(),
		AmountAtoms: req.GetAmountAtoms(),
		Chain:       machinawallet.Chain(req.GetChain()),
		IntentKind:  req.GetIntentKind(),
	}
	if intent.IntentKind == "" {
		intent.IntentKind = "transfer"
	}
	tx, err := g.client.SubmitTransaction(ctx, req.GetWalletId(), intent)
	if err != nil {
		return nil, toGRPCStatus(err)
	}
	resp := &pb.SubmittedTransaction{
		TxId:   tx.TxID,
		Status: string(tx.Status),
	}
	if tx.ChainTxHash != nil {
		resp.ChainTxHash = *tx.ChainTxHash
	}
	return resp, nil
}

func walletToProto(w *machinawallet.Wallet) *pb.Wallet {
	return &pb.Wallet{
		Id:          w.ID,
		OwnerUserId: w.OwnerUserID,
		Chain:       string(w.Chain),
		Address:     w.Address,
		Status:      string(w.Status),
		CustodyKind: string(w.CustodyKind),
	}
}

// toGRPCStatus maps SDK errors to gRPC status codes.
func toGRPCStatus(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, machinawallet.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, machinawallet.ErrAuth):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, machinawallet.ErrPolicyDenied):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, machinawallet.ErrApprovalRequired):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, machinawallet.ErrRateLimit):
		return status.Error(codes.ResourceExhausted, err.Error())
	case errors.Is(err, machinawallet.ErrValidation):
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return status.Error(codes.Internal, err.Error())
}

func main() {
	client, err := machinawallet.NewClient(machinawallet.Config{
		BaseURL:   envOr("MACHINA_BASE_URL", machinawallet.DefaultBaseURL),
		AppID:     mustEnv("MACHINA_APP_ID"),
		AppSecret: mustEnv("MACHINA_APP_SECRET"),
		Timeout:   10 * time.Second,
	})
	if err != nil {
		log.Fatalf("init client: %v", err)
	}

	addr := envOr("ADDR", ":50051")
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	srv := grpc.NewServer()
	pb.RegisterWalletGatewayServer(srv, &gateway{client: client})
	log.Printf("wallet-gateway listening on %s", addr)
	if err := srv.Serve(lis); err != nil {
		log.Fatal(err)
	}
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
