// State is 4 bit data (0000xyyz),
// where the rightmost bit (z) is indicating if it should record space or nah.
// The 2 bits (yy) are indicating the state (as defined with type "state")
// The remaining 1 bit (x) indicating that the parser should skip all spaces until it found non-space char

package parser

type state uint8

var default_state state = state(0b0010)

func (s state) getArgtype() argtype {
	return argtype((s & 0b0110) >> 1)
}

func (s state) doRecordSpace() uint8 {
	return uint8(s) & 0b0001
}

func (s state) doSkipAllSpace() uint8 {
	return (uint8(s) & 0b1000) >> 3
}

func (s state) parseState() (argtype, uint8) {
	return s.getArgtype(), s.doRecordSpace()
}

func (s *state) changeState(ss uint8, at argtype, rs uint8) {
	if ss != 0 && ss != 1 {
		panic("invalid ss value")
	}
	if rs != 0 && rs != 1 {
		panic("invalid rs value")
	}
	if at > 3 {
		panic("invalid at value")
	}

	*s = state((ss << 3) | (uint8(at) << 1) | rs)
}

func (s *state) changeSkipSpace(ss uint8) {
	s.changeState(ss, s.getArgtype(), s.doRecordSpace())
}

func (s *state) changeArgtype(at argtype) {
	s.changeState(s.doSkipAllSpace(), at, s.doRecordSpace())
}

func (s *state) changeRecordSpace(rs uint8) {
	s.changeState(s.doSkipAllSpace(), s.getArgtype(), rs)
}
