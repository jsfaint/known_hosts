package main

import (
	"errors"
	"os"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModelInit(t *testing.T) {
	m := Model{
		hosts:    []string{"github.com ssh-rsa key"},
		filtered: []string{"github.com ssh-rsa key"},
		mode:     viewList,
	}

	cmd := m.Init()

	if cmd == nil {
		t.Error("Model.Init() should return a command")
	}
}

func TestModelUpdate(t *testing.T) {
	tests := []struct {
		name      string
		model     Model
		msg       tea.Msg
		wantMode  viewMode
		wantError bool
	}{
		{
			name: "handle key message",
			model: Model{
				hosts:    []string{"github.com ssh-rsa key"},
				filtered: []string{"github.com ssh-rsa key"},
				mode:     viewList,
			},
			msg:       tea.KeyMsg{Type: tea.KeyUp},
			wantMode:  viewList,
			wantError: false,
		},
		{
			name: "handle error message",
			model: Model{
				mode: viewList,
			},
			msg:       errMsg{err: errors.New("test error")},
			wantMode:  viewList,
			wantError: true,
		},
		{
			name: "handle hosts loaded message",
			model: Model{
				mode: viewList,
			},
			msg:       hostsLoadedMsg{hosts: []string{"github.com ssh-rsa key"}},
			wantMode:  viewList,
			wantError: false,
		},
		{
			name: "handle tick message",
			model: Model{
				mode: viewList,
			},
			msg:       TickMsg(time.Now()),
			wantMode:  viewList,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newModel, _ := tt.model.Update(tt.msg)
			model, ok := newModel.(Model)
			if !ok {
				t.Fatalf("Update() should return Model, got %T", newModel)
			}

			if model.mode != tt.wantMode {
				t.Errorf("Model.Update() mode = %v, want %v", model.mode, tt.wantMode)
			}

			if tt.wantError && model.err == nil {
				t.Error("Model.Update() should have error")
			}
		})
	}
}

func TestModelView(t *testing.T) {
	tests := []struct {
		name         string
		model        Model
		wantContains []string
	}{
		{
			name: "list view with hosts",
			model: Model{
				hosts:    []string{"github.com ssh-rsa key"},
				filtered: []string{"github.com ssh-rsa key"},
				mode:     viewList,
			},
			wantContains: []string{"Known Hosts Manager", "github.com", "Controls:"},
		},
		{
			name: "list view with empty hosts",
			model: Model{
				hosts:    []string{},
				filtered: []string{},
				mode:     viewList,
			},
			wantContains: []string{"Known Hosts Manager", "No hosts found"},
		},
		{
			name: "list view with search",
			model: Model{
				hosts:      []string{"github.com ssh-rsa key", "gitlab.com ssh-rsa key"},
				filtered:   []string{"github.com ssh-rsa key"},
				mode:       viewList,
				search:     "git",
				isSearching: false,
			},
			wantContains: []string{"Known Hosts Manager", "Filter: git", "github.com"},
		},
		{
			name: "list view in search mode",
			model: Model{
				hosts:      []string{"github.com ssh-rsa key"},
				filtered:   []string{"github.com ssh-rsa key"},
				mode:       viewList,
				search:     "git",
				isSearching: true,
			},
			wantContains: []string{"Known Hosts Manager", "Search: git", "_"},
		},
		{
			name: "confirm delete view",
			model: Model{
				hosts:    []string{"github.com ssh-rsa key"},
				filtered: []string{"github.com ssh-rsa key"},
				cursor:   0,
				mode:     viewConfirmDelete,
			},
			wantContains: []string{"Confirm Deletion", "Delete this host?", "github.com", "Press 'y' to confirm"},
		},
		{
			name: "error view",
			model: Model{
				err:  errors.New("test error"),
				mode: viewList,
			},
			wantContains: []string{"Error: test error", "Press 'q' to quit"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view := tt.model.View()

			for _, expected := range tt.wantContains {
				if !contains(view, expected) {
					t.Errorf("Model.View() should contain %q, got:\n%s", expected, view)
				}
			}
		})
	}
}

