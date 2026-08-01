// Package prismbuild is the SDK facade for PRISM Build, a product
// delivery control plane for cross-repository initiative coordination.
//
// All behavior lives in the pkg/* subpackages; this root package provides
// the high-level Client that CLI, MCP, and external importers use.
package prismbuild

import (
	"github.com/ProductBuildersHQ/prism-build/pkg/store"
)

// Client is the top-level entry point for the PRISM Control SDK.
// It wraps the service layer that CLI and MCP adapters share.
type Client struct {
	store store.Store
}

// Config holds connection and behavioral settings for a Client.
type Config struct {
	// DSN is the MySQL-compatible data source name for the Dolt server.
	DSN string
}

// New creates a Client backed by the given store implementation.
func New(s store.Store) *Client {
	return &Client{store: s}
}
