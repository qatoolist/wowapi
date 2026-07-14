package compatfixture

// Timeout changed from a duration-backed integer to a string.
type Timeout string

// Client is the public client contract.
type Client struct{}

// Get fetches a resource by identifier.
func (Client) Get(id string) error { return nil }
