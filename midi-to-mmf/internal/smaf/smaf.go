// Package smaf encodes and decodes Yamaha SMAF (.mmf) files with a score
// track (MA-2/MA-3 "Mobile Standard", uncompressed).
//
// Layout reference: two real-world ringtone files were used as ground truth,
// cross-checked against the go-smaf parser (github.com/but80/go-smaf) and
// Rockbox's smaf metadata parser.
package smaf

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// TimeBase codes for the MTR header. Each code selects a base length in
// milliseconds used to interpret Duration (D) and GateTime (G) values in the
// sequence data.
var TimeBaseMS = map[uint8]int{
	0x00: 1,
	0x01: 2,
	0x02: 4,
	0x03: 5,
	0x10: 10,
	0x11: 20,
	0x12: 40,
	0x13: 50,
}

// InitSysex is the Yamaha sound-chip initialization exclusive found at the
// start of every score sequence in the reference files:
//
//	F0 07 43 79 07 7F 07 01 F7
var InitSysex = []byte{0x43, 0x79, 0x07, 0x7F, 0x07, 0x01}

// File is an SMAF score-track file ready to encode.
type File struct {
	Title    string // written into OPDA/Dch as tag "ST" when non-empty
	Artist   string // tag "AN"
	Composer string // tag "SW"

	TimeBaseD uint8 // MTR Duration base code; see TimeBaseMS
	TimeBaseG uint8 // MTR GateTime base code

	Pairs []Pair // ordered sequence data; must end with EndOfTrack()
}

// Pair is one duration+event step of the sequence data. Duration is expressed
// in TimeBaseD units and is the silence that precedes Event.
type Pair struct {
	Duration int
	Event    Event
}

// Event is one score-track message. Exactly one of the pointers is set.
type Event struct {
	Note  *NoteEvent
	CC    *CCEvent
	PC    *PCEvent
	Bend  *BendEvent
	SysEx []byte // raw payload between F0 ... F7 (terminator added by encoder)
	End   bool   // end-of-sequence marker (FF 2F 00)
}

// NoteEvent starts a note on a channel; it sounds for Gate units of
// TimeBaseG. When ExplicitVel is false the encoder emits the compact form
// (status 0x80|ch) which reuses the channel's previously sent velocity.
type NoteEvent struct {
	Channel     int
	Note        int
	Velocity    int
	Gate        int
	ExplicitVel bool
}

// CCEvent passes a MIDI control change through to the synthesizer channel.
type CCEvent struct {
	Channel, CC, Value int
}

// PCEvent selects a preset tone for the channel.
type PCEvent struct{ Channel, Program int }

// BendEvent carries a 14-bit pitch bend as raw LSB/MSB bytes (MIDI order).
type BendEvent struct{ Channel, LSB, MSB int }

// EndOfTrack returns the terminating event for a sequence.
func EndOfTrack() Event { return Event{End: true} }

// --- encoding -------------------------------------------------------------

func putChunk(dst []byte, name string, payload []byte) []byte {
	dst = append(dst, name...)
	var l [4]byte
	binary.BigEndian.PutUint32(l[:], uint32(len(payload)))
	dst = append(dst, l[:]...)
	return append(dst, payload...)
}

// VarLen encodes v as an SMAF/MIDI variable-length quantity (max 3 bytes).
func VarLen(v int) ([]byte, error) {
	if v < 0 || v >= 1<<21 {
		return nil, fmt.Errorf("smaf: value %d out of range for variable-length quantity", v)
	}
	var buf [5]byte
	i := len(buf)
	buf[i-1] = byte(v & 0x7F)
	v >>= 7
	for v > 0 {
		i--
		buf[i-1] = byte(v&0x7F) | 0x80
		v >>= 7
	}
	return buf[i-1:], nil
}

// Encode builds the complete .mmf file image.
func (f *File) Encode() ([]byte, error) {
	body, err := f.encodeBody()
	if err != nil {
		return nil, err
	}
	out := make([]byte, 0, len(body)+16)
	out = append(out, "MMMD"...)
	var l [4]byte
	binary.BigEndian.PutUint32(l[:], uint32(len(body)+2)) // + trailing CRC
	out = append(out, l[:]...)
	out = append(out, body...)
	crc := CRC16(out)
	return append(out, byte(crc>>8), byte(crc)), nil
}

