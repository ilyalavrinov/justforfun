package main

import "fmt"
import "os"
import "bufio"
import "strings"
import "strconv"

type item string
type command string

const (
	RADAR item = "RADAR"
	TRAP  item = "TRAP"

	REQUEST command = "REQUEST"
	MOVE    command = "MOVE"
	DIG     command = "DIG"
	WAIT    command = "WAIT"
)

const (
	TYPEMYROBOT    = 0
	TYPEENEMYROBOT = 1
	TYPEMYRADAR    = 2
	TYPEMYTRAP     = 3
)

const (
	CARRYNOTHING = -1
	CARRYRADAR   = 2
	CARRYTRAP    = 3
	CARRYORE     = 4
)

func cmdRequest(i item) string {
	return fmt.Sprintf("%s %s", REQUEST, i)
}

func cmdMove(x, y int) string {
	return fmt.Sprintf("%s %d %d", MOVE, x, y)
}

func cmdDig(x, y int) string {
	return fmt.Sprintf("%s %d %d", DIG, x, y)
}

func cmdWait() {
	fmt.Print(WAIT)
}

func log(args ...interface{}) {
	fmt.Fprint(os.Stderr, args...)
}

type coord struct {
	x, y int
}

type cell struct {
	ore  int
	hole bool
}

type entity struct {
	id    string
	what  int
	carry int
}

type state struct {
	width, height int

	myScore, enemyScore int

	field    map[coord]cell
	entities map[coord]entity

	radarCooldown, trapCooldown int
	scanner                     *bufio.Scanner
}

func newState() *state {
	s := state{}
	s.field = make(map[coord]cell, s.height*s.width)
	s.entities = make(map[coord]entity, s.height*s.width)

	s.scanner = bufio.NewScanner(os.Stdin)
	s.scanner.Buffer(make([]byte, 1000000), 1000000)

	s.scanner.Scan()
	fmt.Sscan(s.scanner.Text(), &s.width, &s.height)

	return &s
}

func (s *state) update() {
	s.scanner.Scan()
	fmt.Sscan(s.scanner.Text(), &s.myScore, &s.enemyScore)

	for y := 0; y < s.height; y++ {
		s.scanner.Scan()
		inputs := strings.Split(s.scanner.Text(), " ")
		for x := 0; x < s.width; x++ {
			c := cell{}

			ore := inputs[2*x]
			if ore == "?" {
				c.ore = -1
			} else {
				c.ore, _ = strconv.Atoi(ore)
			}

			hole, _ := strconv.Atoi(inputs[2*x+1])
			if hole == 1 {
				c.hole = true
			} else {
				c.hole = false
			}

			s.field[coord{x, y}] = c
		}
	}
	var entityCount int
	s.scanner.Scan()
	fmt.Sscan(s.scanner.Text(), &entityCount, &s.radarCooldown, &s.trapCooldown)
	for i := 0; i < entityCount; i++ {
		// id: unique id of the entity
		// type: 0 for your robot, 1 for other robot, 2 for radar, 3 for trap
		// y: position of the entity
		// item: if this entity is a robot, the item it is carrying (-1 for NONE, 2 for RADAR, 3 for TRAP, 4 for ORE)
		e := entity{}
		s.scanner.Scan()
		var x, y int
		fmt.Sscan(s.scanner.Text(), &e.id, &e.what, &e.carry)

		s.entities[coord{x, y}] = e
	}
}

func (s *state) turn() {
	cmdWait()
}

func main() {
	s := newState()

	for {
		s.update()
		s.turn()
	}
}
