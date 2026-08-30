package midi

import (
	"bytes"
	"testing"
)

// buildSMF assembles a format-0 SMF from raw track bytes.
func buildSMF(t *testing.T, ppq int, tracks ...[]byte) []byte {
	t.Helper()
	out := []byte{'M', 'T', 'h', 'd'}
	put32 := func(v uint32) { out = append(out, byte(v>>24), byte(v>>16), byte(v>>8), byte(v)) }
	put16 := func(v uint16) { out = append(out, byte(v>>8), byte(v)) }
	put32(6)
	put16(0)
	put16(uint16(len(tracks)))
	put16(uint16(ppq))
	for _, tr := range tracks {
		out = append(out, 'M', 'T', 'r', 'k')
		put32(uint32(len(tr)))
		out = append(out, tr...)
	}
	return out
}

var varlenTests = []struct {
	v    int64
	want []byte
}{
	{0, []byte{0}},
	{127, []byte{0x7F}},
	{128, []byte{0x81, 0x00}},
	{16383, []byte{0xFF, 0x7F}},
}

func TestVarLenDecode(t *testing.T) {
	for _, c := range varlenTests {
		r := &reader{b: c.want}
		got, err := r.varlen()
		if err != nil || got != c.v {
			t.Errorf("varlen(% x) = %d, %v; want %d", c.want, got, err, c.v)
		}
	}
}

func TestParseBasic(t *testing.T) {
	trk := []byte{
		0x00, 0xFF, 0x51, 0x03, 0x07, 0xA1, 0x20, // tempo 500000
		0x00, 0xC0, 0x10, // PC ch0 #16
		0x00, 0x90, 0x3C, 0x64, // note on C4
		0x60, 0x80, 0x3C, 0x40, // note off after 96 ticks
		0x00, 0xB0, 0x07, 0x7F, // CC7
		0x00, 0xE0, 0x00, 0x40, // pitch bend center
		0x00, 0xFF, 0x2F, 0x00,
	}
	song, err := Parse(buildSMF(t, 480, trk))
	if err != nil {
		t.Fatal(err)
	}
	if song.TicksPerQN != 480 || len(song.TempoChanges) != 1 || song.TempoChanges[0].A != 500000 {
		t.Fatalf("header/tempo mismatch: %+v", song)
	}
	if len(song.Events) != 5 {
		t.Fatalf("got %d events, want 5", len(song.Events))
	}
	kinds := []Kind{KProgChange, KNoteOn, KNoteOff, KCtrlChange, KPitchBend}
	for i, k := range kinds {
		if song.Events[i].Kind != k {
			t.Errorf("event %d kind %d, want %d", i, song.Events[i].Kind, k)
		}
	}
	if song.Events[3].Tick != 96 { // note-off and CC share tick 96
		t.Errorf("CC tick %d", song.Events[3].Tick)
	}
	if us := song.TimeMicros(96); us != 100_000 { // 96 * 500000/480
		t.Errorf("TimeMicros(96)=%d", us)
	}
}

func TestRunningStatusAndVelocityZero(t *testing.T) {
	trk := []byte{
		0x00, 0x90, 0x3C, 0x40, // on C4
		0x30, 0x3E, 0x40, // running status: on D4 (no repeated 0x90)
		0x30, 0x3C, 0x00, // running status with vel 0 -> note OFF
		0x00, 0xFF, 0x2F, 0x00,
	}
	song, err := Parse(buildSMF(t, 96, trk))
	if err != nil {
		t.Fatal(err)
	}
	var ons, offs int
	for _, e := range song.Events {
		switch e.Kind {
		case KNoteOn:
			ons++
			if e.Note() != 0x3C && e.Note() != 0x3E {
				t.Errorf("unexpected note %d", e.Note())
			}
		case KNoteOff:
			offs++
			if e.Tick != 96 {
				t.Errorf("vel-0 off tick %d", e.Tick)
			}
		}
	}
	if ons != 2 || offs != 1 {
		t.Errorf("ons=%d offs=%d", ons, offs)
	}
}

