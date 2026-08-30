package midi

import (
	"os"
	"testing"
)

func TestDebugTiming(t *testing.T) {
	data, _ := os.ReadFile("/tmp/smaf-ref/train.mid")
	s, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("ppq=%d tempos=%d", s.TicksPerQN, len(s.TempoChanges))
	for i, tc := range s.TempoChanges {
		if i < 6 || i > len(s.TempoChanges)-3 {
			t.Logf("  tempo[%d] tick=%d us/qn=%d (%.1f bpm)", i, tc.Tick, tc.A, 60e6/float64(tc.A))
		}
	}
	total := s.TotalTicks()
	t.Logf("total ticks=%d -> %.3f s", total, float64(s.TimeMicros(total))/1e6)
	last := s.Events[len(s.Events)-1]
	t.Logf("last event tick=%d -> %.3f s", last.Tick, float64(s.TimeMicros(last.Tick))/1e6)
}
