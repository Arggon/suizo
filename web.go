package main

import (
	"flag"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strconv"
)

// The web UI is a thin, server-side-rendered view over the same JSON state the
// CLI uses: every request re-reads the store, mutations go through store.update
// so the advisory lock serializes them against CLI writes. Stdlib only, zero
// JavaScript; pages self-refresh via a meta refresh tag.

// defaultAddr is where `suizo serve` listens when neither --addr nor
// SUIZO_ADDR say otherwise. Loopback only: the UI is a local convenience,
// never a network service.
const defaultAddr = "127.0.0.1:8080"

// page is the data every page shares: the tournament document plus nav state.
type page struct {
	Title      string
	Tournament *Tournament
	Data       any
}

var tmplFuncs = template.FuncMap{
	// pts renders a score without a trailing .0 (1, 1.5), matching the CLI.
	"pts": formatPoints,
	// name resolves a player id to its display name, falling back to the id.
	"name": func(t *Tournament, id string) string {
		if p := t.player(id); p != nil {
			return p.Name
		}
		return id
	},
}

var tmpl = template.Must(template.New("page").Funcs(tmplFuncs).Parse(
	`{{define "head"}}<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta http-equiv="refresh" content="5">
<title>{{.Title}} - suizo</title>
</head>
<body>
<nav>
<a href="/">Standings</a> |
<a href="/rounds">Rounds</a> |
<a href="/players">Players</a>
</nav>
<h1>{{.Title}}</h1>
{{end}}` +
		`{{define "standings"}}{{template "head" .}}
{{if not .Tournament.Players}}<p>No players yet.</p>{{else}}<table border="1">
<tr><th>Rk</th><th>Name</th><th>Pts</th><th>Played</th></tr>
{{range .Data}}<tr><td>{{.Rank}}</td><td>{{.Name}}</td><td>{{pts .Points}}</td><td>{{.Played}}</td></tr>
{{end}}</table>{{end}}
</body>
</html>{{end}}` +
		`{{define "rounds"}}{{template "head" .}}
<h2>Round {{.Data.Number}}</h2>
{{if not .Data.Boards}}<p>No boards.</p>{{else}}<ul>
{{range .Data.Boards}}<li>
{{if .Match.IsBye}}{{name $.Tournament .Match.White}} gets a bye{{else}}{{name $.Tournament .Match.White}} vs {{name $.Tournament .Match.Black}}{{end}}
{{if .Match.Result}}<strong>{{.Match.Result}}</strong>{{else if not .Match.IsBye}}
<form method="post" action="/results">
<input type="hidden" name="round" value="{{$.Data.Number}}">
<input type="hidden" name="board" value="{{.Board}}">
<select name="result">
<option value="1-0">1-0</option>
<option value="0-1">0-1</option>
<option value="0.5-0.5">0.5-0.5</option>
</select>
<button type="submit">Report</button>
</form>{{end}}
</li>
{{end}}</ul>{{end}}
</body>
</html>{{end}}` +
		`{{define "players"}}{{template "head" .}}
{{if not .Tournament.Players}}<p>No players yet.</p>{{else}}<ul>
{{range .Tournament.Players}}<li>{{.ID}}: {{.Name}}</li>
{{end}}</ul>{{end}}
<form method="post" action="/players">
<input type="text" name="name" placeholder="Player name" required>
<button type="submit">Add player</button>
</form>
</body>
</html>{{end}}`))

// latestRoundPage is the view model for the rounds page: the round number
// plus the per-board status rows for that round.
type latestRoundPage struct {
	Number int
	Boards []BoardStatus
}

// webServer serves the tournament UI from a store.
type webServer struct {
	store *store
}

// handler builds the route table. Loopback binding and the fresh-read-per-
// request rule are enforced by serve(), not here.
func (w *webServer) handler() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", w.handleStandings)
	mux.HandleFunc("GET /rounds", w.handleRounds)
	mux.HandleFunc("GET /players", w.handlePlayers)
	mux.HandleFunc("POST /results", w.handleResults)
	mux.HandleFunc("POST /players", w.handlePlayersPost)
	return mux
}

// load reads the tournament fresh from disk for this request: the page always
// reflects the state file, including changes made by the CLI in parallel.
func (w *webServer) load(rsp http.ResponseWriter) (*Tournament, bool) {
	t, err := w.store.load()
	if err != nil {
		http.Error(rsp, err.Error(), http.StatusInternalServerError)
		return nil, false
	}
	return t, true
}

func (w *webServer) handleStandings(rsp http.ResponseWriter, req *http.Request) {
	t, ok := w.load(rsp)
	if !ok {
		return
	}
	render(rsp, "standings", "Standings", page{Title: "Standings", Tournament: t, Data: t.Standings()})
}

func (w *webServer) handleRounds(rsp http.ResponseWriter, req *http.Request) {
	t, ok := w.load(rsp)
	if !ok {
		return
	}
	boards, err := t.RoundStatus(0) // 0 selects the latest round
	if err != nil {
		render(rsp, "rounds", "Round", page{Title: "Rounds", Tournament: t, Data: latestRoundPage{}})
		return
	}
	num := 0
	if n := len(t.Rounds); n > 0 {
		num = t.Rounds[n-1].Number
	}
	render(rsp, "rounds", "Rounds", page{Title: "Rounds", Tournament: t,
		Data: latestRoundPage{Number: num, Boards: boards}})
}

func (w *webServer) handlePlayers(rsp http.ResponseWriter, req *http.Request) {
	t, ok := w.load(rsp)
	if !ok {
		return
	}
	render(rsp, "players", "Players", page{Title: "Players", Tournament: t})
}

// handleResults persists one board result from the rounds page form, then
// redirects back so a refresh re-GETs the page (PRG pattern).
func (w *webServer) handleResults(rsp http.ResponseWriter, req *http.Request) {
	round, err1 := strconv.Atoi(req.PostFormValue("round"))
	board, err2 := strconv.Atoi(req.PostFormValue("board"))
	result := req.PostFormValue("result")
	if err1 != nil || err2 != nil {
		http.Error(rsp, "round and board must be integers", http.StatusBadRequest)
		return
	}
	if err := w.store.update(func(t *Tournament) error {
		return t.SetResult(round, board, result)
	}); err != nil {
		http.Error(rsp, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(rsp, req, "/rounds", http.StatusSeeOther)
}

// handlePlayersPost registers a new player from the players page form; an
// empty name is rejected by the domain, surfaced as a 400.
func (w *webServer) handlePlayersPost(rsp http.ResponseWriter, req *http.Request) {
	if err := w.store.update(func(t *Tournament) error {
		_, err := t.addPlayer(req.PostFormValue("name"))
		return err
	}); err != nil {
		http.Error(rsp, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(rsp, req, "/players", http.StatusSeeOther)
}

// render executes the named page template with auto-escaped data.
func render(rsp http.ResponseWriter, name, title string, data page) {
	rsp.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(rsp, name, data); err != nil {
		http.Error(rsp, err.Error(), http.StatusInternalServerError)
	}
}

// runServe starts the web UI and blocks until the process is interrupted.
func runServe(args []string, s *store, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(stderr)
	addr := fs.String("addr", defaultAddr, "listen address (host:port)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if env := os.Getenv("SUIZO_ADDR"); env != "" && *addr == defaultAddr {
		*addr = env
	}
	w := &webServer{store: s}
	fmt.Fprintf(stdout, "serving on http://%s\n", *addr)
	if err := http.ListenAndServe(*addr, w.handler()); err != nil {
		fmt.Fprintf(stderr, "suizo: %v\n", err)
		return 1
	}
	return 0
}