func TestTempoChangeIntegration(t *testing.T) {
	trk := []byte{
		0x00, 0xFF, 0x51, 0x03, 0x07, 0xA1, 0x20, // 500000 us/qn
		0x00, 0x90, 0x3C, 0x64,
		0x60, 0xFF, 0x51, 0x03, 0x03, 0xD0, 0x90, // at tick 96: 250000 us/qn (tempo doubles)
		0x60, 0x80, 0x3C, 0x40, // off at tick 192
		0x00, 0xFF, 0x2F, 0x00,
	}
	song, err := Parse(buildSMF(t, 480, trk))
	if err != nil {
		t.Fatal(err)
	}
	off := song.TimeMicros(192)
	// exact accumulation then one division: (96*500000 + 96*250000) / 480
	want := (int64(96)*500000 + int64(96)*250000) / 480
	if off != want {
		t.Errorf("TimeMicros(192)=%d want %d", off, want)
	}
}

func TestSMPTE(t *testing.T) {
	trk := []byte{
		0x00, 0x90, 0x3C, 0x40,
		0x64, 0x80, 0x3C, 0x40, // 100 ticks later
		0x00, 0xFF, 0x2F, 0x00,
	}
	body := buildSMF(t, 96, trk)
	// overwrite division word with SMPTE: -30 fps (0xE2), 100 subframes
	body[12] = 0xE2
	body[13] = 100
	song, err := Parse(body)
	if err != nil {
		t.Fatal(err)
	}
	if !song.IsSMPTE || song.TicksPerSecond != 3000 {
		t.Fatalf("SMPTE parse wrong: %+v", song)
	}
	if us := song.TimeMicros(150); us != 50_000 { // 150 ticks at 3000 tps
		t.Errorf("TimeMicros(150)=%d want 50000", us)
	}
}

func TestCorruptTrackIsTruncatedNotFatal(t *testing.T) {
	trk := []byte{
		0x00, 0x90, 0x3C, 0x40, // valid note
		0x10, 0xA3, 0x7C, 0x89, // corrupt event: data byte 0x89 > 0x7F
		0xFF, 0xFF, 0xFF, // padding
		0x00, 0xFF, 0x2F, 0x00,
	}
	song, err := Parse(buildSMF(t, 96, trk))
	if err != nil {
		t.Fatalf("corrupt track should not fail parse: %v", err)
	}
	if song.TruncatedTracks != 1 {
		t.Errorf("TruncatedTracks=%d, want 1", song.TruncatedTracks)
	}
	if len(song.Events) < 1 {
		t.Error("valid events before corruption were lost")
	}
}

func TestUnknownChunkSkipped(t *testing.T) {
	body := buildSMF(t, 96, []byte{0x00, 0xFF, 0x2F, 0x00})
	hdr := body[:14]
	trk := body[14:]
	var out bytes.Buffer
	out.Write(hdr)
	junk := append([]byte{'J', 'U', 'N', 'K', 0x00, 0x00, 0x00, 0x04, 1, 2, 3, 4}, trk...)
	out.Write(junk)
	if _, err := Parse(out.Bytes()); err != nil {
		t.Fatalf("unknown chunk should be skipped: %v", err)
	}
}

func TestGarbageInputNeverPanics(t *testing.T) {
	// Deterministic pseudo-random byte soup: parser must return an error or a
	// partial song, never panic.
	seed := uint32(0x12345678)
	next := func() byte {
		seed = seed*1664525 + 1013904223
		return byte(seed >> 24)
	}
	for round := 0; round < 500; round++ {
		n := int(next()%64) + 8
		buf := make([]byte, n)
		for i := range buf {
			buf[i] = next()
		}
		if round%3 == 0 {
			copy(buf, "MThd") // sometimes well-formed start, corrupt later
		}
		func() {
			defer func() {
				if p := recover(); p != nil {
					t.Fatalf("panic on input % x: %v", buf, p)
				}
			}()
			_, _ = Parse(buf)
		}()
	}
}
