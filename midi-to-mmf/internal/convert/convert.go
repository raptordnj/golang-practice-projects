// Package convert turns a parsed MIDI song into an SMAF score track.
//
// SMAF score timing is metric: durations count units of TimeBaseD
// milliseconds, so the MIDI tempo map is integrated into absolute times and
// quantized onto the SMAF tick grid. Notes become single events carrying a
// gate time; there are no note-off messages.
package convert

import (
	"fmt"

	"github.com/aurnob/midi2mmf/internal/midi"
	"github.com/aurnob/midi2mmf/internal/smaf"
)

// Options tunes the conversion.
type Options struct {
	Title    string
	Artist   string
	Composer string

	TimeBaseDMS int // Duration base in ms; 0 = auto-select
	TimeBaseGMS int // GateTime base in ms; 0 = same as D
	MaxErrorMS  int // auto-selection tolerance (default 4)
}

// Stats summarizes what a conversion produced.
type Stats struct {
	Notes        int
	CCs          int
	PCs          int
	Bends        int
	SkippedCCs   int
	ChannelsUsed [16]bool
	TimeBaseDMS  int
	TimeBaseGMS  int
	LengthMS     int64
}

// timeBaseCodes ordered coarsest first for auto-selection.
var timeBaseCodes = []struct {
	code uint8
	ms   int
}{
	{0x13, 50}, {0x12, 40}, {0x11, 20}, {0x10, 10},
	{0x03, 5}, {0x02, 4}, {0x01, 2}, {0x00, 1},
}

func codeForMS(ms int) (uint8, error) {
	for _, tb := range timeBaseCodes {
		if tb.ms == ms {
			return tb.code, nil
		}
	}
	return 0, fmt.Errorf("convert: no TimeBase code for %d ms (valid: 1, 2, 4, 5, 10, 20, 40, 50)", ms)
}

// scoreItem is one pending output event with its absolute position.
type scoreItem struct {
	us   int64
	item smaf.Event
}

