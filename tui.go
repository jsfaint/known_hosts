package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the TUI application state
type Model struct {
	hosts       []string
	filtered    []string
	cursor      int
	search      string
	isSearching bool
	mode        viewMode
	err         error
	status      string
}

type viewMode int

const (
	viewList viewMode = iota
	viewConfirmDelete
)

func (m Model) Init() tea.Cmd {
	return loadHosts()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	case errMsg:
		m.err = msg.err
		return m, nil
	case hostsLoadedMsg:
		m.hosts = msg.hosts
		m.filtered = msg.hosts
		return m, nil
	}
	return m, nil
}

func (m Model) View() string {
	if m.err != nil {
		return m.renderError()
	}

	switch m.mode {
	case viewList:
		return m.renderList()
	case viewConfirmDelete:
		return m.renderConfirmDelete()
	default:
		return ""
	}
}

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87")).
			Bold(true)

	searchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))
)

// renderError displays error message
func (m Model) renderError() string {
	return errorStyle.Render("Error: "+m.err.Error()) + "\n\nPress 'q' to quit"
}

// renderList displays the host list
func (m Model) renderList() string {
	var s strings.Builder

	// Title
	s.WriteString(titleStyle.Render("Known Hosts Manager") + "\n\n")

	// Summary
	s.WriteString(footerStyle.Render(m.renderSummary()) + "\n")
	if m.status != "" {
		s.WriteString(statusStyle.Render(m.status) + "\n")
	}
	s.WriteString("\n")

	// Search bar
	if m.isSearching {
		s.WriteString(searchStyle.Render("Search: ") + m.search + "_\n\n")
	} else if m.search != "" {
		s.WriteString(searchStyle.Render("Filter: ") + m.search + "\n\n")
	}

	// Host list
	if len(m.filtered) == 0 {
		switch {
		case len(m.hosts) == 0:
			s.WriteString(normalStyle.Render("No known hosts available"))
		case m.search != "":
			s.WriteString(normalStyle.Render("No hosts found for filter: " + m.search))
		default:
			s.WriteString(normalStyle.Render("No hosts found"))
		}
	} else {
		for i, hostLine := range m.filtered {
			cursor := " "
			if i == m.cursor {
				cursor = ">"
			}

			host, err := NewHost(hostLine)
			if err != nil {
				continue
			}

			line := cursor + " " + host.DisplayName()
			if i == m.cursor {
				s.WriteString(selectedStyle.Render(line))
			} else {
				s.WriteString(normalStyle.Render(line))
			}
			s.WriteString("\n")
		}
	}

	// Footer
	s.WriteString("\n" + footerStyle.Render("Controls: ↑↓/Home/End navigate | d delete | / search | q quit"))

	return s.String()
}

func (m Model) renderSummary() string {
	if len(m.hosts) == 0 {
		return "Showing 0 hosts"
	}

	summary := fmt.Sprintf("Showing %d of %d hosts", len(m.filtered), len(m.hosts))
	if len(m.filtered) > 0 {
		summary += fmt.Sprintf(" | Selected %d/%d", m.cursor+1, len(m.filtered))
	}

	return summary
}

// renderConfirmDelete displays delete confirmation
func (m Model) renderConfirmDelete() string {
	hostLine := m.filtered[m.cursor]
	host, err := NewHost(hostLine)
	if err != nil {
		return errorStyle.Render("Error: " + err.Error())
	}

	var s string
	s += titleStyle.Render("Confirm Deletion") + "\n\n"
	s += normalStyle.Render("Delete this host?\n\n")
	s += selectedStyle.Render(host.DisplayName()) + "\n\n"
	s += footerStyle.Render("Press Enter or 'y' to confirm, 'n' to cancel")

	return s
}

// handleKeyMsg processes keyboard input
func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case viewList:
		return m.handleListKeyMsg(msg)
	case viewConfirmDelete:
		return m.handleConfirmKeyMsg(msg)
	}
	return m, nil
}

// handleListKeyMsg processes keys in list view
func (m Model) handleListKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyEsc:
		return m, tea.Quit

	case tea.KeyUp:
		if m.cursor > 0 {
			m.cursor--
		}
	case tea.KeyDown:
		if m.cursor < len(m.filtered)-1 {
			m.cursor++
		}
	case tea.KeyHome:
		m.cursor = 0
	case tea.KeyEnd:
		m.cursor = len(m.filtered) - 1

	case tea.KeyRunes:
		if m.isSearching {
			return m.handleSearchInput(msg)
		}
		switch msg.String() {
		case "/":
			m.isSearching = true
			m.search = ""
			m.cursor = 0
		case "q":
			return m, tea.Quit
		case "d":
			if len(m.filtered) > 0 {
				m.mode = viewConfirmDelete
			}
		}

	case tea.KeyEnter:
		if m.isSearching {
			m.isSearching = false
		}
	}

	return m, nil
}

func (m Model) handleSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "/":
	case "q":
		m.isSearching = false
		m.search = ""
		m.filtered = m.hosts
	case "\x7f": // Backspace
		if len(m.search) > 0 {
			m.search = m.search[:len(m.search)-1]
			m.filterHosts()
		}
	case "\r":
		m.isSearching = false
	default:
		m.search += msg.String()
		m.filterHosts()
	}
	return m, nil
}

// handleConfirmKeyMsg processes keys in confirmation view
func (m Model) handleConfirmKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		return m.deleteCurrentSelection()
	case tea.KeyRunes:
		switch msg.String() {
		case "y", "Y":
			return m.deleteCurrentSelection()
		case "n", "N", "q":
			m.mode = viewList
			m.status = "Deletion cancelled"
		}
	case tea.KeyCtrlC, tea.KeyEsc:
		m.mode = viewList
		m.status = "Deletion cancelled"
	}

	return m, nil
}

func (m Model) deleteCurrentSelection() (tea.Model, tea.Cmd) {
	hostLine := m.filtered[m.cursor]
	m.hosts = Delete(m.hosts, hostLine)
	m.filtered = Delete(m.filtered, hostLine)
	if len(m.filtered) == 0 {
		m.cursor = 0
	} else if m.cursor >= len(m.filtered) {
		m.cursor = len(m.filtered) - 1
	}
	m.mode = viewList

	// Build status message
	h, err := NewHost(hostLine)
	if err == nil {
		m.status = "Deleted " + h.DisplayName()
	} else {
		m.status = "Deleted " + hostLine
	}
	return m, saveHosts(m.hosts)
}

// filterHosts filters the host list based on search query
func (m *Model) filterHosts() {
	if m.search == "" {
		m.filtered = m.hosts
		return
	}

	m.filtered = Search(m.hosts, m.search)
	if len(m.filtered) > 0 {
		m.cursor = 0
	}
}

// Helper messages and commands

type errMsg struct{ err error }

type hostsLoadedMsg struct{ hosts []string }

func loadHosts() tea.Cmd {
	return func() tea.Msg {
		hosts, err := ReadFile()
		if err != nil {
			return errMsg{err}
		}
		return hostsLoadedMsg{hosts: hosts}
	}
}

func saveHosts(hosts []string) tea.Cmd {
	return func() tea.Msg {
		if err := SaveFile(hosts); err != nil {
			return errMsg{err}
		}
		return nil
	}
}
