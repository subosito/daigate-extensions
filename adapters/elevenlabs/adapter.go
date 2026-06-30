package elevenlabs

import (
	"github.com/subosito/daigate/adaptersdk"
	"github.com/subosito/daigate/credential/inject"
)

// DefaultInject applies when providers.yaml omits inject (override with inject: map if needed).
var DefaultInject = inject.Spec{"xi-api-key": "${key}"}

// Adapter registers ElevenLabs translate speech.
//
// Operator wiring (providers.yaml):
//
//	providers:
//	  elevenlabs:
//	    credential_profile: elevenlabs
//	    # inject: { xi-api-key: "${key}" }   # optional override; default from adapter
//	    surfaces:
//	      speech:
//	        adapter: elevenlabs
//	        base_url: https://api.elevenlabs.io
//
//	models:
//	  eleven-multilingual-v2:
//	    modalities:
//	      speech:
//	        wire: openai-audio-speech
//	        providers:
//	          - provider_ref: elevenlabs
//	            surface: speech
//	            model: eleven_multilingual_v2
type Adapter struct{}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) Name() string { return "elevenlabs" }

func (a *Adapter) Register(reg *adaptersdk.Registry) error {
	adaptersdk.RegisterSpeechAdapter(reg, a.Name(), &SpeechHandler{})
	return nil
}