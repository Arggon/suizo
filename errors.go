package main

import "errors"

var errEmptyName = errors.New("player name must not be empty")

var errTooFewPlayers = errors.New("need at least 2 players to pair")

var errNoByeCandidate = errors.New("bye required but every player already had one")

var errNoRematchFreePairing = errors.New("no rematch-free pairing exists")
