package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

func main() {
	conn, _ := net.Dial("tcp", "127.0.0.1:3000")
	defer conn.Close()

	r := bufio.NewReader(conn)
	for {
		line, _, err := r.ReadLine()
		if err != nil {
			break
		}
		sline := string(line)
		fmt.Println(sline)

		var player string
		if strings.HasPrefix(sline, "start ") {
			player = strings.TrimPrefix(sline, "start ")
			fmt.Println("Player:", player)
			fmt.Fprintln(conn, "state")
			continue
		}
		if strings.HasPrefix(sline, "state ") {
			s := strings.TrimPrefix(sline, "state ")
			printState(s)
			continue
		}
		if sline == "turn" {
			fmt.Print("Place: ")

			var pos int
			fmt.Scanf("%d", &pos)
			fmt.Fprintln(conn, "place", pos)
			continue
		}
		if sline == "end" {
			fmt.Println("Game End")
			return
		}
	}
}

func printState(s string) {
	ss := []rune(s)
	for i, c := range ss {
		fmt.Printf("%c", c)
		if i%3 == 2 {
			fmt.Println()
		}
	}
}
