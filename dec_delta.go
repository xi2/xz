/*
 * Delta decoder
 *
 * Author: Lasse Collin <lasse.collin@tukaani.org>
 *
 * Translation to Go: Michael Cross <https://xi2.org/x/xz>
 *
 * This file has been put into the public domain.
 * You can do whatever you want with this file.
 */

package xz

type xzDecDelta struct {
	delta    [256]byte
	pos      byte
	distance int // in range [1, 256]
}

/*
 * Decode raw stream which has a delta filter as the first filter.
 */
func xzDecDeltaRun(s *xzDecDelta, b *xzBuf, chain func(*xzBuf) xzRet) xzRet {
	outStart := b.outPos
	ret := chain(b)
	for i := outStart; i < b.outPos; i++ {
		tmp := b.out[i] + s.delta[byte(s.distance+int(s.pos))]
		s.delta[s.pos] = tmp
		b.out[i] = tmp
		s.pos--
	}
	return ret
}

/*
 * Allocate memory for a delta decoder
 */
func xzDecDeltaCreate(distance int) *xzDecDelta {
	return &xzDecDelta{
		distance: distance,
	}
}
