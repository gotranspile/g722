package g722

import "math"

func saturate(amp int32) int16 {
	var amp16 int16
	amp16 = int16(amp)
	if int(amp) == int(amp16) {
		return amp16
	}
	if int(amp) > math.MaxInt16 {
		return math.MaxInt16
	}
	return math.MinInt16
}
func decodeBlock4(s *Decoder, band int, d int) {
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
	if s.band[band].sg[0] == s.band[band].sg[2] {
		wd3 = 128
	} else {
		wd3 = math.MinInt8
	}
	wd3 += wd2 >> 7
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
func decoderInit(s *Decoder, rate int, options int) *Decoder {
	if s == nil {
		if (func() *Decoder {
			s = new(Decoder)
			return s
		}()) == nil {
			return nil
		}
	}
	*s = Decoder{}
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
func decode(s *Decoder, amp []int16, g722_data []uint8, len_ int) int {
	var (
		wl         [8]int  = [8]int{-60, -30, 58, 172, 334, 538, 1198, 3042}
		rl42       [16]int = [16]int{0, 7, 6, 5, 4, 3, 2, 1, 7, 6, 5, 4, 3, 2, 1, 0}
		ilb        [32]int = [32]int{2048, 2093, 2139, 2186, 2233, 2282, 2332, 2383, 2435, 2489, 2543, 2599, 2656, 2714, 2774, 2834, 2896, 2960, 3025, 3091, 3158, 3228, 3298, 3371, 3444, 3520, 3597, 3676, 3756, 3838, 3922, 4008}
		wh         [3]int  = [3]int{0, -214, 798}
		rh2        [4]int  = [4]int{2, 1, 2, 1}
		qm2        [4]int  = [4]int{-7408, -1616, 7408, 1616}
		qm4        [16]int = [16]int{0, -20456, -12896, -8968, -6288, -4240, -2584, -1200, 20456, 12896, 8968, 6288, 4240, 2584, 1200, 0}
		qm5        [32]int = [32]int{-280, -280, -23352, -17560, -14120, -11664, -9752, -8184, -6864, -5712, -4696, -3784, -2960, -2208, -1520, -880, 23352, 17560, 14120, 11664, 9752, 8184, 6864, 5712, 4696, 3784, 2960, 2208, 1520, 880, 280, -280}
		qm6        [64]int = [64]int{-136, -136, -136, -136, -24808, -21904, -19008, -16704, -14984, -13512, -12280, -11192, -10232, -9360, -8576, -7856, -7192, -6576, -6000, -5456, -4944, -4464, -4008, -3576, -3168, -2776, -2400, -2032, -1688, -1360, -1040, -728, 24808, 21904, 19008, 16704, 14984, 13512, 12280, 11192, 10232, 9360, 8576, 7856, 7192, 6576, 6000, 5456, 4944, 4464, 4008, 3576, 3168, 2776, 2400, 2032, 1688, 1360, 1040, 728, 432, 136, -432, -136}
		qmf_coeffs [12]int = [12]int{3, -11, 12, 32, -210, 951, 3876, -805, 362, -156, 53, -11}
		dlowt      int
		rlow       int
		ihigh      int
		dhigh      int
		rhigh      int
		xout1      int
		xout2      int
		wd1        int
		wd2        int
		wd3        int
		code       int
		outlen     int
		i          int
		j          int
	)
	outlen = 0
	rhigh = 0
	for j = 0; j < len_; {
		if s.packed {
			if s.inBits < s.bitsPerSample {
				s.inBuffer |= uint(int(g722_data[func() int {
					p := &j
					x := *p
					*p++
					return x
				}()]) << s.inBits)
				s.inBits += 8
			}
			code = int(s.inBuffer & uint((1<<s.bitsPerSample)-1))
			s.inBuffer >>= uint(s.bitsPerSample)
			s.inBits -= s.bitsPerSample
		} else {
			code = int(g722_data[func() int {
				p := &j
				x := *p
				*p++
				return x
			}()])
		}
		switch s.bitsPerSample {
		default:
			fallthrough
		case 8:
			wd1 = code & 0x3F
			ihigh = (code >> 6) & 0x3
			wd2 = qm6[wd1]
			wd1 >>= 2
		case 7:
			wd1 = code & 0x1F
			ihigh = (code >> 5) & 0x3
			wd2 = qm5[wd1]
			wd1 >>= 1
		case 6:
			wd1 = code & 0xF
			ihigh = (code >> 4) & 0x3
			wd2 = qm4[wd1]
		}
		wd2 = (s.band[0].det * wd2) >> 15
		rlow = s.band[0].s + wd2
		if rlow > 16383 {
			rlow = 16383
		} else if rlow < -16384 {
			rlow = -16384
		}
		wd2 = qm4[wd1]
		dlowt = (s.band[0].det * wd2) >> 15
		wd2 = rl42[wd1]
		wd1 = (s.band[0].nb * math.MaxInt8) >> 7
		wd1 += wl[wd2]
		if wd1 < 0 {
			wd1 = 0
		} else if wd1 > 18432 {
			wd1 = 18432
		}
		s.band[0].nb = wd1
		wd1 = (s.band[0].nb >> 6) & 31
		wd2 = 8 - (s.band[0].nb >> 11)
		if wd2 < 0 {
			wd3 = ilb[wd1] << (-wd2)
		} else {
			wd3 = ilb[wd1] >> wd2
		}
		s.band[0].det = wd3 << 2
		decodeBlock4(s, 0, dlowt)
		if !s.eightK {
			wd2 = qm2[ihigh]
			dhigh = (s.band[1].det * wd2) >> 15
			rhigh = dhigh + s.band[1].s
			if rhigh > 16383 {
				rhigh = 16383
			} else if rhigh < -16384 {
				rhigh = -16384
			}
			wd2 = rh2[ihigh]
			wd1 = (s.band[1].nb * math.MaxInt8) >> 7
			wd1 += wh[wd2]
			if wd1 < 0 {
				wd1 = 0
			} else if wd1 > 22528 {
				wd1 = 22528
			}
			s.band[1].nb = wd1
			wd1 = (s.band[1].nb >> 6) & 31
			wd2 = 10 - (s.band[1].nb >> 11)
			if wd2 < 0 {
				wd3 = ilb[wd1] << (-wd2)
			} else {
				wd3 = ilb[wd1] >> wd2
			}
			s.band[1].det = wd3 << 2
			decodeBlock4(s, 1, dhigh)
		}
		if s.ituTestMode {
			amp[func() int {
				p := &outlen
				x := *p
				*p++
				return x
			}()] = int16(rlow << 1)
			amp[func() int {
				p := &outlen
				x := *p
				*p++
				return x
			}()] = int16(rhigh << 1)
		} else {
			if s.eightK {
				amp[func() int {
					p := &outlen
					x := *p
					*p++
					return x
				}()] = int16(rlow << 1)
			} else {
				for i = 0; i < 22; i++ {
					s.x[i] = s.x[i+2]
				}
				s.x[22] = rlow + rhigh
				s.x[23] = rlow - rhigh
				xout1 = 0
				xout2 = 0
				for i = 0; i < 12; i++ {
					xout2 += s.x[i*2] * qmf_coeffs[i]
					xout1 += s.x[i*2+1] * qmf_coeffs[11-i]
				}
				amp[func() int {
					p := &outlen
					x := *p
					*p++
					return x
				}()] = int16(xout1 >> 11)
				amp[func() int {
					p := &outlen
					x := *p
					*p++
					return x
				}()] = int16(xout2 >> 11)
			}
		}
	}
	return outlen
}
