package register

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/subosito/daigate/gateway"
	"github.com/subosito/daigate/ingress/keyring"
	emailissuer "github.com/subosito/daigate-extensions/issuer/email"
)

func init() {
	gateway.RegisterAdminIssuer("email", mountEmailIssuer)
}

func mountEmailIssuer(mux *http.ServeMux, db *sql.DB, ks keyring.KeyStore, entry gateway.IssuerEntry) error {
	cfg, err := emailissuer.DecodeIssuerYAML(entry.Config)
	if err != nil {
		return fmt.Errorf("issuer %q: %w", entry.Driver, err)
	}
	return emailissuer.MountIssuer(mux, db, ks, cfg)
}