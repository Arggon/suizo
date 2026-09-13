package main

import "errors"

var errEmptyName = errors.New("player name must not be empty")

var errTooFewPlayers = errors.New("need at least 2 players to pair")

var errNoByeCandidate = errors.New("bye required but every player already had one")

var errNoRematchFreePairing = errors.New("no rematch-free pairing exists")

var errRoundPending = errors.New("last round still has pending matches")

var errIllegalResult = errors.New(`result must be one of "1-0", "0-1", "0.5-0.5"`)

var errNoRounds = errors.New("no rounds started yet")

var errByeResult = errors.New("cannot report a result for a bye")
