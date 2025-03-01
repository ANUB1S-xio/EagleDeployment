package menu

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestInitialMainMenu(t *testing.T) {
	model := InitialMainMenu()
	if len(model.choices) != 7 {
		t.Errorf("Expected 7 choices, got %d", len(model.choices))
	}
	if model.cursor != 0 {
		t.Errorf("Expected cursor to be at position 0, got %d", model.cursor)
	}
	if model.selected != -1 {
		t.Errorf("Expected no selection, got %d", model.selected)
	}
}

func TestMainMenuModel_Update(t *testing.T) {
	model := InitialMainMenu()

	// Test moving cursor down
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model = updatedModel.(MainMenuModel)
	if model.cursor != 1 {
		t.Errorf("Expected cursor to be at position 1, got %d", model.cursor)
	}

	// Test moving cursor up
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	model = updatedModel.(MainMenuModel)
	if model.cursor != 0 {
		t.Errorf("Expected cursor to be at position 0, got %d", model.cursor)
	// Test selecting an option
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	model = updatedModel.(MainMenuModel)
	if model.selected != 0 {
		t.Errorf("Expected selected to be 0, got %d", model.selected)
	}
		t.Errorf("Expected selected to be 0, got %d", model.selected)
	}
}

func TestRunMainMenu(t *testing.T) {
	// This test is more complex as it involves running the full program
	// For simplicity, we will just check if it returns a valid selection
	selected := RunMainMenu()
	if selected < 0 || selected > 6 {
		t.Errorf("Expected a valid selection between 0 and 6, got %d", selected)
	}
}
