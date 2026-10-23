package g722

import (
	"bytes"
	"math"
	"strconv"
	"testing"
)

// Reference hashes were produced by github.com/sippy/libg722 (the spandsp
// G.722 codec, which passes the ITU test vectors) through its Python binding
// G722 1.2.8, encoding each input with a fresh encoder and decoding the result
// with a fresh decoder at 16 kHz.
var referenceVectors = []struct {
	input   string
	rate    int
	encoded string
	decoded string
}{
	{"const_pos", Rate64000, "b1f66d52489b3a1cbc0a5cce7902c74bf925375fc48aac128c5ba3d25a06ccb4", "ec4dd93e39270bebd0152e264cfbda99398311c2d6ad1c16f09160bdd3ed2a01"},
	{"const_pos", Rate56000, "cb71068cd2b1821bcfc6510e07c271ff1baafbd838c42a91c06c47d174c048a1", "b7238efb7bf1ff4b8c49c8971d3798e6ca78efd03d487cd335dddfd75d084da3"},
	{"const_pos", Rate48000, "aef7b3ae37b9466689d677b726936603b10105f0bb2f75502419628e65ce492f", "56e6c0a109f5d153e34e00d8d2dc0ec3c374884afed661f307b36d2fb4751a20"},
	{"const_neg", Rate64000, "b363a1dc124fb113db4916fec968b0e398d49fccd500b3ec24912290edca3ae8", "d39ba2a8032308f7ff87b78a7639fb601c1e80725222c181c6ef76303bee380a"},
	{"const_neg", Rate56000, "3a78ce04ad1b82e4376186e76b26bf3460c68b9449a46bb96680d54fc2bc26f7", "b132748172b8bc45d62f01ecb9edc3d9f372b4cbf3dd730aebdf8d06fab9d0d5"},
	{"const_neg", Rate48000, "3c4c76e63312934e88b54908f39fb577c198053696918065ee7bc83bde092509", "b22d8b3cb881efe23a92d0fdda0c61e211f3948e4d1bd26d370e63e561ab9152"},
	{"square_fs", Rate64000, "317f9f5df94d5fc5911795a02e97f0d119a8eac10e57d588b80ce1a23d82f795", "5ea5770d822023c1636a39032a5c2d0c6d33ad433fdbdb912a762d11b652ad8a"},
	{"square_fs", Rate56000, "42edfaae71b6cd72c7fe90ebd2ca96e600add334d1905049ed546d49de97bc10", "be5f2e28635a238dd43b573dd055164b49fbad2d40a44343e900da7cefd49d46"},
	{"square_fs", Rate48000, "ae6fdad9b99c1011e011f215dcfc860854eb51df06a49998f23c7fdc0765c797", "39a54bbbda9b79bc6f9778c0004fe1137bda519d3411e24a5650d7d2e71bbb64"},
	{"noise", Rate64000, "b451e5818610e0b25cd15d6477ff1f2df8bca3134f1c99838680084e7e3d1ab8", "e3b62c43191294d1e2466b34a28e392405a353d11b20b9d60f7f3d362b588b7d"},
	{"noise", Rate56000, "916ac8a6b808ca8128609e5c94bd41c4f025ee65f8d05e86251683b9582c45f1", "29dd0969baea7ce701b5f2e93780d39ed8fc1281bb2325a0a9acfbe8453a84e3"},
	{"noise", Rate48000, "42a6b1903bbd76792d256132e58dcea774b040152012b351208dba4f82c8fd0f", "f39b1b77815d836181c31772442ac6eda95599853ee82c344f154cc644d3c829"},
	{"triangle_12db", Rate64000, "97285329dfa4b6efa20baa2c855e7a50652ba48590ff061c8bfee94a30f11a60", "b3875549a5fbefa34d3fe599ced10b533cfcc1af02c13d7613901b321e70980d"},
	{"triangle_12db", Rate56000, "db32db07861906129e3a0c763c50c006764c845370afde9c3b862c608f8de9a2", "340eb0db1d4a746b8856ce60c170d7839101b5f17d7e9a32faf3f83d2bf6da64"},
	{"triangle_12db", Rate48000, "d91bba6df280dc45e597eb41f3d62cab9e2c63bd14fdc2432bb1f0a2cfadedb4", "f28244b5a1ef83c703b13fced9b361a2aff584e0898677f612d5615ad95122b8"},
}

const referenceSamples = 16000

