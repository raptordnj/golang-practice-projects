package game

import "testing"

func benchState() (*TraditionalBaghBandiRules, GameState) {
	r := NewDefaultRules()
	s := r.NewGame()
	// A representative mid-game position: goats spread out, tiger loose in
	// the middle with captures available.
	for _, n := range []string{"a1", "b1", "c1", "e1", "a2", "e2", "e3", "b3"} {
		id, _ := ParseNotation(n)
		s.Occupancy[id] = Goat
		s.GoatsPlaced++
	}
	s.Turn = TigerPlayer
	return r, s
}

func BenchmarkLegalMovesTiger(b *testing.B) {
	r, s := benchState()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.LegalMoves(s)
	}
}

func BenchmarkAppendLegalMovesTiger(b *testing.B) {
	r, s := benchState()
	buf := make([]Move, 0, 32)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf = r.AppendLegalMoves(buf[:0], s)
	}
	_ = buf
}

func BenchmarkLegalMovesGoatPlacement(b *testing.B) {
	r := NewDefaultRules()
	s := r.NewGame()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.LegalMoves(s)
	}
}

func BenchmarkApplyMoveCapture(b *testing.B) {
	r, s := benchState()
	m := CaptureMove(ID(2, 2), ID(2, 1), ID(2, 0))
	if !r.IsValidMove(s, m) {
		b.Fatalf("benchmark fixture move is illegal")
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.ApplyMove(s, m)
	}
}

func BenchmarkApplyMovePlacement(b *testing.B) {
	r := NewDefaultRules()
	s := r.NewGame()
	m := PlaceMove(ID(0, 0))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.ApplyMove(s, m)
	}
}
