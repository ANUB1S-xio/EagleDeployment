package menu

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestInitialMainMenu(t *testing.T) {
	model := InitialMainMenu()
	// Updated to reflect the actual number of choices in the menu (3)
	if len(model.choices) != 3 {
		t.Errorf("Expected 3 choices, got %d", len(model.choices))
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
	}

	// Test selecting an option
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	model = updatedModel.(MainMenuModel)
	if model.selected != 0 {
		t.Errorf("Expected selected to be 0, got %d", model.selected)
	}
}

func TestRunMainMenu(t *testing.T) {
	// This test is more complex as it involves running the full program
	// For simplicity, we will just check if it returns a valid selection
	// Updated to reflect the actual valid range (-1 for exit/cancel or 0-2 for menu items)
	selected := RunMainMenu()
	if selected < -1 || selected > 2 {
		t.Errorf("Expected a valid selection between -1 and 2, got %d", selected)
	}
}
