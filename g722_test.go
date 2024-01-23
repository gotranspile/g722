package g722

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"os"
	"testing"
)

func TestG722(t *testing.T) {
	data, err := os.ReadFile("testdata/sample.g722")
	if err != nil {
		t.Fatal(err)
	}
	t.Run("8KHz", func(t *testing.T) {
		pcm := decodeWithFlags(data, FlagSampleRate8000)

		if got := hashData(pcm16bytes(pcm)); got != "a577a27175bba3b47e04d7db4b9743c6db7e057265b5293a592d7b7ba4bc7a62" {
			t.Errorf("unexpected PCM hash: %s", got)
		}

		g722 := Encode(pcm, RateDefault, FlagSampleRate8000)
		if got := hashData(g722); got != "6a04cd08047dd1ee2440d18ff969b209f6805abf7f70f9b64a1efe0fe3b03924" {
			t.Errorf("unexpected G722 hash: %s [%d]", got, len(g722))
		}
	})
	t.Run("16KHz", func(t *testing.T) {
		pcm := decodeWithFlags(data, 0)

		if got := hashData(pcm16bytes(pcm)); got != "790a4e6558707c2ae88cedbc3535bce1f731f4bcff8d0993cd2e34ad1d40ca53" {
			t.Errorf("unexpected PCM hash: %s", got)
		}

		g722 := Encode(pcm, RateDefault, 0)
		if got := hashData(g722); got != "f1d1816d75e33937c5a64163c1a5006d9c6d1fed28e295b57356a9638392683a" {
			t.Errorf("unexpected G722 hash: %s [%d]", got, len(g722))
		}
	})
}

func decodeWithFlags(data []byte, flags Flags) []int16 {
	return Decode(data, RateDefault, flags)
}

func pcm16bytes(amp []int16) []byte {
	out := make([]byte, 2*len(amp))
	for i, v := range amp {
		binary.LittleEndian.PutUint16(out[2*i:], uint16(v))
	}
	return out
}

func hashData(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
