package main

import "testing"

func TestAddPlayer(t *testing.T) {
	tests := []struct {
		name      string
		existing  []string
		add       string
		wantID    string
		wantErr   error
		wantNames []string
	}{
		{
			name:      "first player gets p1",
			add:       "Ana",
			wantID:    "p1",
			wantNames: []string{"Ana"},
		},
		{
			name:      "ids increment",
			existing:  []string{"Ana"},
			add:       "Beto",
			wantID:    "p2",
			wantNames: []string{"Ana", "Beto"},
		},
		{
			name:      "empty name rejected",
			add:       "",
			wantErr:   errEmptyName,
			wantNames: []string{},
		},
		{
			name:      "duplicate names are allowed",
			existing:  []string{"Ana"},
			add:       "Ana",
			wantID:    "p2",
			wantNames: []string{"Ana", "Ana"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trn := &Tournament{Name: "test"}
			for _, n := range tt.existing {
				if _, err := trn.addPlayer(n); err != nil {
					t.Fatalf("addPlayer(%q) setup: %v", n, err)
				}
			}

			got, err := trn.addPlayer(tt.add)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("addPlayer(%q) error = %v, want %v", tt.add, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("addPlayer(%q) unexpected error: %v", tt.add, err)
			}
			if got.ID != tt.wantID {
				t.Errorf("addPlayer(%q).ID = %q, want %q", tt.add, got.ID, tt.wantID)
			}
			if len(trn.Players) != len(tt.wantNames) {
				t.Fatalf("tournament has %d players, want %d", len(trn.Players), len(tt.wantNames))
			}
			for i, want := range tt.wantNames {
				if trn.Players[i].Name != want {
					t.Errorf("Players[%d].Name = %q, want %q", i, trn.Players[i].Name, want)
				}
			}
		})
	}
}

func TestNextPlayerID(t *testing.T) {
	tests := []struct {
		name    string
		idSetup []string
		want    string
	}{
		{"empty tournament starts at p1", nil, "p1"},
		{"sequential ids", []string{"p1", "p2"}, "p3"},
		{"largest suffix wins regardless of order", []string{"p10", "p2"}, "p11"},
		{"non-numeric ids are ignored", []string{"x"}, "p1"},
		{"beyond ten", []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8", "p9", "p10"}, "p11"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trn := &Tournament{Name: "test"}
			for _, id := range tt.idSetup {
				trn.Players = append(trn.Players, Player{ID: id, Name: "player " + id})
			}
			if got := trn.nextPlayerID(); got != tt.want {
				t.Errorf("nextPlayerID() with ids %v = %q, want %q", tt.idSetup, got, tt.want)
			}
		})
	}
}

func TestPlayerLookup(t *testing.T) {
	trn := &Tournament{Name: "test", Players: []Player{
		{ID: "p1", Name: "Ana"},
		{ID: "p2", Name: "Beto"},
	}}

	tests := []struct {
		name    string
		id      string
		wantNil bool
		want    string
	}{
		{"found first", "p1", false, "Ana"},
		{"found second", "p2", false, "Beto"},
		{"missing id", "p9", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := trn.player(tt.id)
			if tt.wantNil {
				if got != nil {
					t.Errorf("player(%q) = %+v, want nil", tt.id, got)
				}
				return
			}
			if got == nil {
				t.Fatalf("player(%q) = nil, want found", tt.id)
			}
			if got.Name != tt.want {
				t.Errorf("player(%q).Name = %q, want %q", tt.id, got.Name, tt.want)
			}
		})
	}
}
