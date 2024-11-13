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

package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"hypershift-dev-console/pkg/config"
	"hypershift-dev-console/pkg/tui/home"
	"hypershift-dev-console/pkg/tui/keys"
	"hypershift-dev-console/pkg/tui/recipes"
	"hypershift-dev-console/pkg/tui/recipes/run"
	"hypershift-dev-console/pkg/tui/styles"
)

type uiState int

const (
	homeUI uiState = iota
	recipesUI
	runRecipeUI
	unknown
)

type Model struct {
	home       tea.Model
	recipes    tea.Model
	runRecipe  tea.Model
	keyMap     *keys.KeyMap
	currentUI  uiState
	styles     styles.Styles
	windowSize tea.WindowSizeMsg
	cfg        *config.Config
}

func NewModel(cfg *config.Config) Model {

	return Model{
		home:      home.New(),
		currentUI: homeUI,
		cfg:       cfg,
	}
}

func (m *Model) updateKeybindins() {

	switch m.currentUI {
	case homeUI:
		m.keyMap.Enter.SetEnabled(true)
		m.keyMap.Create.SetEnabled(true)
		m.keyMap.Delete.SetEnabled(true)

		m.keyMap.Cancel.SetEnabled(false)

	default:
		m.keyMap.Enter.SetEnabled(true)
		m.keyMap.Create.SetEnabled(true)
		m.keyMap.Delete.SetEnabled(true)
		m.keyMap.Cancel.SetEnabled(false)
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowSize = msg
	case home.SelectMessage:
		m.currentUI = recipesUI
		m.recipes = recipes.New(m.cfg)
		cmds = append(cmds, m.recipes.Init())
	case recipes.SelectMessage:
		m.currentUI = runRecipeUI
		m.runRecipe = run.New(msg.Recipe)
		cmds = append(cmds, m.runRecipe.Init())
	}

	switch m.currentUI {
	case homeUI:
		m.home, cmd = m.home.Update(msg)
	case recipesUI:
		m.recipes, cmd = m.recipes.Update(msg)
	case runRecipeUI:
		m.runRecipe, cmd = m.runRecipe.Update(msg)
	}

	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	switch m.currentUI {
	case homeUI:
		return m.home.View()
	case recipesUI:
		return m.recipes.View()
	case runRecipeUI:
		return m.runRecipe.View()
	default:
		return ""
	}
}
