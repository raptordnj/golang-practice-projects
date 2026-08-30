// Package midi parses Standard MIDI Files (SMF) into absolute-time events.
package midi

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
)

// Event kinds emitted by the parser. Everything the SMAF score track can
// represent is kept; the rest is summarized as counts.
type Kind uint8

const (
	KNoteOn Kind = iota
	KNoteOff
	KCtrlChange
	KProgChange
	KPitchBend
	KTempo
)

// Event is a single parsed MIDI message with its absolute tick position.
type Event struct {
	Tick    int64
	Kind    Kind
	Channel int // 0-15; 0 for Tempo
	A       int // note number / controller number / bend LSB / tempo usec low
	B       int // velocity / controller value / bend MSB / tempo usec high
}

// Note, Velocity accessors keep call sites readable.
func (e Event) Note() int     { return e.A }
func (e Event) Velocity() int { return e.B }

// Song is a parsed MIDI file.
type Song struct {
	Format     int
	TicksPerQN int  // when IsSMPTE == false
	IsSMPTE    bool // SMPTE timing: TicksPerSecond is meaningful instead
	SMPTEFPS   int
	SMPTESub   int

	// TicksPerSecond is only set for SMPTE files.
	TicksPerSecond int

	TempoChanges []Event // KTempo events sorted by tick (tick 0 entry always present)
	Events       []Event // all channel events across tracks, sorted by (tick, insertion order)

	TruncatedTracks int // tracks abandoned mid-stream due to malformed data
}

var (
	ErrNotMIDI = errors.New("midi: not a Standard MIDI File")
	ErrTrunc   = errors.New("midi: unexpected end of file")
)

type reader struct {
	b   []byte
	pos int
}

func (r *reader) byte() (uint8, error) {
	if r.pos >= len(r.b) {
		return 0, ErrTrunc
	}
	v := r.b[r.pos]
	r.pos++
	return v, nil
}

func (r *reader) bytes(n int) ([]byte, error) {
	if n < 0 || r.pos+n > len(r.b) {
		return nil, ErrTrunc
	}
	v := r.b[r.pos : r.pos+n]
	r.pos += n
	return v, nil
}

func (r *reader) u16be() (int, error) {
	b, err := r.bytes(2)
	if err != nil {
		return 0, err
	}
	return int(b[0])<<8 | int(b[1]), nil
}

func (r *reader) u32be() (int64, error) {
	b, err := r.bytes(4)
	if err != nil {
		return 0, err
	}
	return int64(b[0])<<24 | int64(b[1])<<16 | int64(b[2])<<8 | int64(b[3]), nil
}

// varlen reads an SMF variable-length quantity.
func (r *reader) varlen() (int64, error) {
	var v int64
	for i := 0; i < 4; i++ {
		b, err := r.byte()
		if err != nil {
			return 0, err
		}
		v = v<<7 | int64(b&0x7F)
		if b&0x80 == 0 {
			return v, nil
		}
	}
	return 0, errors.New("midi: malformed variable-length quantity")
}

// Parse decodes a complete SMF byte stream.
func Parse(data []byte) (*Song, error) {
	r := &reader{b: data}
	magic, err := r.bytes(4)
	if err != nil || !bytes.Equal(magic, []byte("MThd")) {
		return nil, ErrNotMIDI
	}
	hlen, err := r.u32be()
	if err != nil {
		return nil, err
	}
	if hlen < 6 {
		return nil, fmt.Errorf("midi: bad header length %d", hlen)
	}
	format, err := r.u16be()
	if err != nil {
		return nil, err
	}
	ntrk, err := r.u16be()
	if err != nil {
		return nil, err
	}
	div, err := r.u16be()
	if err != nil {
		return nil, err
	}
	if _, err := r.bytes(int(hlen) - 6); err != nil { // skip any extra header bytes
		return nil, err
	}

	s := &Song{Format: format}
	if div&0x8000 != 0 {
		s.IsSMPTE = true
		s.SMPTEFPS = 256 - int(div>>8&0xFF) // stored as two's complement negative
		s.SMPTESub = int(div & 0xFF)
		s.TicksPerSecond = s.SMPTEFPS * s.SMPTESub
	} else {
		s.TicksPerQN = div
		if s.TicksPerQN == 0 {
			return nil, errors.New("midi: zero ticks per quarter note")
		}
	}

	for t := 0; t < ntrk; {
		id, err := r.bytes(4)
		if err != nil {
			return nil, err
		}
		if !bytes.Equal(id, []byte("MTrk")) {
			// Unknown chunk: skip it by its length and keep looking.
			clen, err := r.u32be()
			if err != nil {
				return nil, err
			}
			if _, err := r.bytes(int(clen)); err != nil {
				return nil, err
			}
			continue // retry this track index
		}
		if err := parseTrackBody(r, s); err != nil {
			return nil, fmt.Errorf("midi: track %d: %w", t, err)
		}
		t++
	}
	sort.SliceStable(s.Events, func(i, j int) bool { return s.Events[i].Tick < s.Events[j].Tick })
	if len(s.TempoChanges) == 0 && !s.IsSMPTE {
		s.TempoChanges = append(s.TempoChanges, Event{Tick: 0, Kind: KTempo, A: 500000})
	}
	sort.SliceStable(s.TempoChanges, func(i, j int) bool { return s.TempoChanges[i].Tick < s.TempoChanges[j].Tick })
	return s, nil
}

