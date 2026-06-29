package compose

import (
	"github.com/subosito/daigate/adaptersdk"
	corecompose "github.com/subosito/daigate/compose"
	"github.com/subosito/daigate-extensions/adapters/elevenlabs"
	"github.com/subosito/daigate/passthrough"
)

// ExtAdapters returns optional vendor translate adapters shipped in this module.
func ExtAdapters() []adaptersdk.Adapter {
	return []adaptersdk.Adapter{
		elevenlabs.New(),
	}
}

// AllAdapters returns core passthrough plus ExtAdapters — convenience for operator binaries.
func AllAdapters() []adaptersdk.Adapter {
	return append([]adaptersdk.Adapter{passthrough.New()}, ExtAdapters()...)
}

// FromConfig filters AllAdapters by daigate.yaml adapters.enable.
func FromConfig(enable []string) (*adaptersdk.Registry, error) {
	return corecompose.FromConfig(enable, AllAdapters())
}