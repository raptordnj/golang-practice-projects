package main

import "testing"

func BenchmarkAppendApproach(b *testing.B) {
	for i := 0; i < b.N; i++ {
		numbers := []int{1, 2, 3, 4, 5}
		_ = append(numbers[:4], numbers[5:]...)
	}
}

func BenchmarkCopyApproach(b *testing.B) {
	for i := 0; i < b.N; i++ {
		numbers := []int{1, 2, 3, 4, 5}
		copy(numbers[4:], numbers[5:])
		numbers = numbers[:len(numbers)-1]
	}
}

// Larger slice benchmarks
func BenchmarkAppendApproachLarge(b *testing.B) {
	for i := 0; i < b.N; i++ {
		numbers := make([]int, 1000)
		for j := range numbers {
			numbers[j] = j
		}
		_ = append(numbers[:500], numbers[501:]...)
	}
}

func BenchmarkCopyApproachLarge(b *testing.B) {
	for i := 0; i < b.N; i++ {
		numbers := make([]int, 1000)
		for j := range numbers {
			numbers[j] = j
		}
		copy(numbers[500:], numbers[501:])
		numbers = numbers[:len(numbers)-1]
	}
}
