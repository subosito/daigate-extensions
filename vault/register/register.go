package register

import (
	"fmt"

	"github.com/subosito/daigate/credential/store"
	"github.com/subosito/daigate/gateway"
	vaultstore "github.com/subosito/daigate-extensions/vault"
)

func init() {
	gateway.RegisterCredentialBackend("vault", openBackend("vault"))
	gateway.RegisterCredentialBackend("hybrid", openBackend("hybrid"))
}

func openBackend(mode string) gateway.CredentialBackendOpener {
	return func(f *gateway.ConfigFile, sqlite *store.SQLite) (store.Store, error) {
		cfg, err := vaultstore.DecodeVaultYAML(f.Credential.BackendConfig)
		if err != nil {
			return nil, fmt.Errorf("credential backend %q: %w", mode, err)
		}
		vc, err := vaultstore.NewClient(cfg)
		if err != nil {
			return nil, err
		}
		return vaultstore.NewBackend(sqlite, vc, cfg, mode), nil
	}
}