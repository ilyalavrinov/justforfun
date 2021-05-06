package main

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"strings"
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
	g.myPacs = make([]pacman, 0, visiblePacCount)
	g.opponentPacs = make([]pacman, 0, visiblePacCount)
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
	g.pellets = make([]pellet, 0, visiblePelletCount)
	for i := 0; i < visiblePelletCount; i++ {
		p := pellet{}
		scanner.Scan()
		fmt.Sscan(scanner.Text(), &p.x, &p.y, &p.value)
		g.pellets = append(g.pellets, p)
	}
	sort.Slice(g.pellets, func(i, j int) bool {
		pi := g.pellets[i]
		pj := g.pellets[j]
		if pi.value != pj.value {
			return pi.value < pj.value
		}
		if pi.x != pj.x {
			return pi.x < pj.x
		}
		return pi.y < pj.y
	})
}

type coord struct {
	x, y int
}

func dist(a, b coord) float64 {
	return math.Sqrt(math.Pow(float64(a.x-b.x), 2) + math.Pow(float64(a.y-b.y), 2))
}

func enemyNearby(my pacman, state *gamestate) *pacman {
	for _, enemy := range state.opponentPacs {
		if dist(coord{my.x, my.y}, coord{enemy.x, enemy.y}) <= 2 {
			return &enemy
		}
	}
	return nil
}

var countertypes map[string]string = map[string]string{
	"ROCK":     "PAPER",
	"PAPER":    "SCISSORS",
	"SCISSORS": "ROCK",
}

func countertype(t string) string {
	return countertypes[t]
}

func randomCoord(f *gamefield) coord {
	found := false
	var c coord
	for !found {
		x := rand.Intn(f.width)
		y := rand.Intn(f.height)
		sym := f.rows[y][x]
		if string(sym) == " " {
			c = coord{x, y}
			found = true
		}
	}
	return c
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1000000), 1000000)

	field := newGameField(scanner)
	state := newGameState(field)

	for {
		state.rescan(scanner)

		cmds := make([]string, 0, len(state.myPacs))

		targets := make(map[int]coord, len(state.myPacs))
		takenCoords := make(map[coord]bool, len(state.myPacs))
		randomMove := make([]int, 0, len(state.myPacs))

		for _, pac := range state.myPacs {
			if pac.abilityCooldown == 0 {
				enemyPac := enemyNearby(pac, state)
				if enemyPac != nil {
					cmds = append(cmds, fmt.Sprintf("SWITCH %d %s", pac.id, countertype(enemyPac.typeId)))
				} else {
					cmds = append(cmds, fmt.Sprintf("SPEED %d", pac.id))
				}
				continue
			}

			var bestChoice pellet
			for _, p := range state.pellets {
				c := coord{p.x, p.y}
				if takenCoords[c] {
					continue
				}

				if p.value > bestChoice.value {
					bestChoice = p
				} else if p.value == bestChoice.value {
					// TODO: calc dis correctly via BFS or something
					d1 := dist(coord{pac.x, pac.y}, c)
					d2 := dist(coord{pac.x, pac.y}, coord{bestChoice.x, bestChoice.y})
					if d1 < d2 {
						bestChoice = p
					}
				}
			}
			if bestChoice.value != 0 {
				c := coord{bestChoice.x, bestChoice.y}
				targets[pac.id] = c
				takenCoords[c] = true
				continue
			}

			randomMove = append(randomMove, pac.id)
		}

		for pacId, c := range targets {
			cmds = append(cmds, fmt.Sprintf("MOVE %d %d %d", pacId, c.x, c.y))
		}
		for _, pacId := range randomMove {
			c := randomCoord(field)
			cmds = append(cmds, fmt.Sprintf("MOVE %d %d %d", pacId, c.x, c.y))
		}
		fmt.Println(strings.Join(cmds, " | "))
	}
}
