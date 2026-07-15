package compatfixture

import "time"

// Timeout is the public request timeout type.
type Timeout time.Duration

// Client is the public client contract.
type Client struct{}

// Get fetches a resource by identifier.
func (Client) Get(id string) error { return nil }
