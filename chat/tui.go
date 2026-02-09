package chat

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// 1. STYLE DEFINITIONS
var (
	headerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
	userStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true) // Green
	aiStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Bold(true) // Purple
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)

// 2. MESSAGES (Events) - these need to be udpated
type responseMsg string

// 3. THE MODEL (State)
type model struct {
	viewport  viewport.Model
	textInput textinput.Model
	spinner   spinner.Model
	history   []string
	err       error
	isReady   bool
	isLoading bool
	querier   *GraphQuerier
}

type Tui struct {
	program *tea.Program
	querier *GraphQuerier
}

func NewTui(q *GraphQuerier) *Tui {
	p := tea.NewProgram(initialModel(q), tea.WithAltScreen())

	return &Tui{
		program: p,
		querier: q,
	}
}

func (t *Tui) Init() error {
	_, err := t.program.Run()

	if err != nil {
		return err
	}

	return nil
}

func initialModel(q *GraphQuerier) model {
	ti := textinput.New()
	ti.Placeholder = "Ask about your graph..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		textInput: ti,
		spinner:   s,
		querier:   q,
		history:   []string{headerStyle.Render("--- Graph Analyzer Chat ---")},
	}
}

// 4. THE UPDATE LOOP (Logic)
func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// we need to be aware that in the case that we write in the graph (new file for example)
	// if we try to access the graph node in here while it is being written we will have a race condition
	// when do this also create a mutex in the model ClassGraph to safe-guard that

	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
		spCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyEnter:
			if m.isLoading {
				return m, nil // Don't allow multiple inputs while loading
			}
			input := m.textInput.Value()
			if strings.TrimSpace(input) == "" {
				return m, nil
			}

			// Add User message to history
			m.history = append(m.history, fmt.Sprintf("%s %s", userStyle.Render("You:"), input))
			m.textInput.Reset()
			m.isLoading = true
			m.updateViewport()

			// Trigger the "Fake" LLM/Graph response
			//
			return m, m.queryGraph(input)
		}

	case responseMsg:
		m.isLoading = false
		m.history = append(m.history, fmt.Sprintf("%s %s", aiStyle.Render("AI:"), string(msg)))
		m.updateViewport()
		return m, nil

	case tea.WindowSizeMsg:
		// Handle terminal resizing
		if !m.isReady {
			m.viewport = viewport.New(msg.Width, msg.Height-4)
			m.isReady = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 4
		}
	}

	// Update sub-components
	m.textInput, tiCmd = m.textInput.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)
	if m.isLoading {
		m.spinner, spCmd = m.spinner.Update(msg)
	}

	return m, tea.Batch(tiCmd, vpCmd, spCmd)
}

// 5. THE VIEW (Rendering)
func (m model) View() string {
	if !m.isReady {
		return "Initializing..."
	}

	var loader string
	if m.isLoading {
		loader = m.spinner.View() + " Thinking..."
	} else {
		loader = " " // Empty space when not loading
	}

	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		m.viewport.View(),
		loader,
		m.textInput.View(),
	)
}

func (m model) queryGraph(query string) tea.Cmd {
	return func() tea.Msg {
		// Replace this with your actual graph search logic later
		time.Sleep(1500 * time.Millisecond)
		res, err := m.querier.Execute(query)

		if err != nil {
			return errorStyle.Render(err.Error())
		}

		return responseMsg(res)
	}
}

// --- HELPER FUNCTIONS ---

func (m *model) updateViewport() {
	m.viewport.SetContent(strings.Join(m.history, "\n"))
	m.viewport.GotoBottom()
}
