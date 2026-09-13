package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// newTestWebServer wires a webServer to a fresh store in a temp dir, seeding
// a named tournament with the given players, pairings and results.
func newTestWebServer(t *testing.T, seed func(t *Tournament)) *webServer {
	t.Helper()
	s, err := openStore(t.TempDir() + "/suizo.json")
	if err != nil {
		t.Fatalf("openStore: %v", err)
	}
	if seed != nil {
		if err := s.update(func(t *Tournament) error { seed(t); return nil }); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	return &webServer{store: s}
}

// postForm issues a POST with a urlencoded form body and Content-Type set,
// the way a browser form submission arrives.
func postForm(h http.Handler, path string, form url.Values) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestWebStandings(t *testing.T) {
	w := newTestWebServer(t, func(t *Tournament) {
		for _, name := range []string{"Ana", "Beto", "Carla"} {
			_, _ = t.addPlayer(name)
		}
		_, _, _ = t.StartRound()
		_ = t.SetResult(1, 1, "1-0")
	})
	rec := httptest.NewRecorder()
	w.handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{"Ana", "Beto", "Carla", "Standings", "refresh"} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %q", want)
		}
	}
}

func TestWebStandingsEmpty(t *testing.T) {
	w := newTestWebServer(t, nil)
	rec := httptest.NewRecorder()
	w.handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "No players yet") {
		t.Errorf("body missing empty-state text")
	}
}

func TestWebRoundsShowsBoardsAndForm(t *testing.T) {
	w := newTestWebServer(t, func(t *Tournament) {
		for _, name := range []string{"Ana", "Beto", "Carla"} {
			_, _ = t.addPlayer(name)
		}
		_, _, _ = t.StartRound() // 1 game + 1 bye, all pending
	})
	rec := httptest.NewRecorder()
	w.handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/rounds", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{`action="/results"`, "gets a bye", "Round 1"} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %q", want)
		}
	}
}

func TestWebResultsPost(t *testing.T) {
	t.Run("legal result round-trips", func(t *testing.T) {
		w := newTestWebServer(t, func(t *Tournament) {
			for _, name := range []string{"Ana", "Beto", "Carla"} {
				_, _ = t.addPlayer(name)
			}
			_, _, _ = t.StartRound()
		})
		form := url.Values{"round": {"1"}, "board": {"1"}, "result": {"0.5-0.5"}}
		rec := postForm(w.handler(), "/results", form)
		if rec.Code != http.StatusSeeOther {
			t.Fatalf("status = %d, want 303", rec.Code)
		}
		// The result must be persisted: a fresh GET shows it, no form left.
		rec = httptest.NewRecorder()
		w.handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/rounds", nil))
		body := rec.Body.String()
		if !strings.Contains(body, "0.5-0.5") || strings.Contains(body, `action="/results"`) {
			t.Errorf("result not rendered as done: %s", body)
		}
	})

	t.Run("illegal result rejected", func(t *testing.T) {
		w := newTestWebServer(t, func(t *Tournament) {
			for _, name := range []string{"Ana", "Beto"} {
				_, _ = t.addPlayer(name)
			}
			_, _, _ = t.StartRound()
		})
		for _, bad := range []string{"1:0", "resigns"} {
			form := url.Values{"round": {"1"}, "board": {"1"}, "result": {bad}}
			rec := postForm(w.handler(), "/results", form)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("result %q: status = %d, want 400", bad, rec.Code)
			}
		}
		// And nothing was written.
		tt, err := w.store.load()
		if err != nil {
			t.Fatal(err)
		}
		if got := tt.Rounds[0].Matches[0].Result; got != "" {
			t.Errorf("result = %q, want empty", got)
		}
	})

	t.Run("non-integer board rejected", func(t *testing.T) {
		w := newTestWebServer(t, func(t *Tournament) {
			for _, name := range []string{"Ana", "Beto"} {
				_, _ = t.addPlayer(name)
			}
			_, _, _ = t.StartRound()
		})
		form := url.Values{"round": {"1"}, "board": {"x"}, "result": {"1-0"}}
		rec := postForm(w.handler(), "/results", form)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", rec.Code)
		}
	})
}

func TestWebPlayersPageAndAdd(t *testing.T) {
	w := newTestWebServer(t, nil)

	t.Run("add and redirect", func(t *testing.T) {
		form := url.Values{"name": {"Ana Garc\u00eda"}}
		rec := postForm(w.handler(), "/players", form)
		if rec.Code != http.StatusSeeOther {
			t.Fatalf("status = %d, want 303", rec.Code)
		}
		tt, err := w.store.load()
		if err != nil {
			t.Fatal(err)
		}
		if len(tt.Players) != 1 || tt.Players[0].Name != "Ana Garc\u00eda" {
			t.Fatalf("players = %+v, want one Ana", tt.Players)
		}
	})

	t.Run("GET lists player", func(t *testing.T) {
		rec := httptest.NewRecorder()
		w.handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/players", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), `action="/players"`) {
			t.Errorf("body missing add form")
		}
		if !strings.Contains(rec.Body.String(), "Ana Garc\u00eda") {
			t.Errorf("body missing escaped player name: %s", rec.Body.String())
		}
	})

	t.Run("empty name rejected", func(t *testing.T) {
		form := url.Values{"name": {""}}
		rec := postForm(w.handler(), "/players", form)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", rec.Code)
		}
	})
}

func TestWebHTMLEscaping(t *testing.T) {
	w := newTestWebServer(t, func(t *Tournament) {
		_, _ = t.addPlayer(`<script>alert(1)</script>`)
	})
	for _, path := range []string{"/", "/players"} {
		rec := httptest.NewRecorder()
		w.handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if strings.Contains(rec.Body.String(), "<script>") {
			t.Errorf("%s: unescaped script tag in output", path)
		}
	}
}
