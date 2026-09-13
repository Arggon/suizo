# suizo

Swiss-system tournament manager for amateur events — chess, futbolito, table
tennis, anything pairable. One static binary, one JSON file, zero infrastructure.

Named after the Swiss system (sistema suizo), the tournament format where nobody
is eliminated: everyone plays every round, pairings favor equal records, and the
best score after N rounds wins. It is how almost every amateur club tournament
in the world is actually run.

## Stack

| Piece    | Choice                                   | Why (short)                                     |
| -------- | ---------------------------------------- | ----------------------------------------------- |
| Language | Go 1.27.x                                | one static binary, fast tests, batteries incl.  |
| CLI      | `flag` + `os.Args`, stdlib only          | the surface is 3 subcommands; a framework is a liability |
| Storage  | single JSON file (`suizo.json`)          | no server, human-inspectable, trivially backed up |
| Web (later) | `net/http` stdlib                     | issue #4                                        |
| Lint     | golangci-lint v2                         |                                                 |

See [docs/DECISIONS.md](docs/DECISIONS.md) for the full reasoning and
[docs/FORMAT.md](docs/FORMAT.md) for the on-disk state format.

## Status

MVP skeleton: player registration via `suizo players add|list`. Pairing,
standings, rounds and the web UI are tracked as issues.

## Usage

```console
$ suizo players add "Ana García"
p1
$ suizo players add "Beto Ruiz"
p2
$ suizo players list
p1	Ana García
p2	Beto Ruiz
```

State lives in `./suizo.json` (override with `$SUIZO_FILE`).

## Development

```console
$ go test ./...
$ go vet ./...
$ golangci-lint run
$ go build -o suizo .
```

## License

MIT.
