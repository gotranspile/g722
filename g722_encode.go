package g722

import "math"

func encodeBlock4(s *Encoder, band int, d int) {
	var (
		wd1 int
		wd2 int
		wd3 int
		i   int
	)
	s.band[band].d[0] = d
	s.band[band].r[0] = int(saturate(int32(s.band[band].s + d)))
	s.band[band].p[0] = int(saturate(int32(s.band[band].sz + d)))
	for i = 0; i < 3; i++ {
		s.band[band].sg[i] = s.band[band].p[i] >> 15
	}
	wd1 = int(saturate(int32(s.band[band].a[1] << 2)))
	if s.band[band].sg[0] == s.band[band].sg[1] {
		wd2 = -wd1
	} else {
		wd2 = wd1
	}
	if wd2 > math.MaxInt16 {
		wd2 = math.MaxInt16
	}
	wd3 = (wd2 >> 7) + int(func() int {
		if s.band[band].sg[0] == s.band[band].sg[2] {
			return 128
		}
		return -128
	}())
	wd3 += (s.band[band].a[2] * 32512) >> 15
	if wd3 > 12288 {
		wd3 = 12288
	} else if wd3 < -12288 {
		wd3 = -12288
	}
	s.band[band].ap[2] = wd3
	s.band[band].sg[0] = s.band[band].p[0] >> 15
	s.band[band].sg[1] = s.band[band].p[1] >> 15
	if s.band[band].sg[0] == s.band[band].sg[1] {
		wd1 = 192
	} else {
		wd1 = -192
	}
	wd2 = (s.band[band].a[1] * 32640) >> 15
	s.band[band].ap[1] = int(saturate(int32(wd1 + wd2)))
	wd3 = int(saturate(int32(15360 - s.band[band].ap[2])))
	if s.band[band].ap[1] > wd3 {
		s.band[band].ap[1] = wd3
	} else if s.band[band].ap[1] < -wd3 {
		s.band[band].ap[1] = -wd3
	}
	if d == 0 {
		wd1 = 0
	} else {
		wd1 = 128
	}
	s.band[band].sg[0] = d >> 15
	for i = 1; i < 7; i++ {
		s.band[band].sg[i] = s.band[band].d[i] >> 15
		if s.band[band].sg[i] == s.band[band].sg[0] {
			wd2 = wd1
		} else {
			wd2 = -wd1
		}
		wd3 = (s.band[band].b[i] * 32640) >> 15
		s.band[band].bp[i] = int(saturate(int32(wd2 + wd3)))
	}
	for i = 6; i > 0; i-- {
		s.band[band].d[i] = s.band[band].d[i-1]
		s.band[band].b[i] = s.band[band].bp[i]
	}
	for i = 2; i > 0; i-- {
		s.band[band].r[i] = s.band[band].r[i-1]
		s.band[band].p[i] = s.band[band].p[i-1]
		s.band[band].a[i] = s.band[band].ap[i]
	}
	wd1 = int(saturate(int32(s.band[band].r[1] + s.band[band].r[1])))
	wd1 = (s.band[band].a[1] * wd1) >> 15
	wd2 = int(saturate(int32(s.band[band].r[2] + s.band[band].r[2])))
	wd2 = (s.band[band].a[2] * wd2) >> 15
	s.band[band].sp = int(saturate(int32(wd1 + wd2)))
	s.band[band].sz = 0
	for i = 6; i > 0; i-- {
		wd1 = int(saturate(int32(s.band[band].d[i] + s.band[band].d[i])))
		s.band[band].sz += (s.band[band].b[i] * wd1) >> 15
	}
	s.band[band].sz = int(saturate(int32(s.band[band].sz)))
	s.band[band].s = int(saturate(int32(s.band[band].sp + s.band[band].sz)))
}
func encoderInit(s *Encoder, rate int, options int) *Encoder {
	if s == nil {
		if (func() *Encoder {
			s = new(Encoder)
			return s
		}()) == nil {
			return nil
		}
	}
	*s = Encoder{}
	if rate == 48000 {
		s.bitsPerSample = 6
	} else if rate == 56000 {
		s.bitsPerSample = 7
	} else {
		s.bitsPerSample = 8
	}
	if (options & FlagSampleRate8000) != 0 {
		s.eightK = true
	}
	if (options&FlagPacked) != 0 && s.bitsPerSample != 8 {
		s.packed = true
	} else {
		s.packed = false
	}
	s.band[0].det = 32
	s.band[1].det = 8
	return s
}
func encode(s *Encoder, g722_data []uint8, amp []int16, len_ int) int {
	var (
		q6         [32]int = [32]int{0, 35, 72, 110, 150, 190, 233, 276, 323, 370, 422, 473, 530, 587, 650, 714, 786, 858, 940, 1023, 1121, 1219, 1339, 1458, 1612, 1765, 1980, 2195, 2557, 2919, 0, 0}
		iln        [32]int = [32]int{0, 63, 62, 31, 30, 29, 28, 27, 26, 25, 24, 23, 22, 21, 20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 0}
		ilp        [32]int = [32]int{0, 61, 60, 59, 58, 57, 56, 55, 54, 53, 52, 51, 50, 49, 48, 47, 46, 45, 44, 43, 42, 41, 40, 39, 38, 37, 36, 35, 34, 33, 32, 0}
		wl         [8]int  = [8]int{-60, -30, 58, 172, 334, 538, 1198, 3042}
		rl42       [16]int = [16]int{0, 7, 6, 5, 4, 3, 2, 1, 7, 6, 5, 4, 3, 2, 1, 0}
		ilb        [32]int = [32]int{2048, 2093, 2139, 2186, 2233, 2282, 2332, 2383, 2435, 2489, 2543, 2599, 2656, 2714, 2774, 2834, 2896, 2960, 3025, 3091, 3158, 3228, 3298, 3371, 3444, 3520, 3597, 3676, 3756, 3838, 3922, 4008}
		qm4        [16]int = [16]int{0, -20456, -12896, -8968, -6288, -4240, -2584, -1200, 20456, 12896, 8968, 6288, 4240, 2584, 1200, 0}
		qm2        [4]int  = [4]int{-7408, -1616, 7408, 1616}
		qmf_coeffs [12]int = [12]int{3, -11, 12, 32, -210, 951, 3876, -805, 362, -156, 53, -11}
		ihn        [3]int  = [3]int{0, 1, 0}
		ihp        [3]int  = [3]int{0, 3, 2}
		wh         [3]int  = [3]int{0, -214, 798}
		rh2        [4]int  = [4]int{2, 1, 2, 1}
		dlow       int
		dhigh      int
		el         int
		wd         int
		wd1        int
		ril        int
		wd2        int
		il4        int
		ih2        int
		wd3        int
		eh         int
		mih        int
		i          int
		j          int
		xlow       int
		xhigh      int
		g722_bytes int
		sumeven    int
		sumodd     int
		ihigh      int
		ilow       int
		code       int
	)
	g722_bytes = 0
	xhigh = 0
	for j = 0; j < len_; {
		if s.ituTestMode {
			xlow = func() int {
				xhigh = int(amp[func() int {
					p := &j
					x := *p
					*p++
					return x
				}()]) >> 1
				return xhigh
			}()
		} else {
			if s.eightK {
				xlow = int(amp[func() int {
					p := &j
					x := *p
					*p++
					return x
				}()]) >> 1
			} else {
				for i = 0; i < 22; i++ {
					s.x[i] = s.x[i+2]
				}
				s.x[22] = int(amp[func() int {
					p := &j
					x := *p
					*p++
					return x
				}()])
				s.x[23] = int(amp[func() int {
					p := &j
					x := *p
					*p++
					return x
				}()])
				sumeven = 0
				sumodd = 0
				for i = 0; i < 12; i++ {
					sumodd += s.x[i*2] * qmf_coeffs[i]
					sumeven += s.x[i*2+1] * qmf_coeffs[11-i]
				}
				xlow = (sumeven + sumodd) >> 14
				xhigh = (sumeven - sumodd) >> 14
			}
		}
		el = int(saturate(int32(xlow - s.band[0].s)))
		if el >= 0 {
			wd = el
		} else {
			wd = -(el + 1)
		}
		for i = 1; i < 30; i++ {
			wd1 = (q6[i] * s.band[0].det) >> 12
			if wd < wd1 {
				break
			}
		}
		if el < 0 {
			ilow = iln[i]
		} else {
			ilow = ilp[i]
		}
		ril = ilow >> 2
		wd2 = qm4[ril]
		dlow = (s.band[0].det * wd2) >> 15
		il4 = rl42[ril]
		wd = (s.band[0].nb * math.MaxInt8) >> 7
		s.band[0].nb = wd + wl[il4]
		if s.band[0].nb < 0 {
			s.band[0].nb = 0
		} else if s.band[0].nb > 18432 {
			s.band[0].nb = 18432
		}
		wd1 = (s.band[0].nb >> 6) & 31
		wd2 = 8 - (s.band[0].nb >> 11)
		if wd2 < 0 {
			wd3 = ilb[wd1] << (-wd2)
		} else {
			wd3 = ilb[wd1] >> wd2
		}
		s.band[0].det = wd3 << 2
		encodeBlock4(s, 0, dlow)
		if s.eightK {
			code = (ilow | 0xC0) >> (8 - s.bitsPerSample)
		} else {
			eh = int(saturate(int32(xhigh - s.band[1].s)))
			if eh >= 0 {
				wd = eh
			} else {
				wd = -(eh + 1)
			}
			wd1 = (s.band[1].det * 564) >> 12
			if wd >= wd1 {
				mih = 2
			} else {
				mih = 1
			}
			if eh < 0 {
				ihigh = ihn[mih]
			} else {
				ihigh = ihp[mih]
			}
			wd2 = qm2[ihigh]
			dhigh = (s.band[1].det * wd2) >> 15
			ih2 = rh2[ihigh]
			wd = (s.band[1].nb * math.MaxInt8) >> 7
			s.band[1].nb = wd + wh[ih2]
			if s.band[1].nb < 0 {
				s.band[1].nb = 0
			} else if s.band[1].nb > 22528 {
				s.band[1].nb = 22528
			}
			wd1 = (s.band[1].nb >> 6) & 31
			wd2 = 10 - (s.band[1].nb >> 11)
			if wd2 < 0 {
				wd3 = ilb[wd1] << (-wd2)
			} else {
				wd3 = ilb[wd1] >> wd2
			}
			s.band[1].det = wd3 << 2
			encodeBlock4(s, 1, dhigh)
			code = ((ihigh << 6) | ilow) >> (8 - s.bitsPerSample)
		}
		if s.packed {
			s.outBuffer |= uint(code << s.outBits)
			s.outBits += s.bitsPerSample
			if s.outBits >= 8 {
				g722_data[func() int {
					p := &g722_bytes
					x := *p
					*p++
					return x
				}()] = uint8(s.outBuffer & math.MaxUint8)
				s.outBits -= 8
				s.outBuffer >>= 8
			}
		} else {
			g722_data[func() int {
				p := &g722_bytes
				x := *p
				*p++
				return x
			}()] = uint8(int8(code))
		}
	}
	return g722_bytes
}
