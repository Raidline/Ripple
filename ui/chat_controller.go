package ui

import (
	"fmt"
	"raidline/ripple/domain"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TuiChat struct {
	program *tea.Program
	state   *domain.StateCoordinator
}

// 1. STYLE DEFINITIONS
var (
	headerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
	userStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true) // Green
	aiStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Bold(true) // Purple
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)

const liveChangesHeader = "LIVE FILE CHANGES"

// 2. MESSAGES (Events) - these need to be udpated
type responseMsg string
type liveChangesMsg []string

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

	width  int
	height int
}

func NewTui(stateControl *domain.StateCoordinator) *TuiChat {
	m := initialModel(stateControl.IsLiveWatchMode())
	p := tea.NewProgram(m, tea.WithAltScreen())

	return &TuiChat{
		program: p,
		state:   stateControl,
	}
}

func (t *TuiChat) Init() error {
	_, err := t.program.Run()

	if err != nil {
		return err
	}

	if t.state.IsLiveWatchMode() {
		t.state.CreateGoroutine(func() {
			for {
				select {
				case <-t.state.Ctx.Done():
					return
				case changes, ok := <-t.state.LiveChangeChan:
					if !ok {
						return
					}

					impactsMsg := make([]string, len(changes.Impacts)+1)

					impactsMsg[0] = fmt.Sprintf("Causing file : [%s], Impacts: \n", changes.CausingFile)

					for i, v := range changes.Impacts {
						impactsMsg[i+1] = fmt.Sprintf("├─▶ %s\n", v)
					}

					t.program.Send(liveChangesMsg(impactsMsg))
				}
			}
		})
	}

	return nil
}

func initialModel(watchMode bool) model {
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
		watchMode:   watchMode,
		history:     []string{headerStyle.Render("--- Graph Analyzer Chat ---")},
		liveChanges: []string{headerStyle.Render("")},
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

			return m, m.queryGraph(input)
		}

	case responseMsg:
		m.isLoading = false
		m.history = append(m.history, fmt.Sprintf("%s %s", aiStyle.Render("AI:"), string(msg)))
		m.updateViewport()
		return m, nil
	case liveChangesMsg:
		m.liveChanges = append(m.liveChanges, msg...)
		m.updateSideViewport()
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
			headerStyle.Render(liveChangesHeader),
			m.sideViewport.View(),
		),
	)

	return lipgloss.JoinHorizontal(lipgloss.Top, chatBox, sidebarBox)
}

func (m model) queryGraph(query string) tea.Cmd {
	return func() tea.Msg {
		// res, err := m.querier.Execute(query)

		// if err != nil {
		// 	return errorStyle.Render(err.Error())
		// }

		return responseMsg(fmt.Sprintf("Answer for : [%s]", query))
	}
}

// --- HELPER FUNCTIONS ---

func (m *model) updateViewport() {
	m.chatViewport.SetContent(strings.Join(m.history, "\n"))
	m.chatViewport.GotoBottom()
}

func (m *model) updateSideViewport() {
	m.sideViewport.SetContent(strings.Join(m.liveChanges, "\n"))
	m.sideViewport.GotoTop()
}
