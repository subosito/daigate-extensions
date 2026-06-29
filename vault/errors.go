package vault

import "errors"

// ErrNotFound means the KV path does not exist.
var ErrNotFound = errors.New("vault: secret not found")