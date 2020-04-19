package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
)

var (
	muLastGameID sync.Mutex
	lastGameID   = 0
)

func generateGameID() int {
	muLastGameID.Lock()
	defer muLastGameID.Unlock()

	lastGameID++
	return lastGameID
}

func main() {
	log.Println("OXGame listening on :3000")
	l, _ := net.Listen("tcp", ":3000")

	var (
		muGames sync.Mutex
		games   = make(map[int]*OXGame)
	)
	var (
		muWaitingGameID sync.Mutex
		waitingGameID   int
	)

	for {
		conn, _ := l.Accept()

		go func() {
			defer conn.Close()

			var game *OXGame
			var player Player
			var gameID int

			muWaitingGameID.Lock()
			if waitingGameID == 0 {
				game = NewOXGame(X)
				gameID = generateGameID()
				muGames.Lock()
				games[gameID] = game
				muGames.Unlock()
				player = X
				waitingGameID = gameID
			} else {
				gameID = waitingGameID
				waitingGameID = 0
				game = games[gameID]
				player = O
			}
			muWaitingGameID.Unlock()

			game.SetPlayer(player, conn)

			if game.Ready() {
				for pconn, player := range game.conns {
					fmt.Fprintln(pconn, "start", player)
				}
				fmt.Fprintln(game.Conn(game.turn), "turn")
			}

			r := bufio.NewReader(conn)

			for {
				line, _, err := r.ReadLine()
				sline := string(line)
				if err != nil {
					fmt.Fprintln(game.Conn(X), "end")
					fmt.Fprintln(game.Conn(O), "end")
					break
				}

				fmt.Println(gameID, ":", player, ":", sline)
				if sline == "state" {
					fmt.Fprintln(conn, "state", game.State())
				}

				// place 1
				if strings.HasPrefix(sline, "place ") {
					sPos := strings.TrimPrefix(sline, "place ")
					pos, err := strconv.Atoi(sPos)
					if err != nil {
						fmt.Fprintln(conn, "err")
						fmt.Fprintln(conn, "turn")
						continue
					}

					err = game.Place(player, pos)
					if err != nil {
						fmt.Fprintln(conn, "err", err)
						fmt.Fprintln(conn, "turn")
						continue
					}

					{
						state := game.State()
						fmt.Fprintln(game.Conn(X), "state", state)
						fmt.Fprintln(game.Conn(O), "state", state)
					}

					if winner := game.Winner(); winner != None {
						fmt.Fprintln(game.Conn(X), "winner", winner)
						fmt.Fprintln(game.Conn(O), "winner", winner)

						game.Conn(X).Close()
						game.Conn(O).Close()
						muGames.Lock()
						delete(games, gameID)
						muGames.Unlock()
						break
					}

					{
						pConn := game.Conn(player.Swap())
						fmt.Fprintln(pConn, "turn")
					}
				}
			}
		}()
	}
}

type Player int

const (
	None Player = iota
	X
	O
)

func (p Player) String() string {
	switch p {
	default:
		return "_"
	case X:
		return "X"
	case O:
		return "O"
	}
}

func (p Player) Swap() Player {
	switch p {
	default:
		panic("invalid")
	case X:
		return O
	case O:
		return X
	}
}

type OXGame struct {
	board [9]Player
	turn  Player
	conns map[net.Conn]Player
}

func NewOXGame(turn Player) *OXGame {
	return &OXGame{
		turn:  turn,
		conns: make(map[net.Conn]Player),
	}
}

func (game *OXGame) Ready() bool {
	return len(game.conns) == 2
}

func (game *OXGame) SetPlayer(p Player, conn net.Conn) {
	game.conns[conn] = p
}

func (game *OXGame) Conn(p Player) net.Conn {
	for conn, player := range game.conns {
		if player == p {
			return conn
		}
	}
	return nil
}

// Print prints current board
// O O O
// O O O
// O O O
func (game *OXGame) String() string {
	var r string

	for i := range game.board {
		r += fmt.Sprint(game.board[i])
		if i%3 == 2 {
			r += "\n"
		}
	}
	return r
}

func (game *OXGame) Place(p Player, pos int) error {
	if pos < 0 || pos >= 9 {
		return fmt.Errorf("invalid")
	}
	if game.board[pos] != None {
		return fmt.Errorf("invalid")
	}
	if game.turn != p {
		return fmt.Errorf("invalid")
	}
	if game.Winner() != None {
		return fmt.Errorf("invalid")
	}

	game.board[pos] = p
	game.turn = p.Swap()
	return nil
}

func (game *OXGame) Winner() Player {
	b := game.board
	f := func(p1, p2, p3 int) Player {
		if b[p1] == None {
			return None
		}
		if b[p1] == b[p2] && b[p2] == b[p3] {
			return b[p1]
		}
		return None
	}

	winStates := [][]int{
		{0, 1, 2},
		{3, 4, 5},
		{6, 7, 8},
		{0, 3, 6},
		{1, 4, 7},
		{2, 5, 8},
		{0, 4, 8},
		{2, 4, 6},
	}

	for _, s := range winStates {
		if r := f(s[0], s[1], s[2]); r != None {
			return r
		}
	}

	return None
}

func (game *OXGame) State() string {
	var r string
	for i := range game.board {
		r += game.board[i].String()
	}
	return r
}