func parseTrackBody(r *reader, s *Song) error {
	tlen, err := r.u32be()
	if err != nil {
		return err
	}
	body, err := r.bytes(int(tlen))
	if err != nil {
		return err
	}
	tr := &reader{b: body}

	var tick int64
	running := -1
	insert := func(e Event) {
		e.Tick = tick
		s.Events = append(s.Events, e)
	}

	// A decode failure inside a track stops that track. Real-world writers
	// sometimes emit corrupt partial events and pad the chunk remainder with
	// 0xFF bytes; players cope by dropping the rest of the track, and so do we.
	restIsPadding := func() bool {
		for _, c := range tr.b[tr.pos:] {
			if c != 0xFF && c != 0x00 {
				return false
			}
		}
		return true
	}
	fail := func(err error) error {
		if restIsPadding() {
			return errTrackEnd
		}
		s.TruncatedTracks++
		return fmt.Errorf("%w (%v)", errTrackEnd, err)
	}

	step := func() error {
		delta, err := tr.varlen()
		if err != nil {
			return fail(err)
		}
		tick += delta

		status, err := tr.byte()
		if err != nil {
			return fail(err)
		}
		if status < 0x80 {
			if running < 0 {
				return fail(errors.New("midi: running status without preceding status byte"))
			}
			tr.pos-- // push back data byte
			status = uint8(running)
		} else {
			running = -1
		}

		switch {
		case status == 0xFF: // meta
			metaType, err := tr.byte()
			if err != nil {
				return fail(err)
			}
			mlen, err := tr.varlen()
			if err != nil {
				return fail(err)
			}
			mbody, err := tr.bytes(int(mlen))
			if err != nil {
				return fail(err)
			}
			switch metaType {
			case 0x51: // tempo
				if len(mbody) == 3 {
					us := int64(mbody[0])<<16 | int64(mbody[1])<<8 | int64(mbody[2])
					s.TempoChanges = append(s.TempoChanges, Event{Tick: tick, Kind: KTempo, A: int(us)})
				}
			case 0x2F: // end of track
				return errTrackEnd
			}
		case status == 0xF0 || status == 0xF7: // sysex
			slen, err := tr.varlen()
			if err != nil {
				return fail(err)
			}
			if _, err := tr.bytes(int(slen)); err != nil {
				return fail(err)
			}
			running = -1
		default:
			hi := status & 0xF0
			ch := int(status & 0x0F)
			switch hi {
			case 0xC0, 0xD0: // one data byte
				d, err := tr.byte()
				if err != nil || d > 0x7F {
					return fail(errInvalidStream)
				}
				running = int(status)
				if hi == 0xC0 {
					insert(Event{Kind: KProgChange, Channel: ch, A: int(d)})
				}
				// channel pressure (0xD0) has no score-track equivalent: dropped
			default: // two data bytes
				d1, err := tr.byte()
				if err != nil {
					return fail(err)
				}
				d2, err := tr.byte()
				if err != nil || d1 > 0x7F || d2 > 0x7F {
					return fail(errInvalidStream)
				}
				running = int(status)
				switch hi {
				case 0x80:
					insert(Event{Kind: KNoteOff, Channel: ch, A: int(d1)})
				case 0x90:
					if d2 == 0 {
						insert(Event{Kind: KNoteOff, Channel: ch, A: int(d1)})
					} else {
						insert(Event{Kind: KNoteOn, Channel: ch, A: int(d1), B: int(d2)})
					}
				case 0xA0: // poly pressure: no equivalent, dropped
				case 0xB0:
					insert(Event{Kind: KCtrlChange, Channel: ch, A: int(d1), B: int(d2)})
				case 0xE0:
					insert(Event{Kind: KPitchBend, Channel: ch, A: int(d1), B: int(d2)})
				}
			}
		}
		return nil
	}

	for tr.pos < len(tr.b) {
		if err := step(); err != nil {
			break // clean EOT, padding, or corruption: the track ends here
		}
	}
	return nil
}

var (
	errInvalidStream = errors.New("midi: data byte out of range (stream corruption)")
	errTrackEnd      = errors.New("midi: end of track")
)
