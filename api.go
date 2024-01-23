package g722

const (
	RateDefault = Rate64000
	Rate64000   = 64000
	Rate56000   = 56000
	Rate48000   = 48000
)

type Flags byte

func Encode(pcm []int16, rate int, flags Flags) []byte {
	var enc Encoder
	encoderInit(&enc, rate, int(flags))

	div := 2
	if flags&FlagSampleRate8000 != 0 {
		div = 1
	}
	n := len(pcm) / div
	if len(pcm)%div != 0 {
		n++
	}
	out := make([]byte, n)
	n = encode(&enc, out, pcm, len(pcm))
	return out[:n]
}

func NewEncoder(rate int, flags Flags) *Encoder {
	var enc Encoder
	encoderInit(&enc, rate, int(flags))
	return &enc
}

func (enc *Encoder) Encode(dst []byte, src []int16) int {
	return encode(enc, dst, src, len(src))
}

func Decode(g722 []byte, rate int, flags Flags) []int16 {
	var dec Decoder
	decoderInit(&dec, rate, int(flags))

	mul := 2
	if flags&FlagSampleRate8000 != 0 {
		mul = 1
	}
	n := len(g722) * mul
	out := make([]int16, n)
	n = decode(&dec, out, g722, len(g722))
	return out[:n]
}

func NewDecoder(rate int, flags Flags) *Decoder {
	var dec Decoder
	decoderInit(&dec, rate, int(flags))
	return &dec
}

func (dec *Decoder) Decode(dst []int16, src []byte) int {
	return decode(dec, dst, src, len(src))
}
