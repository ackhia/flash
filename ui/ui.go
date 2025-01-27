package ui

import (
	"fmt"
	"log"
	"strings"

	"github.com/ackhia/flash/node"
	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type page int

const (
	mainPage page = iota
	myNodePage
	sendTransactionPage
	viewPeersPage
	viewLogsPage
)

type Model struct {
	currentPage    page
	menuOptions    []string
	selectedOption int
	peerIDInput    textinput.Model
	amountInput    textinput.Model
	table          table.Model
	viewport       viewport.Model
	peerID         string
	peerMA         string
	balance        float64
	connectedPeers int
	totalCoins     float64
	peers          []peer
	logs           []string
	node           *node.Node
}

type peer struct {
	ID      string
	Balance float64
}

func initialModel() Model {
	menu := []string{
		"My Node",
		"Send Transaction",
		"View Peers",
		"View Logs",
	}

	columns := []table.Column{
		{Title: "Peer ID", Width: 30},
		{Title: "Balance", Width: 10},
	}
	t := table.New(
		table.WithColumns(columns),
		table.WithHeight(10),
	)

	peerIDInput := textinput.New()
	peerIDInput.Placeholder = "Enter Peer ID"

	amountInput := textinput.New()
	amountInput.Placeholder = "Enter Amount"

	vp := viewport.New(40, 10)
	return Model{
		currentPage:    mainPage,
		menuOptions:    menu,
		peerIDInput:    peerIDInput,
		amountInput:    amountInput,
		table:          t,
		viewport:       vp,
		peerID:         "",
		balance:        0,
		connectedPeers: 0,
		totalCoins:     0,
		peers: []peer{
			{"peer1", 50.0},
			{"peer2", 30.0},
		},
		logs: []string{},
	}
}

func (m *Model) Init() tea.Cmd {
	m.refreshModel()
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "up":
			if m.currentPage == mainPage && m.selectedOption > 0 {
				m.selectedOption--
			}
		case "down":
			if m.currentPage == mainPage && m.selectedOption < len(m.menuOptions)-1 {
				m.selectedOption++
			}
		case "enter":
			if m.currentPage == mainPage {
				m.currentPage = page(m.selectedOption + 1)
				m.refreshModel()
				if m.currentPage == sendTransactionPage {
					m.peerIDInput.Focus()
					m.amountInput.Blur()
				}
			} else if m.currentPage == sendTransactionPage {
				if m.peerIDInput.Focused() {
					m.peerIDInput.Blur()
					m.amountInput.Focus()
				} else if m.amountInput.Focused() {
					peerID := strings.TrimSpace(m.peerIDInput.Value())
					amount := strings.TrimSpace(m.amountInput.Value())
					if peerID != "" && amount != "" {
						m.logs = append(m.logs, fmt.Sprintf("Transaction sent: PeerID=%s, Amount=%s", peerID, amount))
						m.peerIDInput.SetValue("")
						m.amountInput.SetValue("")
						m.peerIDInput.Focus()
					}
				}
			}
		case "tab":
			if m.currentPage == sendTransactionPage {
				if m.peerIDInput.Focused() {
					m.peerIDInput.Blur()
					m.amountInput.Focus()
				} else {
					m.peerIDInput.Focus()
					m.amountInput.Blur()
				}
			}
		case "c":
			if m.currentPage == myNodePage {
				clipboard.WriteAll(m.peerMA)
			}

		case "esc":
			if m.currentPage == mainPage {
				return m, tea.Quit
			}
			m.currentPage = mainPage
			m.peerIDInput.Blur()
			m.amountInput.Blur()
		}

	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 2
	}

	if m.currentPage == sendTransactionPage {
		var cmd1, cmd2 tea.Cmd
		m.peerIDInput, cmd1 = m.peerIDInput.Update(msg)
		m.amountInput, cmd2 = m.amountInput.Update(msg)
		return m, tea.Batch(cmd1, cmd2)
	}

	return m, nil
}

func (m Model) View() string {
	switch m.currentPage {
	case mainPage:
		return m.viewMainMenu()
	case myNodePage:
		return m.viewMyNode()
	case sendTransactionPage:
		return m.viewSendTransaction()
	case viewPeersPage:
		return m.viewPeers()
	case viewLogsPage:
		return m.viewLogs()
	}
	return ""
}

func (m Model) viewMainMenu() string {
	var sb strings.Builder
	sb.WriteString("Select an option\n\n")
	for i, opt := range m.menuOptions {
		cursor := " "
		if i == m.selectedOption {
			cursor = ">"
		}
		sb.WriteString(fmt.Sprintf("%s %s\n", cursor, opt))
	}
	sb.WriteString("\nPress ESC to quit")
	return sb.String()
}

func (m Model) viewMyNode() string {
	return fmt.Sprintf(
		"My Node:\n\n%-30s %s\n%-30s %s\n%-30s %.2f\n%-30s %d\n%-30s %.2f\n\nPress ESC to go back. Press c to copy Peer Multiaddress to clipboard",
		"Peer ID:", m.peerID,
		"Peer Multiaddress:", m.peerMA,
		"Balance:", m.balance,
		"Connected Peers:", m.connectedPeers,
		"Coins in Circulation:", m.totalCoins,
	)
}

func (m Model) viewSendTransaction() string {
	return fmt.Sprintf(
		"Send Transaction:\n\n%s\n%s\n\nPress ENTER to send, TAB to switch, ESC to go back.",
		m.peerIDInput.View(),
		m.amountInput.View(),
	)
}

func (m Model) viewPeers() string {
	rows := []table.Row{}
	for _, p := range m.peers {
		rows = append(rows, table.Row{p.ID, fmt.Sprintf("%.2f", p.Balance)})
	}
	m.table.SetCursor(-1)
	m.table.SetRows(rows)
	return fmt.Sprintf("Peers:\n\n%s\n\nPress ESC to go back.", m.table.View())
}

func (m Model) viewLogs() string {
	return fmt.Sprintf(
		"Logs:\n\n%s\n\nPress ESC to go back.",
		strings.Join(m.logs, "\n"),
	)
}

func (m *Model) refreshModel() {
	m.peerID = m.node.Host.ID().String()
	m.balance = m.node.Balances[m.node.Host.ID().String()]
	m.totalCoins = m.node.TotalCoins
	m.connectedPeers = len(m.node.Host.Network().Peers())
	m.peers = []peer{}
	for _, p := range m.node.Host.Network().Peers() {
		m.peers = append(m.peers, peer{
			ID:      p.String(),
			Balance: m.node.Balances[p.String()],
		})
	}

	ma, err := node.CreateMultiaddress(m.node)
	if err != nil {
		log.Print("Could not create multiaddress")
		return
	}
	m.peerMA = ma
}

func Show(n *node.Node) {
	m := initialModel()
	m.node = n
	p := tea.NewProgram(&m)
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}
