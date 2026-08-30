package smaf

import (
	"encoding/binary"
	"fmt"
)

// Decoded is the result of reading an SMAF score-track file back.
type Decoded struct {
	FileSize   int
	CRCValid   bool
	Chunks     []string // top-level chunk signatures in file order
	FormatType byte     // MTR format type (2 = Mobile Standard uncompressed)
	TimeBaseD  uint8
	TimeBaseG  uint8

	Title    string
	Artist   string
	Composer string

	Pairs      []Pair
	EndedClean bool // sequence terminated with FF 2F 00
}

type chunkRef struct {
	name    string
	size    int
	payload []byte
}

// Decode parses an .mmf image produced by this tool or by any other writer,
// as far as it uses the uncompressed Mobile Standard score format.
func Decode(data []byte) (*Decoded, error) {
	if len(data) < 10 {
		return nil, fmt.Errorf("smaf: file too short (%d bytes)", len(data))
	}
	if string(data[:4]) != "MMMD" {
		return nil, fmt.Errorf("smaf: missing MMMD signature")
	}
	d := &Decoded{FileSize: len(data)}
	want := binary.BigEndian.Uint16(data[len(data)-2:])
	got := CRC16(data[:len(data)-2])
	d.CRCValid = want == got

	declared := int(binary.BigEndian.Uint32(data[4:8]))
	if 8+declared > len(data) {
		return nil, fmt.Errorf("smaf: declared size %d exceeds file length %d", declared, len(data))
	}
	body := data[8 : 8+declared-2]

	var mtr *chunkRef
	pos := 0
	for pos+8 <= len(body) {
		c := chunkRef{name: string(body[pos : pos+4])}
		c.size = int(binary.BigEndian.Uint32(body[pos+4 : pos+8]))
		pos += 8
		if pos+c.size > len(body) {
			return nil, fmt.Errorf("smaf: chunk %q overruns body", c.name)
		}
		c.payload = body[pos : pos+c.size]
		pos += c.size
		d.Chunks = append(d.Chunks, c.name)
		switch c.name[:3] {
		case "CNT":
			if len(c.payload) > 5 {
				parseTextTags(c.payload[5:], d) // CNTI options are "ST:val,AN:val," style text
			}
		case "OPD":
			if err := decodeOPDA(c.payload, d); err != nil {
				return nil, err
			}
		case "MTR":
			mtr = &c
		default:
			// audio tracks and friends are skipped
		}
	}
	if mtr == nil {
		return nil, fmt.Errorf("smaf: no MTR score track chunk found")
	}
	return d, decodeMTR(*mtr, d)
}

func decodeOPDA(payload []byte, d *Decoded) error {
	pos := 0
	for pos+8 <= len(payload) {
		name := payload[pos : pos+4]
		size := int(binary.BigEndian.Uint32(payload[pos+4 : pos+8]))
		pos += 8
		if pos+size > len(payload) {
			return fmt.Errorf("smaf: OPDA sub-chunk overruns parent")
		}
		sub := payload[pos : pos+size]
		pos += size
		if string(name[:3]) == "Dch" {
			parseTags(sub, d) // code type rides in the name's 4th byte
		}
	}
	return nil
}

// parseTags reads Dch tag entries: 2-byte ASCII tag, u16be length, value.
func parseTags(b []byte, d *Decoded) {
	for len(b) >= 4 {
		tag := string(b[:2])
		n := int(b[2])<<8 | int(b[3])
		b = b[4:]
		if n > len(b) {
			return
		}
		setTag(tag, latinToUTF8(b[:n]), d)
		b = b[n:]
	}
}

// parseTextTags reads the comma-separated "ST:value,AN:value," text used in
// the CNTI chunk.
func parseTextTags(s []byte, d *Decoded) {
	for _, field := range splitCommas(string(s)) {
		for i := 0; i < len(field); i++ {
			if field[i] == ':' {
				setTag(field[:i], []byte(field[i+1:]), d)
				break
			}
		}
	}
}

// splitCommas splits on commas that are not escaped with a backslash.
func splitCommas(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\\':
			i++
		case ',':
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}

func setTag(tag string, val []byte, d *Decoded) {
	v := string(val)
	switch tag {
	case "ST":
		d.Title = v
	case "AN":
		d.Artist = v
	case "SW":
		d.Composer = v
	}
}

func latinToUTF8(b []byte) []byte {
	out := make([]byte, 0, len(b))
	for _, c := range b {
		out = append(out, string(rune(c))...)
	}
	return out
}

