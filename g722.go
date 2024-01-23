package g722

const (
	FlagSampleRate8000 = 0x1
	FlagPacked         = 0x2
)

type Encoder struct {
	ituTestMode   bool
	packed        bool
	eightK        bool
	bitsPerSample int
	x             [24]int
	band          [2]struct {
		s   int
		sp  int
		sz  int
		r   [3]int
		a   [3]int
		ap  [3]int
		p   [3]int
		d   [7]int
		b   [7]int
		bp  [7]int
		sg  [7]int
		nb  int
		det int
	}
	inBuffer  uint
	inBits    int
	outBuffer uint
	outBits   int
}
type Decoder struct {
	ituTestMode   bool
	packed        bool
	eightK        bool
	bitsPerSample int
	x             [24]int
	band          [2]struct {
		s   int
		sp  int
		sz  int
		r   [3]int
		a   [3]int
		ap  [3]int
		p   [3]int
		d   [7]int
		b   [7]int
		bp  [7]int
		sg  [7]int
		nb  int
		det int
	}
	inBuffer  uint
	inBits    int
	outBuffer uint
	outBits   int
}
