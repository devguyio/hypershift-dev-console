/*
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package run

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/go-task/task/v3"
	taskfile "github.com/go-task/task/v3/taskfile/ast"
	"hypershift-dev-console/pkg/recipes"
	"hypershift-dev-console/pkg/tui/styles"
)

// You generally won't need this unless you're processing stuff with
// complicated ANSI escape sequences. Turn it on if you notice flickering.
//
// Also keep in mind that high performance rendering only works for programs
// that use the full size of the terminal. We're enabling that below with
// tea.EnterAltScreen().
const useHighPerformanceRenderer = false

var (
	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return styles.DefaultStyles().Title.BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.BorderStyle(b)
	}()
)

type RunResult string

type model struct {
	recipe   recipes.Recipe
	ready    bool
	viewport viewport.Model
	content  string
	runCmd   RunRecipe
}

type RunRecipe struct {
	Recipe recipes.Recipe
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

func (r RunRecipe) Run() error {
	//TODO: Implement
	// Initialize the executor
	executor := &task.Executor{
		Dir:    r.Recipe.Dir, // Directory containing your Taskfile.yml
		Stdin:  r.Stdin,
		Stdout: r.Stdout,
		Stderr: r.Stderr,
	}

	// Set up the executor
	if err := executor.Setup(); err != nil {
		fmt.Fprintf(r.Stderr, "Error setting up executor: %v\n", err)
	}

	// Define the task to run
	call := &taskfile.Call{Task: "default"}
	// Run the task
	if err := executor.RunTask(context.Background(), call); err != nil {
		fmt.Fprintf(r.Stderr, "Error running task: %v\n", err)
	}
	return nil
}

func (r RunRecipe) SetStdin(_ io.Reader) {
}

func (r RunRecipe) SetStdout(_ io.Writer) {
}

func (r RunRecipe) SetStderr(_ io.Writer) {
}

func New(recipe recipes.Recipe) tea.Model {
	var stdOut bytes.Buffer
	var stdErr bytes.Buffer
	var stdIn bytes.Buffer
	return model{
		recipe: recipe,
		runCmd: RunRecipe{
			Recipe: recipe,
			Stdin:  &stdIn,
			Stdout: &stdOut,
			Stderr: &stdErr,
		},
	}
}

func (m model) Init() tea.Cmd {

	return tea.Exec(m.runCmd, func(err error) tea.Msg {
		if err != nil {
			return nil
		}
		buf := m.runCmd.Stdout.(*bytes.Buffer)
		return RunResult(buf.String())
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if k := msg.String(); k == "ctrl+c" || k == "q" || k == "esc" {
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.viewport.HighPerformanceRendering = useHighPerformanceRenderer
			m.ready = true

			// This is only necessary for high performance rendering, which in
			// most cases you won't need.
			//
			// Render the viewport one line below the header.
			m.viewport.YPosition = headerHeight + 1
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}
		m.viewport.SetContent(m.content)

		if useHighPerformanceRenderer {
			// Render (or re-render) the whole viewport. Necessary both to
			// initialize the viewport and when the window is resized.
			//
			// This is needed for high-performance rendering only.
			cmds = append(cmds, viewport.Sync(m.viewport))
		}
	case RunResult:
		m.content = string(msg)
		m.viewport.SetContent(m.content)
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}
	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
}

func (m model) headerView() string {
	title := titleStyle.Render(m.recipe.DisplayName)
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m model) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
