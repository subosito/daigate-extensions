package compose_test

import (
	"testing"

	"github.com/subosito/daigate-extensions/compose"
)

func TestFromConfigPassthroughOnly(t *testing.T) {
	reg, err := compose.FromConfig([]string{"passthrough"})
	if err != nil {
		t.Fatal(err)
	}
	if len(reg.ChatHandlers) == 0 {
		t.Fatal("expected chat handlers")
	}
}

func TestFromConfigElevenLabs(t *testing.T) {
	reg, err := compose.FromConfig([]string{"passthrough", "elevenlabs"})
	if err != nil {
		t.Fatal(err)
	}
	if reg.SpeechAdapters["elevenlabs"] == nil {
		t.Fatal("expected elevenlabs speech adapter")
	}
}