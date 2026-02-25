package chat

import (
	"fmt"
	"raidline/ripple/pgk/logger"
	"strings"

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
	chatViewport viewport.Model
	watchMode    bool
	sideViewport viewport.Model // For the file change logs
	textInput    textinput.Model
	spinner      spinner.Model
	history      []string
	liveChanges  []string
	err          error
	isReady      bool
	isLoading    bool
	querier      *GraphQuerier

	width  int
	height int
}

type Tui struct {
	program *tea.Program
	watch   bool
	querier *GraphQuerier
}

// todo: this should be better structured, very whacky
func NewTui(q *GraphQuerier, watchMode bool, fileChangesChan <-chan []string) *Tui {
	m := initialModel(q, watchMode)
	p := tea.NewProgram(initialModel(q, watchMode), tea.WithAltScreen())

	//todo: in here, we need to create a goFunc that would listen for said channel and update the TUI
	// TUI should receive the channel ready to go, we might need somehere we can create it and control it
	//i think the issue is that we need to send a command when this happens for the tui,
	//also creating this here is not working, this is running all the time! being called every time
	go func() {
		logger.Info("is this thing on?")
		for _, change := range <-fileChangesChan {
			logger.Info("is this thing off?")
			m.liveChanges = append(m.liveChanges, change)
			m.updateSideViewport()
		}
	}()

	return &Tui{
		program: p,
		watch:   watchMode,
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

func initialModel(q *GraphQuerier, watchMode bool) model {
	ti := textinput.New()
	ti.Placeholder = "Ask about your graph..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		textInput:   ti,
		spinner:     s,
		querier:     q,
		watchMode:   watchMode,
		history:     []string{headerStyle.Render("--- Graph Analyzer Chat ---")},
		liveChanges: []string{headerStyle.Render("--- Project Live changes ---")},
	}
}

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
				return m, nil
			}
			input := m.textInput.Value()
			if strings.TrimSpace(input) == "" {
				return m, nil
			}

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
		m.height = msg.Height
		m.width = msg.Width
		if m.watchMode {
			sidebarWidth := msg.Width / 3
			chatWidth := msg.Width - sidebarWidth - 2 // -2 for borders/padding

			// Update viewports with new dimensions
			m.chatViewport.Width = chatWidth
			m.chatViewport.Height = msg.Height - 5 // leave room for input field

			m.sideViewport.Width = sidebarWidth
			m.sideViewport.Height = msg.Height - 2
		} else {
			if !m.isReady {
				m.chatViewport = viewport.New(msg.Width, msg.Height-4)
			} else {
				m.chatViewport.Width = msg.Width
				m.chatViewport.Height = msg.Height - 4
			}
		}
		m.isReady = true
	}

	m.textInput, tiCmd = m.textInput.Update(msg)
	m.chatViewport, vpCmd = m.chatViewport.Update(msg)
	if m.isLoading {
		m.spinner, spCmd = m.spinner.Update(msg)
	}

	return m, tea.Batch(tiCmd, vpCmd, spCmd)
}

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

	if !m.watchMode {
		return fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			m.chatViewport.View(),
			loader,
			m.textInput.View())
	}

	chatBox := lipgloss.NewStyle().
		Width(m.chatViewport.Width).
		Padding(0, 1).
		Render(fmt.Sprintf(
			"%s\n\n%s\n\n%s",
			m.chatViewport.View(),
			loader,
			m.textInput.View(),
		))

	sidebarStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true). // Left border only
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Width(m.sideViewport.Width).
		Height(m.height)

	sidebarBox := sidebarStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			headerStyle.Render("LIVE FILE CHANGES"),
			m.sideViewport.View(),
		),
	)

	return lipgloss.JoinHorizontal(lipgloss.Top, chatBox, sidebarBox)
}

func (m model) queryGraph(query string) tea.Cmd {
	return func() tea.Msg {
		res, err := m.querier.Execute(query)

		if err != nil {
			return errorStyle.Render(err.Error())
		}

		return responseMsg(res)
	}
}

// --- HELPER FUNCTIONS ---

func (m *model) updateViewport() {
	m.chatViewport.SetContent(strings.Join(m.history, "\n"))
	m.chatViewport.GotoBottom()
}

func (m *model) updateSideViewport() {
	m.sideViewport.SetContent(strings.Join(m.liveChanges, "\n"))
}