// referenceInput builds the deterministic, integer-only test signals shared
// with the reference generator.
func referenceInput(name string) []int16 {
	pcm := make([]int16, referenceSamples)
	switch name {
	case "const_pos":
		for i := range pcm {
			pcm[i] = math.MaxInt16
		}
	case "const_neg":
		for i := range pcm {
			pcm[i] = math.MinInt16
		}
	case "square_fs":
		for i := range pcm {
			if (i/8)%2 == 0 {
				pcm[i] = math.MaxInt16
			} else {
				pcm[i] = math.MinInt16
			}
		}
	case "noise":
		x := uint32(1)
		for i := range pcm {
			x = (x*1103515245 + 12345) & 0x7fffffff
			pcm[i] = int16(uint16(x >> 15))
		}
	case "triangle_12db":
		for i := range pcm {
			p := i % 64
			var v int
			switch {
			case p < 16:
				v = p * 1024
			case p < 48:
				v = 32*1024 - p*1024
			default:
				v = (p - 64) * 1024
			}
			pcm[i] = int16(v / 2)
		}
	default:
		panic("unknown reference input " + name)
	}
	return pcm
}

func TestReferenceVectors(t *testing.T) {
	for _, vector := range referenceVectors {
		t.Run(vector.input+"/"+strconv.Itoa(vector.rate), func(t *testing.T) {
			encoded := Encode(referenceInput(vector.input), vector.rate, 0)
			if got := hashData(encoded); got != vector.encoded {
				t.Errorf("encoded hash = %s, want %s", got, vector.encoded)
			}
			decoded := Decode(encoded, vector.rate, 0)
			if got := hashData(pcm16bytes(decoded)); got != vector.decoded {
				t.Errorf("decoded hash = %s, want %s", got, vector.decoded)
			}
		})
	}
}

// TestDecodeSaturatesFullScaleInput is the regression test for the QMF output
// wrapping around int16 (gotranspile/g722#1): full-scale input must decode to
// near full-scale output of the same sign once the decoder has converged.
func TestDecodeSaturatesFullScaleInput(t *testing.T) {
	for _, tc := range []struct {
		input    string
		positive bool
	}{
		{"const_pos", true},
		{"const_neg", false},
	} {
		t.Run(tc.input, func(t *testing.T) {
			decoded := Decode(Encode(referenceInput(tc.input), Rate64000, 0), Rate64000, 0)
			// Skip the adaptation transient at the start of the stream.
			for i, sample := range decoded[1000:] {
				if tc.positive && sample < 16384 || !tc.positive && sample > -16384 {
					t.Fatalf("sample %d = %d wrapped away from full scale", i+1000, sample)
				}
			}
		})
	}
}

func TestSaturate(t *testing.T) {
	for _, tc := range []struct {
		in   int32
		want int16
	}{
		{math.MinInt16 - 1, math.MinInt16},
		{math.MinInt16, math.MinInt16},
		{0, 0},
		{math.MaxInt16, math.MaxInt16},
		{math.MaxInt16 + 1, math.MaxInt16},
		{math.MaxInt32, math.MaxInt16},
		{math.MinInt32, math.MinInt16},
	} {
		if got := saturate(tc.in); got != tc.want {
			t.Errorf("saturate(%d) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

// TestStreamingMatchesOneShot proves encoder and decoder state carry across
// calls, so 20 ms RTP frames produce the same stream as a single call.
func TestStreamingMatchesOneShot(t *testing.T) {
	const frameSamples = 320
	pcm := referenceInput("noise")
	oneShotEncoded := Encode(pcm, Rate64000, 0)

	encoder := NewEncoder(Rate64000, 0)
	decoder := NewDecoder(Rate64000, 0)
	payload := make([]byte, frameSamples/2)
	frame := make([]int16, frameSamples)
	var streamedEncoded []byte
	var streamedDecoded []int16
	for offset := 0; offset < len(pcm); offset += frameSamples {
		if n := encoder.Encode(payload, pcm[offset:offset+frameSamples]); n != len(payload) {
			t.Fatalf("encoded %d bytes for a 20 ms frame, want %d", n, len(payload))
		}
		streamedEncoded = append(streamedEncoded, payload...)
		if n := decoder.Decode(frame, payload); n != frameSamples {
			t.Fatalf("decoded %d samples for a 20 ms frame, want %d", n, frameSamples)
		}
		streamedDecoded = append(streamedDecoded, frame...)
	}
	if !bytes.Equal(streamedEncoded, oneShotEncoded) {
		t.Fatal("frame-by-frame encoding differs from one-shot encoding")
	}
	if !bytes.Equal(pcm16bytes(streamedDecoded), pcm16bytes(Decode(oneShotEncoded, Rate64000, 0))) {
		t.Fatal("frame-by-frame decoding differs from one-shot decoding")
	}
}
