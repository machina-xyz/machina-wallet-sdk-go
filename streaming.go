package machinawallet

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// StreamTransactions opens a Server-Sent Events stream of Transaction records
// for the given wallet. The two returned channels deliver events and a single
// terminal error respectively. Both channels close when the context is
// canceled or when the stream ends.
//
// Typical usage:
//
//	events, errs, err := client.StreamTransactions(ctx, walletID)
//	if err != nil {
//	    return err
//	}
//	for {
//	    select {
//	    case tx, ok := <-events:
//	        if !ok { return nil }
//	        handle(tx)
//	    case err := <-errs:
//	        return err
//	    case <-ctx.Done():
//	        return ctx.Err()
//	    }
//	}
func (c *Client) StreamTransactions(ctx context.Context, walletID string) (<-chan Transaction, <-chan error, error) {
	if walletID == "" {
		return nil, nil, &Error{Code: ErrCodeValidation, Message: "walletID is required"}
	}

	path := fmt.Sprintf("/v1/wallets/%s/transactions/stream", walletID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.cfg.BaseURL+path, nil)
	if err != nil {
		return nil, nil, &Error{Code: ErrCodeNetwork, Message: "building stream request: " + err.Error(), Cause: err}
	}
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	req.Header.Set("Cache-Control", "no-cache")

	sig, mode, err := c.signer.Sign(http.MethodGet, path, nil, time.Now())
	if err != nil {
		return nil, nil, &Error{Code: ErrCodeAuth, Message: "signing stream request: " + err.Error(), Cause: err}
	}
	req.Header.Set(HeaderService, c.cfg.AppID)
	req.Header.Set(HeaderTimestamp, strconv.FormatInt(time.Now().Unix(), 10))
	req.Header.Set(HeaderSignature, sig)
	req.Header.Set(HeaderMode, mode)

	// Use a streaming HTTP client without the global timeout so the connection
	// can stay open until the context is canceled.
	streamClient := &http.Client{Transport: c.transport.http.Transport}
	resp, err := streamClient.Do(req)
	if err != nil {
		return nil, nil, &Error{Code: ErrCodeNetwork, Message: err.Error(), Cause: err}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, nil, mapResponseError(resp.StatusCode, body)
	}

	events := make(chan Transaction, 16)
	errs := make(chan error, 1)
	go runSSE(ctx, resp.Body, events, errs)
	return events, errs, nil
}

// runSSE parses a Server-Sent Events stream from rc and pushes parsed
// Transaction events into out. It is the single goroutine that owns the
// response body and the two output channels.
func runSSE(ctx context.Context, rc io.ReadCloser, out chan<- Transaction, errs chan<- error) {
	defer close(out)
	defer close(errs)
	defer func() { _ = rc.Close() }()

	scanner := bufio.NewScanner(rc)
	// SSE events can be large; allow up to 1 MiB lines.
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	var dataLines []string
	flush := func() {
		if len(dataLines) == 0 {
			return
		}
		payload := strings.Join(dataLines, "\n")
		dataLines = dataLines[:0]
		if strings.TrimSpace(payload) == "" {
			return
		}
		var tx Transaction
		if err := json.Unmarshal([]byte(payload), &tx); err != nil {
			// Skip unparseable events rather than tearing down the stream.
			return
		}
		select {
		case out <- tx:
		case <-ctx.Done():
		}
	}

	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			select {
			case errs <- err:
			default:
			}
			return
		}
		line := scanner.Text()
		if line == "" {
			flush()
			continue
		}
		if strings.HasPrefix(line, ":") {
			// Comment line; SSE keepalive.
			continue
		}
		if strings.HasPrefix(line, "data:") {
			dataLines = append(dataLines, strings.TrimPrefix(strings.TrimPrefix(line, "data:"), " "))
			continue
		}
		// Ignore other fields (event:, id:, retry:) for the Transaction stream.
	}
	flush()
	if err := scanner.Err(); err != nil {
		select {
		case errs <- &Error{Code: ErrCodeNetwork, Message: "stream read: " + err.Error(), Cause: err}:
		default:
		}
	}
}