func TestHandleListKeyMsg(t *testing.T) {
	tests := []struct {
		name         string
		model        Model
		msg          tea.KeyMsg
		wantCursor   int
		wantMode     viewMode
		wantSearch   string
		wantFiltered []string
	}{
		{
			name: "move cursor up",
			model: Model{
				filtered: []string{"host1", "host2", "host3"},
				cursor:   1,
				mode:     viewList,
			},
			msg:        tea.KeyMsg{Type: tea.KeyUp},
			wantCursor: 0,
			wantMode:   viewList,
		},
		{
			name: "move cursor down",
			model: Model{
				filtered: []string{"host1", "host2", "host3"},
				cursor:   0,
				mode:     viewList,
			},
			msg:        tea.KeyMsg{Type: tea.KeyDown},
			wantCursor: 1,
			wantMode:   viewList,
		},
		{
			name: "move to home",
			model: Model{
				filtered: []string{"host1", "host2", "host3"},
				cursor:   2,
				mode:     viewList,
			},
			msg:        tea.KeyMsg{Type: tea.KeyHome},
			wantCursor: 0,
			wantMode:   viewList,
		},
		{
			name: "move to end",
			model: Model{
				filtered: []string{"host1", "host2", "host3"},
				cursor:   0,
				mode:     viewList,
			},
			msg:        tea.KeyMsg{Type: tea.KeyEnd},
			wantCursor: 2,
			wantMode:   viewList,
		},
		{
			name: "start search with /",
			model: Model{
				hosts:    []string{"github.com ssh-rsa key", "gitlab.com ssh-rsa key"},
				filtered: []string{"github.com ssh-rsa key", "gitlab.com ssh-rsa key"},
				mode:     viewList,
			},
			msg:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}},
			wantCursor: 0,
			wantMode:   viewList,
			wantSearch: "",
		},
		{
			name: "quit with q",
			model: Model{
				filtered: []string{"host1"},
				mode:     viewList,
			},
			msg:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}},
			wantCursor: 0,
			wantMode:   viewList,
		},
		{
			name: "enter delete mode with d",
			model: Model{
				filtered: []string{"host1"},
				mode:     viewList,
			},
			msg:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}},
			wantCursor: 0,
			wantMode:   viewConfirmDelete,
		},
		{
			name: "type in search mode",
			model: Model{
				hosts:      []string{"github.com ssh-rsa key", "gitlab.com ssh-rsa key"},
				filtered:   []string{"github.com ssh-rsa key", "gitlab.com ssh-rsa key"},
				search:     "g",
				isSearching: true,
				mode:       viewList,
			},
			msg:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}},
			wantCursor: 0,
			wantMode:   viewList,
			wantSearch: "gi",
		},
		{
			name: "backspace in search mode",
			model: Model{
				hosts:      []string{"github.com ssh-rsa key", "gitlab.com ssh-rsa key"},
				filtered:   []string{"github.com ssh-rsa key", "gitlab.com ssh-rsa key"},
				search:     "gi",
				isSearching: true,
				mode:       viewList,
			},
			msg:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{127}}, // Backspace
			wantCursor: 0,
			wantMode:   viewList,
			wantSearch: "g",
		},
		{
			name: "exit search with q",
			model: Model{
				hosts:      []string{"github.com ssh-rsa key", "gitlab.com ssh-rsa key"},
				filtered:   []string{"github.com ssh-rsa key", "gitlab.com ssh-rsa key"},
				search:     "gi",
				isSearching: true,
				mode:       viewList,
			},
			msg:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}},
			wantCursor: 0,
			wantMode:   viewList,
			wantSearch: "",
		},
		{
			name: "unknown key in non-search mode",
			model: Model{
				filtered: []string{"host1"},
				mode:     viewList,
			},
			msg:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}},
			wantCursor: 0,
			wantMode:   viewList,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newModel, _ := tt.model.handleListKeyMsg(tt.msg)
			model, ok := newModel.(Model)
			if !ok {
				t.Fatalf("handleListKeyMsg() should return Model, got %T", newModel)
			}

			if model.cursor != tt.wantCursor {
				t.Errorf("handleListKeyMsg() cursor = %v, want %v", model.cursor, tt.wantCursor)
			}

			if model.mode != tt.wantMode {
				t.Errorf("handleListKeyMsg() mode = %v, want %v", model.mode, tt.wantMode)
			}

			if model.search != tt.wantSearch {
				t.Errorf("handleListKeyMsg() search = %v, want %v", model.search, tt.wantSearch)
			}
		})
	}
}

