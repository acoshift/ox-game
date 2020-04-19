package main

import "testing"

func TestOXGame_Winner(t *testing.T) {
	game := NewOXGame(X)
	game.Place(X, 0)
	game.Place(O, 1)

	if r := game.Winner(); r != None {
		t.Errorf("invalid winner; expected None; got %s", r)
	}

	game.Place(X, 3)
	game.Place(O, 2)
	game.Place(X, 6)
	if r := game.Winner(); r != X {
		t.Errorf("invalid winner; expected X; got %s", r)
	}
}
