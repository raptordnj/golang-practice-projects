// Command midi2mmf converts Standard MIDI Files to Yamaha SMAF (.mmf)
// ringtone files with a score track (MA-2/MA-3 Mobile Standard format).
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aurnob/midi2mmf/internal/convert"
	"github.com/aurnob/midi2mmf/internal/midi"
	"github.com/aurnob/midi2mmf/internal/smaf"

	flag "flag"
)

func main() {
	var (
		title      = flag.String("title", "", "song title embedded in the file (default: input file name)")
		artist     = flag.String("artist", "", "artist name embedded as tag AN")
		composer   = flag.String("composer", "", "composer name embedded as tag SW")
		timeBaseMS = flag.Int("timebase-ms", 0, "SMAF time base in ms (0 = auto-select from 1, 2, 4, 5, 10, 20, 40, 50)")
		maxErrorMS = flag.Int("max-error-ms", 4, "auto time-base selection tolerance in ms")
		info       = flag.Bool("info", false, "inspect an .mmf file instead of converting")
		quiet      = flag.Bool("q", false, "suppress the conversion summary")
	)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `usage: %s [flags] <input.mid> [output.mmf]
       %s -info <file.mmf>

Converts a Standard MIDI File to a Yamaha SMAF (.mmf) ringtone using an
uncompressed MA-2/MA-3 score track.
`, filepath.Base(os.Args[0]), filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()

	args := flag.Args()
	if *info {
		if len(args) != 1 {
			usageErr("-info needs exactly one .mmf file")
		}
		runInfo(args[0])
		return
	}
	switch len(args) {
	case 1:
	case 2:
	default:
		usageErr("expected an input .mid file")
	}

	in := args[0]
	out := ""
	if len(args) == 2 {
		out = args[1]
	} else {
		ext := filepath.Ext(in)
		out = strings.TrimSuffix(in, ext) + ".mmf"
	}

	if *title == "" {
		base := filepath.Base(strings.TrimSuffix(in, filepath.Ext(in)))
		*title = base
	}

	data, err := os.ReadFile(in)
	if err != nil {
		fatal(err)
	}
	song, err := midi.Parse(data)
	if err != nil {
		fatal(fmt.Errorf("%s: %w", in, err))
	}
	if song.TruncatedTracks > 0 {
		fmt.Fprintf(os.Stderr, "warning: %d track(s) had malformed data; events after the corruption were dropped\n",
			song.TruncatedTracks)
	}

	file, stats, err := convert.ToSMAF(song, convert.Options{
		Title:       *title,
		Artist:      *artist,
		Composer:    *composer,
		TimeBaseDMS: *timeBaseMS,
		TimeBaseGMS: *timeBaseMS,
		MaxErrorMS:  *maxErrorMS,
	})
	if err != nil {
		fatal(err)
	}
	image, err := file.Encode()
	if err != nil {
		fatal(err)
	}
	if err := os.WriteFile(out, image, 0o644); err != nil {
		fatal(err)
	}

	if !*quiet {
		fmt.Printf("%s -> %s\n", in, out)
		fmt.Printf("  notes %d  cc %d  pc %d  bend %d  (%d CC dropped)\n",
			stats.Notes, stats.CCs, stats.PCs, stats.Bends, stats.SkippedCCs)
		ch := make([]string, 0, 16)
		for i, used := range stats.ChannelsUsed {
			if used {
				ch = append(ch, fmt.Sprint(i+1))
			}
		}
		fmt.Printf("  channels %s\n", strings.Join(ch, ","))
		fmt.Printf("  length %.1fs  time base %d/%d ms\n",
			float64(stats.LengthMS)/1000, stats.TimeBaseDMS, stats.TimeBaseGMS)
	}
}

func runInfo(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fatal(err)
	}
	d, err := smaf.Decode(data)
	if err != nil {
		fatal(err)
	}
	crc := "OK"
	if !d.CRCValid {
		crc = "INVALID"
	}
	fmt.Printf("%s: %d bytes, CRC %s\n", path, d.FileSize, crc)
	fmt.Printf("  chunks: %s\n", strings.Join(d.Chunks, " "))
	fmt.Printf("  score format type %d, time base D=%#02x G=%#02x\n",
		d.FormatType, d.TimeBaseD, d.TimeBaseG)
	fmt.Printf("  title %q artist %q composer %q\n", d.Title, d.Artist, d.Composer)

	notes, cc, pc, bend, sysex := 0, 0, 0, 0, 0
	chans := map[int]bool{}
	dMS := int64(smaf.TimeBaseMS[d.TimeBaseD])
	var lastTick, audible int64
	for _, p := range d.Pairs {
		lastTick += int64(p.Duration)
		switch {
		case p.Event.Note != nil:
			notes++
			chans[p.Event.Note.Channel] = true
			if end := lastTick + int64(p.Event.Note.Gate); end > audible {
				audible = end
			}
		case p.Event.CC != nil:
			cc++
		case p.Event.PC != nil:
			pc++
		case p.Event.Bend != nil:
			bend++
		case p.Event.SysEx != nil:
			sysex++
		}
	}
	if audible < lastTick {
		audible = lastTick
	}
	ch := make([]string, 0, len(chans))
	for i := range chans {
		ch = append(ch, fmt.Sprint(i+1))
	}
	sort.Strings(ch)
	fmt.Printf("  events: %d notes, %d cc, %d pc, %d bend, %d sysex; clean end: %v\n",
		notes, cc, pc, bend, sysex, d.EndedClean)
	fmt.Printf("  length %.1fs on channels %s\n",
		float64(audible)*float64(dMS)/1000, strings.Join(ch, ","))
}

func usageErr(msg string) {
	fmt.Fprintf(os.Stderr, "%s: %s\n\n", os.Args[0], msg)
	flag.Usage()
	os.Exit(2)
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "midi2mmf: %v\n", err)
	os.Exit(1)
}