func (f *File) encodeBody() ([]byte, error) {
	if f.TimeBaseD == 0 && f.TimeBaseG == 0 {
		f.TimeBaseD, f.TimeBaseG = 0x02, 0x02 // 4 ms, matching reference files
	}
	if _, ok := TimeBaseMS[f.TimeBaseD]; !ok {
		return nil, fmt.Errorf("smaf: invalid TimeBaseD code %#02x", f.TimeBaseD)
	}
	if _, ok := TimeBaseMS[f.TimeBaseG]; !ok {
		return nil, fmt.Errorf("smaf: invalid TimeBaseG code %#02x", f.TimeBaseG)
	}

	// CNTI — contents info. Byte values copied from the reference files:
	// class 0x00, type 0x34, code type 0x01 (Latin-1), copy status/count.
	cnti := []byte{0x00, 0x34, 0x01, 0xFD, 0x00}

	// OPDA — optional data: one Dch chunk holding text tags.
	var entries []byte
	addTag := func(tag, val string) {
		if val == "" {
			return
		}
		entries = append(entries, tag...)
		v := latin1(val)
		entries = append(entries, byte(len(v)>>8), byte(len(v)))
		entries = append(entries, v...)
	}
	addTag("ST", f.Title)
	addTag("AN", f.Artist)
	addTag("SW", f.Composer)

	// MTR — score track.
	mtr := []byte{0x02, 0x00, f.TimeBaseD, f.TimeBaseG} // format 2 (Mobile Std, uncompressed), stream sequence
	mtr = append(mtr, make([]byte, 16)...)              // per-channel status: all zero, like the references

	mtsu := []byte{0xF0, 0x07}
	mtsu = append(mtsu, InitSysex...)
	mtsu = append(mtsu, 0xF7)

	mtsq, err := encodeSequence(f.Pairs)
	if err != nil {
		return nil, err
	}

	mtr = putChunk(mtr, "Mtsu", mtsu)
	mtr = putChunk(mtr, "Mtsq", mtsq)

	// Chunk order matches the reference files: CNTI, OPDA, MTR.
	body := putChunk(nil, "CNTI", cnti)
	if len(entries) > 0 {
		dch := append([]byte{'D', 'c', 'h', 0x01}, 0, 0, 0, 0) // name + size placeholder
		binary.BigEndian.PutUint32(dch[4:], uint32(len(entries)))
		dch = append(dch, entries...)
		body = append(body, putChunk(nil, "OPDA", dch)...)
	}
	body = append(body, putChunk(nil, "MTR\x06", mtr)...)
	return body, nil
}

func encodeSequence(pairs []Pair) ([]byte, error) {
	out := make([]byte, 0, 256)
	lastVel := [16]int{-1}
	for _, p := range pairs {
		d, err := VarLen(p.Duration)
		if err != nil {
			return nil, fmt.Errorf("smaf: duration: %w", err)
		}
		out = append(out, d...)

		switch {
		case p.Event.Note != nil:
			n := p.Event.Note
			if n.Channel < 0 || n.Channel > 15 || n.Note < 0 || n.Note > 127 {
				return nil, fmt.Errorf("smaf: bad note event %+v", n)
			}
			explicit := n.ExplicitVel || lastVel[n.Channel] != n.Velocity
			status := byte(0x80 | n.Channel)
			if explicit {
				status = byte(0x90 | n.Channel)
			}
			out = append(out, status, byte(n.Note))
			if explicit {
				out = append(out, byte(n.Velocity))
				lastVel[n.Channel] = n.Velocity
			}
			g, err := VarLen(n.Gate)
			if err != nil {
				return nil, fmt.Errorf("smaf: gate time: %w", err)
			}
			out = append(out, g...)
		case p.Event.CC != nil:
			c := p.Event.CC
			out = append(out, byte(0xB0|c.Channel), byte(c.CC), byte(c.Value))
		case p.Event.PC != nil:
			p2 := p.Event.PC
			out = append(out, byte(0xC0|p2.Channel), byte(p2.Program))
		case p.Event.Bend != nil:
			b := p.Event.Bend
			out = append(out, byte(0xE0|b.Channel), byte(b.LSB), byte(b.MSB))
		case p.Event.SysEx != nil:
			l, err := VarLen(len(p.Event.SysEx) + 1) // length counts payload + F7
			if err != nil {
				return nil, fmt.Errorf("smaf: sysex length: %w", err)
			}
			out = append(out, 0xF0)
			out = append(out, l...)
			out = append(out, p.Event.SysEx...)
			out = append(out, 0xF7)
		case p.Event.End:
			out = append(out, 0xFF, 0x2F, 0x00)
		default:
			return nil, errors.New("smaf: empty event")
		}
	}
	return out, nil
}

// latin1 maps a string into Latin-1 bytes; runes above U+00FF become '?'.
func latin1(s string) []byte {
	out := make([]byte, 0, len(s))
	for _, r := range s {
		switch {
		case r <= 0xFF:
			out = append(out, byte(r))
		case r <= 0xFFFF:
			out = append(out, '?')
		default:
			out = append(out, '?', '?')
		}
	}
	return out
}

// CRC16 computes CRC-16/CCITT-FALSE (poly 0x1021, init/xorout 0xFFFF).
// The value is appended big-endian at the end of the file and covers every
// preceding byte.
func CRC16(data []byte) uint16 {
	crc := uint16(0xFFFF)
	for _, b := range data {
		crc ^= uint16(b) << 8
		for i := 0; i < 8; i++ {
			if crc&0x8000 != 0 {
				crc = crc<<1 ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc ^ 0xFFFF
}