func TestHandleConfirmKeyMsg(t *testing.T) {
	tests := []struct {
		name         string
		model        Model
		msg          tea.KeyMsg
		wantMode     viewMode
		wantFiltered int
	}{
		{
			name: "confirm deletion with y",
			model: Model{
				hosts:    []string{"github.com ssh-rsa key", "gitlab.com ssh-rsa key"},
				filtered: []string{"github.com ssh-rsa key"},
				cursor:   0,
				mode:     viewConfirmDelete,
			},
			msg:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}},
			wantMode:     viewList,
			wantFiltered: 1, // One host deleted from filtered
		},
		{
					name: "confirm deletion with Y",
					model: Model{
						hosts:    []string{"github.com ssh-rsa key"},
						filtered: []string{"github.com ssh-rsa key"},
						cursor:   0,
						mode:     viewConfirmDelete,
					},
					msg:         tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'Y'}},
					wantMode:    viewList,
					wantFiltered: 1, // Delete function matches hostPart, not full line
				},
		{
			name: "cancel deletion with n",
			model: Model{
				filtered: []string{"github.com ssh-rsa key"},
				cursor:   0,
				mode:     viewConfirmDelete,
			},
			msg:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}},
			wantMode:     viewList,
			wantFiltered: 1,
		},
		{
			name: "cancel deletion with N",
			model: Model{
				filtered: []string{"github.com ssh-rsa key"},
				cursor:   0,
				mode:     viewConfirmDelete,
			},
			msg:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}},
			wantMode:     viewList,
			wantFiltered: 1,
		},
		{
			name: "cancel deletion with q",
			model: Model{
				filtered: []string{"github.com ssh-rsa key"},
				cursor:   0,
				mode:     viewConfirmDelete,
			},
			msg:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}},
			wantMode:     viewList,
			wantFiltered: 1,
		},
		{
			name: "cancel deletion with Esc",
			model: Model{
				filtered: []string{"github.com ssh-rsa key"},
				cursor:   0,
				mode:     viewConfirmDelete,
			},
			msg:          tea.KeyMsg{Type: tea.KeyEsc},
			wantMode:     viewList,
			wantFiltered: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newModel, _ := tt.model.handleConfirmKeyMsg(tt.msg)
			model, ok := newModel.(Model)
			if !ok {
				t.Fatalf("handleConfirmKeyMsg() should return Model, got %T", newModel)
			}

			if model.mode != tt.wantMode {
				t.Errorf("handleConfirmKeyMsg() mode = %v, want %v", model.mode, tt.wantMode)
			}

			if len(model.filtered) != tt.wantFiltered {
				t.Errorf("handleConfirmKeyMsg() filtered length = %v, want %v", len(model.filtered), tt.wantFiltered)
			}
		})
	}
}

func TestFilterHosts(t *testing.T) {
	tests := []struct {
		name       string
		model      Model
		search     string
		wantLength int
	}{
		{
			name: "empty search returns all",
			model: Model{
				hosts: []string{"github.com ssh-rsa key", "gitlab.com ssh-rsa key"},
			},
			search:     "",
			wantLength: 2,
		},
		{
			name: "search filters hosts",
			model: Model{
				hosts: []string{"github.com ssh-rsa key", "gitlab.com ssh-rsa key"},
			},
			search:     "github",
			wantLength: 1,
		},
		{
			name: "search with no matches",
			model: Model{
				hosts: []string{"github.com ssh-rsa key", "gitlab.com ssh-rsa key"},
			},
			search:     "bitbucket",
			wantLength: 0,
		},
		{
			name: "search with partial match",
			model: Model{
				hosts: []string{"github.com ssh-rsa key", "gitlab.com ssh-rsa key"},
			},
			search:     "git",
			wantLength: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.model
			m.search = tt.search
			m.filterHosts()

			if len(m.filtered) != tt.wantLength {
				t.Errorf("filterHosts() filtered length = %v, want %v", len(m.filtered), tt.wantLength)
			}

			if len(m.filtered) > 0 && m.cursor != 0 {
				t.Errorf("filterHosts() should reset cursor to 0, got %v", m.cursor)
			}
		})
	}
}

