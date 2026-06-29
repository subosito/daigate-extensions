package elevenlabs

import (
	"github.com/subosito/daigate/adaptersdk"
)

// Adapter registers ElevenLabs translate speech.
//
// Operator wiring (providers.yaml):
//
//	providers:
//	  elevenlabs:
//	    credential_profile: elevenlabs
//	    inject_preset: xi-api-key
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