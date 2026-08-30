package smaf

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestVarLen(t *testing.T) {
	cases := []struct {
		v    int
		want []byte
	}{
		{0, []byte{0x00}},
		{1, []byte{0x01}},
		{127, []byte{0x7F}},
		{128, []byte{0x81, 0x00}},
		{16383, []byte{0xFF, 0x7F}},
		{16384, []byte{0x81, 0x80, 0x00}},
		{(1 << 21) - 1, []byte{0xFF, 0xFF, 0x7F}},
	}
	for _, c := range cases {
		got, err := VarLen(c.v)
		if err != nil {
			t.Fatalf("VarLen(%d): %v", c.v, err)
		}
		if !bytes.Equal(got, c.want) {
			t.Errorf("VarLen(%d) = % x, want % x", c.v, got, c.want)
		}
	}
	if _, err := VarLen(1 << 21); err == nil {
		t.Error("VarLen(2^21) should overflow")
	}
	if _, err := VarLen(-1); err == nil {
		t.Error("VarLen(-1) should be rejected")
	}
}

func TestCRC16(t *testing.T) {
	// SMAF files carry CRC-CCITT (init 0xFFFF, poly 0x1021) WITH a final
	// inversion — verified against two real-world ringtone files. Note the
	// textbook CCITT-FALSE check value 0x29B1 is the no-final-xor variant.
	if got := CRC16([]byte("123456789")); got != 0xD64E {
		t.Errorf("CRC16(check) = %#04x, want 0xd64e", got)
	}
	if got := CRC16([]byte("123456789")) ^ 0xFFFF; got != 0x29B1 {
		t.Errorf("CRC16 pre-xor = %#04x, want 0x29b1", got)
	}
}

func TestEncodeHeaderAndCRC(t *testing.T) {
	f := &File{
		TimeBaseD: 0x02,
		TimeBaseG: 0x02,
		Pairs: []Pair{
			{0, Event{PC: &PCEvent{Channel: 0, Program: 5}}},
			{10, Event{Note: &NoteEvent{Channel: 0, Note: 60, Velocity: 100, Gate: 20, ExplicitVel: true}}},
			{0, EndOfTrack()},
		},
	}
	img, err := f.Encode()
	if err != nil {
		t.Fatal(err)
	}
	if string(img[:4]) != "MMMD" {
		t.Errorf("missing MMMD signature: % x", img[:4])
	}
	declared := binary.BigEndian.Uint32(img[4:8])
	if int(declared) != len(img)-8 {
		t.Errorf("declared size %d != len-8 %d", declared, len(img)-8)
	}
	if string(img[8:12]) != "CNTI" {
		t.Errorf("first chunk should be CNTI, got %q", img[8:12])
	}
	stored := binary.BigEndian.Uint16(img[len(img)-2:])
	if stored != CRC16(img[:len(img)-2]) {
		t.Error("trailing CRC does not match computed CRC")
	}

	d, err := Decode(img)
	if err != nil {
		t.Fatal(err)
	}
	if !d.CRCValid || !d.EndedClean {
		t.Errorf("CRCValid=%v EndedClean=%v", d.CRCValid, d.EndedClean)
	}
	if d.TimeBaseD != 0x02 || d.TimeBaseG != 0x02 {
		t.Errorf("time bases not preserved: %#02x/%#02x", d.TimeBaseD, d.TimeBaseG)
	}
	if len(d.Pairs) != 2 { // EOT is consumed by the decoder, not appended
		t.Fatalf("got %d pairs, want 2", len(d.Pairs))
	}
	if d.Pairs[1].Event.Note == nil || d.Pairs[1].Event.Note.Note != 60 ||
		d.Pairs[1].Event.Note.Velocity != 100 || d.Pairs[1].Event.Note.Gate != 20 {
		t.Errorf("note round-trip mismatch: %+v", d.Pairs[1].Event.Note)
	}
}

func TestVelocityReuseEncoding(t *testing.T) {
	f := &File{
		Pairs: []Pair{
			{0, Event{Note: &NoteEvent{Channel: 3, Note: 64, Velocity: 90, Gate: 5}}},
			{0, Event{Note: &NoteEvent{Channel: 3, Note: 66, Velocity: 90, Gate: 5}}}, // same vel -> compact form
			{0, Event{Note: &NoteEvent{Channel: 3, Note: 68, Velocity: 80, Gate: 5}}}, // new vel -> explicit
			{0, EndOfTrack()},
		},
	}
	img, err := f.Encode()
	if err != nil {
		t.Fatal(err)
	}
	mtsq := bytes.Index(img, []byte("Mtsq"))
	size := int(binary.BigEndian.Uint32(img[mtsq+4 : mtsq+8]))
	seq := img[mtsq+8 : mtsq+8+size]
	// three notes then EOT; velocity reuse drops the vel byte on note 2
	want := []byte{
		0x00, 0x93, 0x40, 90, 0x05, // explicit velocity
		0x00, 0x83, 0x42, 0x05, // reused velocity (compact form)
		0x00, 0x93, 0x44, 80, 0x05, // changed velocity
		0x00, 0xFF, 0x2F, 0x00,
	}
	if !bytes.Equal(seq, want) {
		t.Errorf("sequence bytes:\n got % x\nwant % x", seq, want)
	}
}

func TestDecodeGroundTruthStructure(t *testing.T) {
	// A hand-built minimal file mirroring the reference layout: MTR header,
	// Mtsu setup chunk and a tiny sequence.
	f := &File{
		Title: "Test",
		Pairs: []Pair{
			{0, Event{SysEx: InitSysex}},
			{250, Event{Note: &NoteEvent{Channel: 9, Note: 38, Velocity: 105, Gate: 30, ExplicitVel: true}}},
			{0, EndOfTrack()},
		},
	}
	img, err := f.Encode()
	if err != nil {
		t.Fatal(err)
	}
	d, err := Decode(img)
	if err != nil {
		t.Fatal(err)
	}
	if d.Title != "Test" {
		t.Errorf("title %q", d.Title)
	}
	if d.FormatType != 2 {
		t.Errorf("format type %d", d.FormatType)
	}
	for _, name := range []string{"CNTI", "OPDA", "MTR\x06"} {
		found := false
		for _, c := range d.Chunks {
			if c == name {
				found = true
			}
		}
		if !found {
			t.Errorf("chunk %q missing from %q", name, d.Chunks)
		}
	}
}

func TestLatin1(t *testing.T) {
	got := latin1("Café • Delvi")
	want := []byte("Caf\xe9 ? Delvi")
	if !bytes.Equal(got, want) {
		t.Errorf("latin1 = % x, want % x", got, want)
	}
}
