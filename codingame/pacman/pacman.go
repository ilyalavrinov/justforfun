package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
)

type pacman struct {
	id              int
	x, y            int
	owner           int
	typeId          string
	abilityCooldown int
	speedTurnsLeft  int
}

type pellet struct {
	x, y  int
	value int
}

type gamefield struct {
	width, height int
	rows          []string
}

func newGameField(scanner *bufio.Scanner) *gamefield {
	g := gamefield{}

	scanner.Scan()
	fmt.Sscan(scanner.Text(), &g.width, &g.height)

	g.rows = make([]string, 0, g.height)
	for i := 0; i < g.height; i++ {
		scanner.Scan()
		g.rows = append(g.rows, scanner.Text()) // one line of the grid: space " " is floor, pound "#" is wall
	}

	return &g
}

type gamestate struct {
	field *gamefield

	myScore, opponentScore int
	myPacs, opponentPacs   []pacman
	pellets                []pellet
}

func newGameState(field *gamefield) *gamestate {
	gs := gamestate{
		field: field,
	}
	return &gs
}

func (g *gamestate) rescan(scanner *bufio.Scanner) {
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &g.myScore, &g.opponentScore)

	var visiblePacCount int
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &visiblePacCount)
	for i := 0; i < visiblePacCount; i++ {
		p := pacman{}
		scanner.Scan()
		fmt.Sscan(scanner.Text(), &p.id, &p.owner, &p.x, &p.y, &p.typeId, &p.speedTurnsLeft, &p.abilityCooldown)

		if p.owner == 1 {
			g.myPacs = append(g.myPacs, p)
		} else {
			g.opponentPacs = append(g.opponentPacs, p)
		}
	}

	var visiblePelletCount int
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &visiblePelletCount)

	for i := 0; i < visiblePelletCount; i++ {
		p := pellet{}
		scanner.Scan()
		fmt.Sscan(scanner.Text(), &p.x, &p.y, &p.value)
		g.pellets = append(g.pellets, p)
	}
	sort.Slice(g.pellets, func(i, j int) bool {
		pi := g.pellets[i]
		pj := g.pellets[j]
		if pi.value >= pj.value {
			return false
		}
		if pi.x >= pj.x {
			return false
		}
		if pi.y >= pj.y {
			return false
		}
		return true
	})
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1000000), 1000000)

	field := newGameField(scanner)
	state := newGameState(field)

	for {
		state.rescan(scanner)

		for i, pac := range state.myPacs {
			if i < len(state.pellets) {
				fmt.Println("MOVE", pac.id, state.pellets[i].x, state.pellets[i].y)
			} else {
				fmt.Fprintln(os.Stderr, "No more pellets!")
				fmt.Println("MOVE", pac.id, pac.x, pac.y)
			}
		}
	}
}
