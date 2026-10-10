// Package tui provides the scrollable Bubble Tea viewer for beautified JSON,
// with chroma syntax highlighting and a Charm viewport gutter for line numbers.
package tui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/atotto/clipboard"

	"github.com/almerleoalmazan/json-beautify/internal/highlight"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))
	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10"))
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
	gutterStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)

type copiedMsg struct{}

// Options configures the viewer.
type Options struct {
	// Theme is the chroma style name (e.g. "dracula").
	Theme string
	// Color enables syntax highlighting.
	Color bool
	// LineNumbers enables the gutter.
	LineNumbers bool
}

// Model is a pager: viewport + header/footer + copy-to-clipboard.
// raw is the plain pretty JSON (used for clipboard); display is what the
// viewport shows (optionally highlighted).
type Model struct {
	raw         string
	display     string
	theme       string
	color       bool
	lineNumbers bool
	viewport    viewport.Model
	ready       bool
	copied      string
}

// New returns a Model holding the pretty JSON to display.
func New(raw string, opt Options) Model {
	if opt.Theme == "" {
		opt.Theme = highlight.DefaultStyle
	}
	m := Model{raw: raw, theme: opt.Theme, color: opt.Color, lineNumbers: opt.LineNumbers}
	m.display = m.render()
	return m
}

// Run launches the TUI and blocks until quit.
func Run(raw string, opt Options) error {
	p := tea.NewProgram(New(raw, opt))
	_, err := p.Run()
	return err
}

func (m *Model) render() string {
	if m.color {
		return highlight.HighlightJSON(m.raw, m.theme)
	}
	return m.raw
}

func lineNumberGutter() viewport.GutterFunc {
	return func(info viewport.GutterContext) string {
		if info.Soft {
			return gutterStyle.Render("      │ ")
		}
		if info.Index >= info.TotalLines {
			return gutterStyle.Render("    ~ │ ")
		}
		return gutterStyle.Render(fmt.Sprintf("%5d │ ", info.Index+1))
	}
}

func (m *Model) applyDisplay() {
	m.display = m.render()
	if m.ready {
		m.viewport.SetContent(m.display)
	}
}

func (m *Model) applyGutter() {
	if m.ready {
		if m.lineNumbers {
			m.viewport.LeftGutterFunc = lineNumberGutter()
		} else {
			m.viewport.LeftGutterFunc = nil
		}
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "y", "c", "Y":
			// Copy raw plain JSON: no ANSI colors, no line numbers.
			if err := clipboard.WriteAll(m.raw); err != nil {
				m.copied = fmt.Sprintf("copy failed: %v", err)
			} else {
				m.copied = "copied to clipboard!"
			}
			return m, tea.Tick(2*time.Second, func(time.Time) tea.Msg {
				return copiedMsg{}
			})
		case "l", "L":
			m.lineNumbers = !m.lineNumbers
			m.applyGutter()
			return m, nil
		case "s", "S":
			m.color = !m.color
			m.applyDisplay()
			return m, nil
		}
	case copiedMsg:
		m.copied = ""
	case tea.WindowSizeMsg:
		headerH := lipgloss.Height(m.headerView())
		footerH := lipgloss.Height(m.footerView())
		if !m.ready {
			m.viewport = viewport.New(
				viewport.WithWidth(msg.Width),
				viewport.WithHeight(msg.Height-headerH-footerH),
			)
			m.viewport.YPosition = headerH
			if m.lineNumbers {
				m.viewport.LeftGutterFunc = lineNumberGutter()
			}
			m.viewport.SetContent(m.display)
			m.ready = true
		} else {
			m.viewport.SetWidth(msg.Width)
			m.viewport.SetHeight(msg.Height - headerH - footerH)
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() tea.View {
	if !m.ready {
		return tea.NewView("\n  Initializing...")
	}
	v := tea.NewView(fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView()))
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

func (m Model) headerView() string {
	title := titleStyle.Render("json-beautify")
	status := ""
	if m.copied != "" {
		status = statusStyle.Render("  ✓ " + m.copied)
	}
	line := strings.Repeat("─", max(0, m.viewport.Width()-lipgloss.Width(title)-lipgloss.Width(status)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, status, line)
}

func (m Model) footerView() string {
	info := fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100)
	help := helpStyle.Render(" ↑/↓ scroll • ←/→ horiz • l numbers • s color • y copy • q quit ")
	line := strings.Repeat("─", max(0, m.viewport.Width()-lipgloss.Width(info)-lipgloss.Width(help)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, help, info)
}
