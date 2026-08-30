# midi2mmf

Convert Standard MIDI Files (`.mid`) to Yamaha SMAF ringtones (`.mmf`) — the
polyphonic format used by early-2000s mobile phones (MA-2/MA-3 sound chips).

Pure Go, no dependencies, stdlib only.

```
go build -o midi2mmf .
./midi2mmf song.mid                 # -> song.mmf
./midi2mmf -title "My Song" a.mid b.mmf
./midi2mmf -info b.mmf              # inspect any score-track .mmf
```

## Flags

| Flag | Meaning |
|---|---|
| `-title`, `-artist`, `-composer` | Text tags embedded in the file's OPDA chunk (Latin-1). Title defaults to the input file name. |
| `-timebase-ms N` | SMAF time base (1/2/4/5/10/20/40/50 ms). Default: auto-select — the coarsest base whose worst-case quantization error stays within `-max-error-ms`. |
| `-max-error-ms N` | Tolerance for auto time-base selection (default 4). |
| `-info` | Don't convert; dump the structure of an `.mmf` instead. |
| `-q` | Suppress the summary. |

## How it works

SMAF score timing is **metric** — durations count milliseconds via a base unit
in the MTR header — so there is nothing like MIDI's tempo meta event. The
converter integrates the MIDI tempo map into absolute microseconds for every
event, picks the coarsest time base that keeps quantization error within
tolerance, and lays events out as duration+event pairs on that grid. Notes
become single note-on messages carrying a gate time (the score format has no
note-offs); velocity is sent explicitly only when it changes, using the
format's compact velocity-reuse form.

Output container: `MMMD` → `CNTI` → `OPDA`(title tags) → `MTR`(`Mtsu` Yamaha
init sysex + `Mtsq` sequence), closed by a big-endian CRC-16/CCITT (init
`0xFFFF`, poly `0x1021`, final inversion) over the whole file.

## What is preserved / dropped

Kept: notes (all 16 channels), program changes, pitch bend, and common control
changes (modulation, volume, pan, expression, sustain, …). Channel 10 keeps its
GM percussion role.

Dropped: bank select / RPN/NRPN payloads (MA preset tones are selected by
program number alone), all-sound-off-style controllers, SysEx from the source
file, key/time signatures (no equivalent).

## Known limitations

- **Instruments**: GM program numbers are passed through unchanged. The MA
  chips' preset tone bank is not identical to General MIDI, so timbres differ;
  a GM→MA voice mapping table would be needed to do better.
- No loop points (`MspI` chunk), no embedded ADPCM/PCM audio (`ATR`/`Mtsp`
  chunks), no Huffman-compressed sequences.
- Type-2 MIDI files are treated as simultaneous tracks.

## Format notes

The binary layout was derived from two real-world ringtone files and
cross-checked against two independent open-source parsers:

- [but80/go-smaf](https://github.com/but80/go-smaf) — Go SMAF parser (score
  track header, event opcodes, varlen, exclusive framing)
- [Rockbox](https://git.rockbox.org/cgit/rb.git/tree/lib/rbcodec/metadata/smaf.c)
  — metadata parser (chunk walk, CNTI/Dch tag layouts, code types)

Event opcodes in the `Mtsq` sequence: `90|ch note vel gate` /
`80|ch note gate` (velocity reuse), `B0|ch cc val`, `C0|ch pc`,
`E0|ch lsb msb`, `F0 len data F7`, NOP `FF 00`, end `FF 2F 00`; each preceded
by an SMAF variable-length duration.

## Development

```
go test ./...   # includes round-trip and malformed-input tolerance tests
go vet ./...
```
