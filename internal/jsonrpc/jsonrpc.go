// Package jsonrpc contains internal HTTP helpers shared by the SDK.
//
// This package is internal and not intended for use by SDK consumers.
package jsonrpc

import (
	"encoding/json"
	"io"
)

// DecodeJSON reads a JSON value from r into v, returning a descriptive error
// on failure. It is exposed here so the test suite and any future internal
// helpers share a single decoder configuration.
func DecodeJSON(r io.Reader, v any) error {
	dec := json.NewDecoder(r)
	dec.UseNumber()
	return dec.Decode(v)
}
