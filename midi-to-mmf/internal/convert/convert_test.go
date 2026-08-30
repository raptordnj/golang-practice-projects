package convert

import (
	"testing"

	"github.com/aurnob/midi2mmf/internal/midi"
	"github.com/aurnob/midi2mmf/internal/smaf"
)

func songWith(t *testing.T, ppq int, tempoUS int64, evs ...midi.Event) *midi.Song {
	t.Helper()
	s := &midi.Song{TicksPerQN: ppq}
	s.TempoChanges = append(s.TempoChanges, midi.Event{Kind: midi.KTempo, A: int(tempoUS)})
	s.Events = append(s.Events, evs...)
	return s
}

func TestNotePairingAndGate(t *testing.T) {
	// 480ppq at 500000us/qn -> 1 tick == 1041.67us; use 96 ticks = 100ms.
	s := songWith(t, 480, 500000,
		midi.Event{Tick: 0, Kind: midi.KNoteOn, Channel: 0, A: 60, B: 100},
		midi.Event{Tick: 96, Kind: midi.KNoteOff, Channel: 0, A: 60},
	)
	f, st, err := ToSMAF(s, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if st.Notes != 1 {
		t.Fatalf("notes=%d", st.Notes)
	}
	var note *smaf.NoteEvent
	for _, p := range f.Pairs {
		if p.Event.Note != nil {
			note = p.Event.Note
		}
	}
	if note == nil {
		t.Fatal("no note in output")
	}
	// Whatever base auto-selection picks, the quantized gate must represent
	// the true 100ms window within one unit.
	gateUS := int64(note.Gate) * int64(st.TimeBaseGMS) * 1000
	if diff := gateUS - 100_000; diff < -int64(st.TimeBaseGMS)*1000 || diff > int64(st.TimeBaseGMS)*1000 {
		t.Errorf("gate=%d units of %dms -> %dus, want within one unit of 100000us",
			note.Gate, st.TimeBaseGMS, gateUS)
	}
}

func TestRetriggerClosesPreviousNote(t *testing.T) {
	// Same pitch struck twice; each off closes one instance (LIFO).
	s := songWith(t, 480, 500000,
		midi.Event{Tick: 0, Kind: midi.KNoteOn, Channel: 0, A: 60, B: 100},
		midi.Event{Tick: 48, Kind: midi.KNoteOn, Channel: 0, A: 60, B: 90},
		midi.Event{Tick: 72, Kind: midi.KNoteOff, Channel: 0, A: 60},  // closes the on@48
		midi.Event{Tick: 144, Kind: midi.KNoteOff, Channel: 0, A: 60}, // closes the on@0
	)
	f, st, err := ToSMAF(s, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if st.Notes != 2 {
		t.Fatalf("notes=%d", st.Notes)
	}
	var gates []int
	for _, p := range f.Pairs {
		if p.Event.Note != nil {
			gates = append(gates, p.Event.Note.Gate)
		}
	}
	// on@0 lasts 144 ticks (~150ms), on@48 lasts 24 ticks (~25ms): distinct.
	if len(gates) != 2 || gates[0] <= gates[1] {
		t.Errorf("gates=%v, want first (long) > second (short)", gates)
	}
}

func TestOrphanNoteOffIgnored(t *testing.T) {
	s := songWith(t, 480, 500000,
		midi.Event{Tick: 10, Kind: midi.KNoteOff, Channel: 1, A: 60},
	)
	f, st, err := ToSMAF(s, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if st.Notes != 0 {
		t.Errorf("orphan produced %d notes", st.Notes)
	}
	if len(f.Pairs) != 2 { // init sysex + EOT only
		t.Errorf("pairs=%d, want 2", len(f.Pairs))
	}
}

func TestUnclosedNoteRingsToEnd(t *testing.T) {
	s := songWith(t, 480, 500000,
		midi.Event{Tick: 0, Kind: midi.KNoteOn, Channel: 0, A: 60, B: 70},
		midi.Event{Tick: 480, Kind: midi.KProgChange, Channel: 5, A: 9},
	)
	_, st, err := ToSMAF(s, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if st.Notes != 1 {
		t.Fatalf("notes=%d", st.Notes)
	}
}

func TestTimeBaseSelectionExplicit(t *testing.T) {
	s := songWith(t, 480, 500000,
		midi.Event{Tick: 0, Kind: midi.KNoteOn, Channel: 0, A: 60, B: 100},
		midi.Event{Tick: 96, Kind: midi.KNoteOff, Channel: 0, A: 60},
	)
	f, _, err := ToSMAF(s, Options{TimeBaseDMS: 4})
	if err != nil {
		t.Fatal(err)
	}
	if f.TimeBaseD != 0x02 || f.TimeBaseG != 0x02 {
		t.Errorf("codes %#02x/%#02x, want 0x02 both", f.TimeBaseD, f.TimeBaseG)
	}
	if _, _, err := ToSMAF(s, Options{TimeBaseDMS: 3}); err == nil {
		t.Error("3ms is not a valid SMAF base and should error")
	}
}

func TestBankSelectDropped(t *testing.T) {
	s := songWith(t, 480, 500000,
		midi.Event{Tick: 0, Kind: midi.KCtrlChange, Channel: 0, A: 0, B: 3},  // bank MSB
		midi.Event{Tick: 0, Kind: midi.KCtrlChange, Channel: 0, A: 7, B: 90}, // volume kept
	)
	f, st, err := ToSMAF(s, Options{})
	if err != nil {
		t.Fatal(err)
	}
	ccs := 0
	for _, p := range f.Pairs {
		if p.Event.CC != nil {
			ccs++
			if p.Event.CC.CC == 0 {
				t.Error("bank select leaked into output")
			}
		}
	}
	if ccs != 1 || st.SkippedCCs != 1 {
		t.Errorf("ccs=%d skipped=%d, want 1/1", ccs, st.SkippedCCs)
	}
}

func TestEndToEndRoundTrip(t *testing.T) {
	// Build a small MIDI-like song, convert, encode, decode, verify content.
	s := songWith(t, 96, 500000,
		midi.Event{Tick: 0, Kind: midi.KProgChange, Channel: 0, A: 20},
		midi.Event{Tick: 0, Kind: midi.KNoteOn, Channel: 0, A: 72, B: 110},
		midi.Event{Tick: 24, Kind: midi.KNoteOn, Channel: 9, A: 36, B: 127},
		midi.Event{Tick: 24, Kind: midi.KNoteOff, Channel: 9, A: 36},
		midi.Event{Tick: 48, Kind: midi.KPitchBend, Channel: 0, A: 0x00, B: 0x50},
		midi.Event{Tick: 96, Kind: midi.KNoteOff, Channel: 0, A: 72},
	)
	f, st, err := ToSMAF(s, Options{Title: "Round Trip"})
	if err != nil {
		t.Fatal(err)
	}
	img, err := f.Encode()
	if err != nil {
		t.Fatal(err)
	}
	d, err := smaf.Decode(img)
	if err != nil {
		t.Fatal(err)
	}
	if !d.EndedClean || !d.CRCValid || d.Title != "Round Trip" {
		t.Fatalf("meta mismatch: ended=%v crc=%v title=%q", d.EndedClean, d.CRCValid, d.Title)
	}
	notes, pcs, bends := 0, 0, 0
	drumSeen := false
	for _, p := range d.Pairs {
		switch {
		case p.Event.Note != nil:
			notes++
			if p.Event.Note.Channel == 9 && p.Event.Note.Note == 36 {
				drumSeen = true
			}
		case p.Event.PC != nil:
			pcs++
		case p.Event.Bend != nil:
			bends++
			if p.Event.Bend.LSB != 0x00 || p.Event.Bend.MSB != 0x50 {
				t.Errorf("bend bytes %02x %02x", p.Event.Bend.LSB, p.Event.Bend.MSB)
			}
		}
	}
	if notes != 2 || pcs != 1 || bends != 1 || !drumSeen {
		t.Errorf("round trip: notes=%d pcs=%d bends=%d drum=%v (stats %+v)",
			notes, pcs, bends, drumSeen, st)
	}
}

func TestMonotonicDurations(t *testing.T) {
	// Many events across channels; every emitted duration must be >= 0 and the
	// timeline strictly accumulates.
	var evs []midi.Event
	for i := 0; i < 200; i++ {
		ch := i % 16
		evs = append(evs, midi.Event{
			Tick: int64(i * 3), Kind: midi.KNoteOn, Channel: ch, A: 40 + ch, B: 80,
		})
	}
	s := &midi.Song{TicksPerQN: 240}
	s.TempoChanges = append(s.TempoChanges, midi.Event{Kind: midi.KTempo, A: 400000})
	s.Events = evs
	f, _, err := ToSMAF(s, Options{})
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range f.Pairs {
		if p.Duration < 0 {
			t.Fatalf("negative duration %d", p.Duration)
		}
	}
}
