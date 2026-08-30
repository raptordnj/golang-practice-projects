package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type screen int

const (
	menuScreen screen = iota
	sendMoneyScreen
	cashOutScreen
	resultScreen
)

type model struct {
	screen      screen
	cursor      int
	menu        []string
	agentTypes  []string
	agentCursor int

	// Send Money inputs
	amountInput  textinput.Model
	monthlyInput textinput.Model
	priyoInput   textinput.Model
	focusIndex   int

	// Cash Out inputs
	cashAmountInput textinput.Model

	resultTitle string
	resultLines []string
	errMsg      string
}

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#E2136E")).MarginBottom(1)
	boxStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#E2136E")).Padding(1, 2)
	helpStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	selStyle   = lipgloss.NewStyle().Background(lipgloss.Color("#E2136E")).Foreground(lipgloss.Color("#FFFFFF")).Padding(0, 1)
	errStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444")).Bold(true)
)

func initialModel() model {
	am := textinput.New()
	am.Placeholder = "e.g. 1000"
	am.Focus()
	am.CharLimit = 15
	am.Width = 24

	mo := textinput.New()
	mo.Placeholder = "0 (already sent this month)"
	mo.CharLimit = 15
	mo.Width = 24

	pr := textinput.New()
	pr.Placeholder = "y / n (default: n)"
	pr.CharLimit = 5
	pr.Width = 24

	ca := textinput.New()
	ca.Placeholder = "e.g. 5000"
	ca.CharLimit = 15
	ca.Width = 24

	return model{
		screen: menuScreen,
		menu:   []string{"Send Money", "Cash Out", "Exit"},
		agentTypes: []string{
			"Regular Agent (1.85% - 18.50 Tk / 1,000)",
			"Priyo Agent (1.49% - 14.90 Tk / 1,000)",
			"Priyo Agent App Offer (1.395% - 13.95 Tk / 1,000)",
			"ATM Cash Out (1.49% - 14.90 Tk / 1,000)",
		},
		amountInput:     am,
		monthlyInput:    mo,
		priyoInput:      pr,
		cashAmountInput: ca,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func calculateSendMoney(amount, monthly float64, isPriyo bool) float64 {
	if amount <= 0 {
		return 0
	}
	if amount <= 50 {
		return 0
	}
	total := monthly + amount
	if isPriyo {
		if total <= 25000 {
			return 0
		} else if total <= 50000 {
			return 5
		}
		return 10
	}
	if total <= 25000 {
		return 5
	}
	return 10
}

func calculateCashOut(amount float64, agent int) float64 {
	if amount <= 0 {
		return 0
	}
	switch agent {
	case 0:
		return amount * 0.0185
	case 1:
		return amount * 0.0149
	case 2:
		return amount * 0.01395
	case 3:
		return amount * 0.0149
	default:
		return amount * 0.0185
	}
}

func (m *model) updateSendMoneyFocus() []tea.Cmd {
	cmds := make([]tea.Cmd, 3)
	if m.focusIndex == 0 {
		cmds[0] = m.amountInput.Focus()
		m.monthlyInput.Blur()
		m.priyoInput.Blur()
	} else if m.focusIndex == 1 {
		m.amountInput.Blur()
		cmds[1] = m.monthlyInput.Focus()
		m.priyoInput.Blur()
	} else if m.focusIndex == 2 {
		m.amountInput.Blur()
		m.monthlyInput.Blur()
		cmds[2] = m.priyoInput.Focus()
	}
	return cmds
}

func (m *model) updateCashOutFocus() []tea.Cmd {
	cmds := make([]tea.Cmd, 1)
	if m.focusIndex == 0 {
		cmds[0] = m.cashAmountInput.Focus()
	} else {
		m.cashAmountInput.Blur()
	}
	return cmds
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		k := msg.String()

		// Global quit
		if k == "ctrl+c" {
			return m, tea.Quit
		}

		// Quit on 'q' only when on menuScreen or resultScreen
		if k == "q" && (m.screen == menuScreen || m.screen == resultScreen) {
			return m, tea.Quit
		}

		// Esc returns to menuScreen
		if k == "esc" {
			if m.screen != menuScreen {
				m.screen = menuScreen
				m.errMsg = ""
				return m, nil
			}
		}

		switch m.screen {
		case menuScreen:
			switch k {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.menu)-1 {
					m.cursor++
				}
			case "enter":
				switch m.cursor {
				case 0:
					m.screen = sendMoneyScreen
					m.focusIndex = 0
					m.errMsg = ""
					m.amountInput.Reset()
					m.monthlyInput.Reset()
					m.priyoInput.Reset()
					cmds := m.updateSendMoneyFocus()
					return m, tea.Batch(append(cmds, textinput.Blink)...)
				case 1:
					m.screen = cashOutScreen
					m.focusIndex = 0
					m.agentCursor = 0
					m.errMsg = ""
					m.cashAmountInput.Reset()
					cmds := m.updateCashOutFocus()
					return m, tea.Batch(append(cmds, textinput.Blink)...)
				case 2:
					return m, tea.Quit
				}
			}

		case sendMoneyScreen:
			switch k {
			case "tab", "down":
				m.focusIndex = (m.focusIndex + 1) % 3
				cmds := m.updateSendMoneyFocus()
				return m, tea.Batch(cmds...)
			case "shift+tab", "up":
				m.focusIndex = (m.focusIndex - 1 + 3) % 3
				cmds := m.updateSendMoneyFocus()
				return m, tea.Batch(cmds...)
			case "enter":
				if m.focusIndex < 2 {
					m.focusIndex++
					cmds := m.updateSendMoneyFocus()
					return m, tea.Batch(cmds...)
				} else {
					amtStr := strings.TrimSpace(m.amountInput.Value())
					amt, err := strconv.ParseFloat(amtStr, 64)
					if err != nil || amt <= 0 {
						m.errMsg = "Please enter a valid positive transfer amount."
						return m, nil
					}

					monStr := strings.TrimSpace(m.monthlyInput.Value())
					var mon float64
					if monStr != "" {
						mon, err = strconv.ParseFloat(monStr, 64)
						if err != nil || mon < 0 {
							m.errMsg = "Please enter a valid monthly amount (>= 0)."
							return m, nil
						}
					}

					m.errMsg = ""
					priyoVal := strings.ToLower(strings.TrimSpace(m.priyoInput.Value()))
					isPriyo := priyoVal == "y" || priyoVal == "yes"

					charge := calculateSendMoney(amt, mon, isPriyo)
					totalDeducted := amt + charge

					priyoLabel := "No (Regular Number)"
					if isPriyo {
						priyoLabel = "Yes (Priyo Number)"
					}

					m.resultTitle = "Send Money Result"
					m.resultLines = []string{
						fmt.Sprintf("Transfer Amount:        %.2f Tk", amt),
						fmt.Sprintf("Monthly Total Sent:     %.2f Tk", mon+amt),
						fmt.Sprintf("Priyo Status:           %s", priyoLabel),
						fmt.Sprintf("Service Charge:         %.2f Tk", charge),
						"────────────────────────────────────────",
						fmt.Sprintf("Total Wallet Deducted:  %.2f Tk", totalDeducted),
					}
					m.screen = resultScreen
					return m, nil
				}
			}

		case cashOutScreen:
			switch k {
			case "tab":
				m.focusIndex = (m.focusIndex + 1) % 2
				cmds := m.updateCashOutFocus()
				return m, tea.Batch(cmds...)
			case "shift+tab":
				m.focusIndex = (m.focusIndex - 1 + 2) % 2
				cmds := m.updateCashOutFocus()
				return m, tea.Batch(cmds...)
			case "up", "k":
				if m.focusIndex == 1 {
					if m.agentCursor > 0 {
						m.agentCursor--
					}
				}
			case "down", "j":
				if m.focusIndex == 1 {
					if m.agentCursor < len(m.agentTypes)-1 {
						m.agentCursor++
					}
				}
			case "enter":
				if m.focusIndex == 0 {
					m.focusIndex = 1
					cmds := m.updateCashOutFocus()
					return m, tea.Batch(cmds...)
				} else {
					amtStr := strings.TrimSpace(m.cashAmountInput.Value())
					amt, err := strconv.ParseFloat(amtStr, 64)
					if err != nil || amt <= 0 {
						m.errMsg = "Please enter a valid positive cash amount."
						return m, nil
					}
					m.errMsg = ""

					charge := calculateCashOut(amt, m.agentCursor)
					totalDeducted := amt + charge

					m.resultTitle = "Cash Out Result"
					m.resultLines = []string{
						fmt.Sprintf("Cash Amount:            %.2f Tk", amt),
						fmt.Sprintf("Cash Out Method:        %s", m.agentTypes[m.agentCursor]),
						fmt.Sprintf("Cash Out Fee:           %.2f Tk", charge),
						"────────────────────────────────────────",
						fmt.Sprintf("Total Wallet Deducted:  %.2f Tk", totalDeducted),
						fmt.Sprintf("Cash In Hand Received:  %.2f Tk", amt),
					}
					m.screen = resultScreen
					return m, nil
				}
			}

		case resultScreen:
			if k == "enter" || k == "esc" {
				m.screen = menuScreen
				m.errMsg = ""
				return m, nil
			}
		}
	}

	var cmds []tea.Cmd
	if m.screen == sendMoneyScreen {
		var cmd tea.Cmd
		switch m.focusIndex {
		case 0:
			m.amountInput, cmd = m.amountInput.Update(msg)
		case 1:
			m.monthlyInput, cmd = m.monthlyInput.Update(msg)
		case 2:
			m.priyoInput, cmd = m.priyoInput.Update(msg)
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	} else if m.screen == cashOutScreen && m.focusIndex == 0 {
		var cmd tea.Cmd
		m.cashAmountInput, cmd = m.cashAmountInput.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	switch m.screen {
	case menuScreen:
		s := titleStyle.Render("bKash Charge Calculator") + "\n"
		s += "Choose service (official rates)\n\n"
		for i, choice := range m.menu {
			if m.cursor == i {
				s += selStyle.Render("> "+choice) + "\n"
			} else {
				s += "  " + choice + "\n"
			}
		}
		s += "\n" + helpStyle.Render("↑/↓ / j/k navigate • enter select • q quit")
		return boxStyle.Render(s)

	case sendMoneyScreen:
		s := titleStyle.Render("bKash Send Money") + "\n\n"
		if m.errMsg != "" {
			s += errStyle.Render("⚠ "+m.errMsg) + "\n\n"
		}
		s += fmt.Sprintf("Amount (Tk):            %s\n", m.amountInput.View())
		s += fmt.Sprintf("Monthly Sent (Tk):      %s\n", m.monthlyInput.View())
		s += fmt.Sprintf("Priyo Number? (y/n):    %s\n", m.priyoInput.View())
		s += "\n" + helpStyle.Render("tab / ↑ / ↓ switch fields • enter proceed • esc menu")
		return boxStyle.Render(s)

	case cashOutScreen:
		s := titleStyle.Render("bKash Cash Out") + "\n\n"
		if m.errMsg != "" {
			s += errStyle.Render("⚠ "+m.errMsg) + "\n\n"
		}
		s += fmt.Sprintf("Amount (Tk): %s\n\n", m.cashAmountInput.View())
		s += "Agent Type / Method:\n"
		for i, a := range m.agentTypes {
			if m.focusIndex == 1 && m.agentCursor == i {
				s += selStyle.Render("> "+a) + "\n"
			} else {
				s += "  " + a + "\n"
			}
		}
		s += "\n" + helpStyle.Render("tab switch focus • ↑/↓ / j/k choose agent • enter proceed • esc menu")
		return boxStyle.Render(s)

	case resultScreen:
		s := titleStyle.Render(m.resultTitle) + "\n\n"
		for _, l := range m.resultLines {
			s += l + "\n"
		}
		s += "\n" + helpStyle.Render("press enter / esc to return to menu • q quit")
		return boxStyle.Render(s)
	}
	return ""
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error:", err)
	}
}
