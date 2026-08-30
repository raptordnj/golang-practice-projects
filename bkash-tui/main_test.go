package main

import (
	"math"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-6
}

func TestCalculateSendMoney(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		monthly  float64
		isPriyo  bool
		expected float64
	}{
		// Non-positive amount
		{name: "Zero amount", amount: 0, monthly: 0, isPriyo: false, expected: 0},
		{name: "Negative amount", amount: -100, monthly: 0, isPriyo: false, expected: 0},

		// Amount <= 50 is always free
		{name: "Amount <= 50 non-priyo", amount: 50, monthly: 0, isPriyo: false, expected: 0},
		{name: "Amount <= 50 priyo", amount: 30, monthly: 10000, isPriyo: true, expected: 0},

		// Priyo number
		{name: "Priyo within 25k limit", amount: 500, monthly: 10000, isPriyo: true, expected: 0},
		{name: "Priyo exactly at 25k limit", amount: 5000, monthly: 20000, isPriyo: true, expected: 0},
		{name: "Priyo between 25k and 50k", amount: 1000, monthly: 25000, isPriyo: true, expected: 5},
		{name: "Priyo exactly at 50k limit", amount: 10000, monthly: 40000, isPriyo: true, expected: 5},
		{name: "Priyo exceeding 50k", amount: 1000, monthly: 50000, isPriyo: true, expected: 10},

		// Non-priyo number
		{name: "Non-priyo within 25k limit", amount: 500, monthly: 5000, isPriyo: false, expected: 5},
		{name: "Non-priyo exactly at 25k limit", amount: 5000, monthly: 20000, isPriyo: false, expected: 5},
		{name: "Non-priyo exceeding 25k limit", amount: 1000, monthly: 25000, isPriyo: false, expected: 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateSendMoney(tt.amount, tt.monthly, tt.isPriyo)
			if !almostEqual(got, tt.expected) {
				t.Errorf("calculateSendMoney(%v, %v, %v) = %v; want %v", tt.amount, tt.monthly, tt.isPriyo, got, tt.expected)
			}
		})
	}
}

func TestCalculateCashOut(t *testing.T) {
	tests := []struct {
		name     string
		amount   float64
		agent    int
		expected float64
	}{
		// Non-positive amount
		{name: "Zero amount", amount: 0, agent: 0, expected: 0},
		{name: "Negative amount", amount: -500, agent: 1, expected: 0},

		// Regular Agent: 1.85% (18.50 / 1000)
		{name: "Regular Agent 1000 Tk", amount: 1000, agent: 0, expected: 18.50},
		{name: "Regular Agent 5000 Tk", amount: 5000, agent: 0, expected: 92.50},

		// Priyo Agent: 1.49% (14.90 / 1000)
		{name: "Priyo Agent 1000 Tk", amount: 1000, agent: 1, expected: 14.90},
		{name: "Priyo Agent 10000 Tk", amount: 10000, agent: 1, expected: 149.00},

		// Priyo Offer: 1.395% (13.95 / 1000)
		{name: "Priyo Offer 1000 Tk", amount: 1000, agent: 2, expected: 13.95},

		// ATM: 1.49% (14.90 / 1000)
		{name: "ATM 1000 Tk", amount: 1000, agent: 3, expected: 14.90},

		// Default fallback to Regular Agent
		{name: "Invalid Agent Index", amount: 1000, agent: 99, expected: 18.50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateCashOut(tt.amount, tt.agent)
			if !almostEqual(got, tt.expected) {
				t.Errorf("calculateCashOut(%v, %v) = %v; want %v", tt.amount, tt.agent, got, tt.expected)
			}
		})
	}
}

func TestModelTransitions(t *testing.T) {
	m := initialModel()
	if m.screen != menuScreen {
		t.Fatalf("expected initial screen to be menuScreen, got %v", m.screen)
	}

	// 1. Enter Send Money screen from menu
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if m.screen != sendMoneyScreen {
		t.Fatalf("expected screen to be sendMoneyScreen, got %v", m.screen)
	}

	// 2. Tab through inputs
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(model)
	if m.focusIndex != 1 {
		t.Fatalf("expected focusIndex to be 1 after tab, got %d", m.focusIndex)
	}

	// 3. Shift-Tab back
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = updated.(model)
	if m.focusIndex != 0 {
		t.Fatalf("expected focusIndex to be 0 after shift+tab, got %d", m.focusIndex)
	}

	// 4. Fill Send Money inputs and calculate
	m.amountInput.SetValue("1000")
	m.monthlyInput.SetValue("5000")
	m.priyoInput.SetValue("y")
	m.focusIndex = 2

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if m.screen != resultScreen {
		t.Fatalf("expected screen to transition to resultScreen, got %v", m.screen)
	}
	if len(m.resultLines) == 0 {
		t.Fatalf("expected resultLines to be populated")
	}

	// 5. Return to menu screen from result screen
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if m.screen != menuScreen {
		t.Fatalf("expected screen to return to menuScreen, got %v", m.screen)
	}

	// 6. Test Cash Out navigation and calculation
	m.cursor = 1 // select Cash Out
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if m.screen != cashOutScreen {
		t.Fatalf("expected screen to be cashOutScreen, got %v", m.screen)
	}

	m.cashAmountInput.SetValue("2000")
	// Enter on amount input switches to agent selector
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if m.focusIndex != 1 {
		t.Fatalf("expected focusIndex to be 1 (agent selection), got %d", m.focusIndex)
	}

	// Enter on agent selector calculates result
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(model)
	if m.screen != resultScreen {
		t.Fatalf("expected screen to transition to resultScreen, got %v", m.screen)
	}

	// ESC on result screen goes back to menu
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(model)
	if m.screen != menuScreen {
		t.Fatalf("expected screen to return to menuScreen, got %v", m.screen)
	}
}