func decodeMTR(c chunkRef, d *Decoded) error {
	p := c.payload
	if len(p) < 20 {
		return fmt.Errorf("smaf: MTR chunk too short")
	}
	d.FormatType = p[0]
	switch d.FormatType {
	case 2: // Mobile Standard, uncompressed
	case 1:
		return fmt.Errorf("smaf: Huffman-compressed sequences are not supported")
	default:
		return fmt.Errorf("smaf: unsupported MTR format type %d", d.FormatType)
	}
	d.TimeBaseD = p[2]
	d.TimeBaseG = p[3]

	pos := 20 // 4-byte header + 16 channel-status bytes
	for pos+8 <= len(p) {
		name := string(p[pos : pos+4])
		size := int(binary.BigEndian.Uint32(p[pos+4 : pos+8]))
		pos += 8
		if pos+size > len(p) {
			return fmt.Errorf("smaf: MTR sub-chunk %q overruns parent", name)
		}
		sub := p[pos : pos+size]
		pos += size
		if name == "Mtsq" {
			if err := decodeSequence(sub, d); err != nil {
				return err
			}
		}
		// Mtsu (setup), Mtsp (stream PCM) and friends carry no score events.
	}
	return nil
}

func decodeSequence(b []byte, d *Decoded) error {
	lastVel := [16]int{64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64, 64}
	pos := 0
	readVar := func() (int, error) {
		var v int
		for i := 0; i < 3; i++ {
			if pos >= len(b) {
				return 0, fmt.Errorf("smaf: truncated variable-length quantity at %d", pos)
			}
			c := int(b[pos])
			pos++
			v = v<<7 | c&0x7F
			if c&0x80 == 0 {
				return v, nil
			}
		}
		return 0, fmt.Errorf("smaf: malformed variable-length quantity at %d", pos)
	}
	need := func(n int) error {
		if pos+n > len(b) {
			return fmt.Errorf("smaf: truncated event at offset %d", pos)
		}
		return nil
	}
	for pos < len(b) {
		dur, err := readVar()
		if err != nil {
			return err
		}
		if err := need(1); err != nil {
			return err
		}
		status := int(b[pos])
		pos++
		ch := status & 0x0F
		if status == 0xFF { // meta; checked first — masked below, 0xFF would land in 0xF0
			meta := int(b[pos])
			pos++
			switch meta {
			case 0x00: // NOP
			case 0x2F:
				if err := need(1); err != nil {
					return err
				}
				pos++
				d.EndedClean = true
				return nil
			default:
				return fmt.Errorf("smaf: unknown meta %#02x at %d", meta, pos-1)
			}
			continue
		}
		switch status & 0xF0 {
		case 0x80, 0x90:
			if err := need(1); err != nil {
				return err
			}
			note := int(b[pos])
			pos++
			vel := lastVel[ch]
			explicit := false
			if status&0xF0 == 0x90 {
				if err := need(1); err != nil {
					return err
				}
				vel = int(b[pos])
				pos++
				lastVel[ch] = vel
				explicit = true
			}
			gate, err := readVar()
			if err != nil {
				return err
			}
			d.Pairs = append(d.Pairs, Pair{dur, Event{Note: &NoteEvent{
				Channel: ch, Note: note, Velocity: vel, Gate: gate, ExplicitVel: explicit,
			}}})
		case 0xB0:
			if err := need(2); err != nil {
				return err
			}
			cc := int(b[pos])
			val := int(b[pos+1])
			pos += 2
			d.Pairs = append(d.Pairs, Pair{dur, Event{CC: &CCEvent{ch, cc, val}}})
		case 0xC0:
			if err := need(1); err != nil {
				return err
			}
			pc := int(b[pos])
			pos++
			d.Pairs = append(d.Pairs, Pair{dur, Event{PC: &PCEvent{ch, pc}}})
		case 0xE0:
			if err := need(2); err != nil {
				return err
			}
			lsb := int(b[pos])
			msb := int(b[pos+1])
			pos += 2
			d.Pairs = append(d.Pairs, Pair{dur, Event{Bend: &BendEvent{ch, lsb, msb}}})
		case 0xF0:
			l, err := readVar()
			if err != nil {
				return err
			}
			if l < 1 {
				return fmt.Errorf("smaf: bad sysex length at %d", pos)
			}
			if err := need(l); err != nil {
				return err
			}
			payload := make([]byte, 0, l-1)
			payload = append(payload, b[pos:pos+l-1]...)
			pos += l
			if b[pos-1] != 0xF7 {
				return fmt.Errorf("smaf: sysex missing F7 terminator at %d", pos-1)
			}
			d.Pairs = append(d.Pairs, Pair{dur, Event{SysEx: payload}})
		}
	}
	return nil
}