func TestLoadHosts(t *testing.T) {
	t.Run("successful load", func(t *testing.T) {
		tmpDir := t.TempDir()
		sshDir := tmpDir + "/.ssh"
		testFile := sshDir + "/known_hosts"

		// Create .ssh directory
		if err := os.MkdirAll(sshDir, 0755); err != nil {
			t.Fatalf("Failed to create .ssh directory: %v", err)
		}

		oldHome := os.Getenv("HOME")
		os.Setenv("HOME", tmpDir)
		defer os.Setenv("HOME", oldHome)

		// Create test file
		testContent := "github.com ssh-rsa key1\ngitlab.com ssh-rsa key2\n"
		if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		cmd := loadHosts()
		msg := cmd()

		loadedMsg, ok := msg.(hostsLoadedMsg)
		if !ok {
			t.Fatalf("loadHosts() should return hostsLoadedMsg, got %T", msg)
		}

		if len(loadedMsg.hosts) != 2 {
			t.Errorf("loadHosts() should load 2 hosts, got %d", len(loadedMsg.hosts))
		}
	})

	t.Run("load with error", func(t *testing.T) {
		tmpDir := t.TempDir()

		oldHome := os.Getenv("HOME")
		os.Setenv("HOME", tmpDir)
		defer os.Setenv("HOME", oldHome)

		// Don't create .ssh directory, this should cause an error
		cmd := loadHosts()
		msg := cmd()

		errMsg, ok := msg.(errMsg)
		if !ok {
			t.Fatalf("loadHosts() should return errMsg on failure, got %T", msg)
		}

		if errMsg.err == nil {
			t.Error("loadHosts() should return error when file doesn't exist")
		}
	})
}

func TestSaveHosts(t *testing.T) {
	tmpDir := t.TempDir()
	sshDir := tmpDir + "/.ssh"
	testFile := sshDir + "/known_hosts"

	// Create .ssh directory
	if err := os.MkdirAll(sshDir, 0755); err != nil {
		t.Fatalf("Failed to create .ssh directory: %v", err)
	}

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	hosts := []string{"github.com ssh-rsa key", "gitlab.com ssh-rsa key"}

	cmd := saveHosts(hosts)
	msg := cmd()

	// saveHosts returns nil on success
	if msg != nil {
		errMsg, ok := msg.(errMsg)
		if ok {
			t.Errorf("saveHosts() should not return error, got: %v", errMsg.err)
		}
	}

	// Verify file was created
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	contentStr := string(content)
	for _, host := range hosts {
		if !contains(contentStr, host) {
			t.Errorf("saveHosts() should contain host: %s", host)
		}
	}
}

func TestTick(t *testing.T) {
	cmd := tick()

	if cmd == nil {
		t.Error("tick() should return a command")
	}

	// Execute the command to get TickMsg
	msg := cmd()
	_, ok := msg.(TickMsg)
	if !ok {
		t.Errorf("tick() should return TickMsg, got %T", msg)
	}
}

func TestErrMsg(t *testing.T) {
	testErr := errors.New("test error")
	errMsg := errMsg{err: testErr}

	errStr := errMsg.Error()
	if errStr != testErr.Error() {
		t.Errorf("errMsg.Error() = %v, want %v", errStr, testErr.Error())
	}
}

func TestRenderListWithIPAndName(t *testing.T) {
	model := Model{
		filtered: []string{"myserver,192.168.1.1 ssh-rsa key"},
		cursor:   0,
		mode:     viewList,
	}

	view := model.View()

	if !contains(view, "myserver, 192.168.1.1") {
		t.Error("renderList() should display host with name and IP")
	}
}

func TestRenderListWithNameOnly(t *testing.T) {
	model := Model{
		filtered: []string{"github.com ssh-rsa key"},
		cursor:   0,
		mode:     viewList,
	}

	view := model.View()

	if !contains(view, "github.com") {
		t.Error("renderList() should display host with name only")
	}
}

func TestRenderListWithIPOnly(t *testing.T) {
	model := Model{
		filtered: []string{"192.168.1.1 ssh-rsa key"},
		cursor:   0,
		mode:     viewList,
	}

	view := model.View()

	if !contains(view, "192.168.1.1") {
		t.Error("renderList() should display host with IP only")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findInString(s, substr))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