// ToSMAF converts a parsed MIDI song into an SMAF file model.
func ToSMAF(song *midi.Song, opt Options) (*smaf.File, Stats, error) {
	st := Stats{TimeBaseDMS: -1, TimeBaseGMS: -1}
	if opt.MaxErrorMS <= 0 {
		opt.MaxErrorMS = 4
	}

	// --- pair notes into windows -------------------------------------
	pending := map[[2]int][]int{} // (channel,note) -> indexes into items
	var items []scoreItem
	addItem := func(us int64, ev smaf.Event) int {
		items = append(items, scoreItem{us: us, item: ev})
		return len(items) - 1
	}

	noteOffUs := make(map[int]int64) // item index -> off time
	closeNote := func(ch, note, velOff int, us int64) bool {
		key := [2]int{ch, note}
		stack := pending[key]
		if len(stack) == 0 {
			return false // orphan note-off
		}
		idx := stack[len(stack)-1]
		pending[key] = stack[:len(stack)-1]
		noteOffUs[idx] = us
		return true
	}

	var skippedCCs int
	droppedCC := map[int]bool{
		0: true, 32: true, // bank select — MA presets use PC only
		6: true, 38: true, // data entry (RPN payload)
		98: true, 99: true, 100: true, 101: true, // NRPN/RPN selectors
		120: true, 121: true, 123: true, // all-sound/reset/notes-off have no score equivalent
		126: true, 127: true, // mono/poly mode
	}

	for _, e := range song.Events {
		us := song.TimeMicros(e.Tick)
		switch e.Kind {
		case midi.KNoteOn:
			idx := addItem(us, smaf.Event{Note: &smaf.NoteEvent{
				Channel: e.Channel, Note: e.Note(), Velocity: e.Velocity(),
			}})
			key := [2]int{e.Channel, e.Note()}
			pending[key] = append(pending[key], idx)
			st.ChannelsUsed[e.Channel] = true
		case midi.KNoteOff:
			closeNote(e.Channel, e.A, e.B, us)
		case midi.KCtrlChange:
			if droppedCC[e.A] {
				skippedCCs++
				continue
			}
			addItem(us, smaf.Event{CC: &smaf.CCEvent{Channel: e.Channel, CC: e.A, Value: e.B}})
			st.CCs++
			st.ChannelsUsed[e.Channel] = true
		case midi.KProgChange:
			addItem(us, smaf.Event{PC: &smaf.PCEvent{Channel: e.Channel, Program: e.A}})
			st.PCs++
			st.ChannelsUsed[e.Channel] = true
		case midi.KPitchBend:
			addItem(us, smaf.Event{Bend: &smaf.BendEvent{Channel: e.Channel, LSB: e.A, MSB: e.B}})
			st.Bends++
			st.ChannelsUsed[e.Channel] = true
		}
	}

	endUs := song.TimeMicros(song.TotalTicks())
	for _, stack := range pending {
		for _, idx := range stack {
			noteOffUs[idx] = endUs // note never released: ring until end
		}
	}

	// --- choose time bases -------------------------------------------
	dMS, gMS := opt.TimeBaseDMS, opt.TimeBaseGMS
	if dMS <= 0 {
		dMS = selectBase(items, noteOffUs, endUs, int64(opt.MaxErrorMS)*1000)
	}
	if gMS <= 0 {
		gMS = dMS
	}
	dCode, err := codeForMS(dMS)
	if err != nil {
		return nil, st, err
	}
	gCode, err := codeForMS(gMS)
	if err != nil {
		return nil, st, err
	}
	st.TimeBaseDMS, st.TimeBaseGMS = dMS, gMS

	// --- lay out pairs ------------------------------------------------
	tickAt := func(us int64) int {
		v := (us + int64(dMS)*500) / (int64(dMS) * 1000) // round to nearest
		if v < 0 {
			v = 0
		}
		if v > (1<<21)-1 {
			v = (1 << 21) - 1
		}
		return int(v)
	}

	f := &smaf.File{
		Title:     opt.Title,
		Artist:    opt.Artist,
		Composer:  opt.Composer,
		TimeBaseD: dCode,
		TimeBaseG: gCode,
	}
	f.Pairs = append(f.Pairs, smaf.Pair{Duration: 0, Event: smaf.Event{SysEx: smaf.InitSysex}})

	prevTick := 0
	maxTick := 0
	emit := func(tick int, ev smaf.Event) {
		d := tick - prevTick
		if d < 0 {
			d = 0
		}
		f.Pairs = append(f.Pairs, smaf.Pair{Duration: d, Event: ev})
		if tick > prevTick {
			prevTick = tick
		}
	}

	for i, it := range items {
		tick := tickAt(it.us)
		ev := it.item
		if n := ev.Note; n != nil {
			offUs, ok := noteOffUs[i]
			if !ok || offUs < it.us {
				offUs = it.us
			}
			gateTicks := int(((offUs - it.us) + int64(gMS)*500) / (int64(gMS) * 1000)) // round to nearest
			if gateTicks < 1 {
				gateTicks = 1
			}
			if gateTicks > (1<<21)-1 {
				gateTicks = (1 << 21) - 1
			}
			n.Gate = gateTicks
			st.Notes++
		}
		emit(tick, ev)
		if tick > maxTick {
			maxTick = tick
		}
		if n := ev.Note; n != nil {
			if end := tick + n.Gate; end > maxTick {
				maxTick = end // notes ring past the last onset
			}
		}
	}
	f.Pairs = append(f.Pairs, smaf.Pair{Duration: 0, Event: smaf.EndOfTrack()})
	st.SkippedCCs = skippedCCs
	st.LengthMS = int64(maxTick) * int64(dMS)
	return f, st, nil
}

// selectBase returns the coarsest duration base whose worst-case rounding
// error over every event time stays within tolerance (microseconds).
func selectBase(items []scoreItem, offs map[int]int64, endUs int64, tolUS int64) int {
	worst := func(base int64) int64 {
		half := base * 500
		unit := base * 1000
		var maxErr int64
		check := func(us int64) {
			rounded := (us + half) / unit * unit
			e := us - rounded
			if e < 0 {
				e = -e
			}
			if e > maxErr {
				maxErr = e
			}
		}
		for _, it := range items {
			check(it.us)
		}
		for _, off := range offs {
			check(off)
		}
		check(endUs)
		return maxErr
	}
	for _, tb := range timeBaseCodes {
		if worst(int64(tb.ms)) <= tolUS {
			return tb.ms
		}
	}
	return 1
}
