package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
)

type item string
type command string

var (
	H int = 15
	W int = 30
)

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
	return fmt.Sprintf("%s %s\n", REQUEST, i)
}

func cmdMove(x, y int) string {
	return fmt.Sprintf("%s %d %d\n", MOVE, x, y)
}

func cmdDig(x, y int) string {
	return fmt.Sprintf("%s %d %d\n", DIG, x, y)
}

func cmdWait() string {
	return fmt.Sprintln(WAIT)
}

func log(args ...interface{}) {
	args = append(args, "\n")
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
	xy    coord
}

type squad struct {
	radarman string
	trapman  string
	other    []string
	all      []string

	target coord
}

func merge(a, b map[string]string) map[string]string {
	res := make(map[string]string, len(a)+len(b))
	for k, v := range a {
		res[k] = v
	}
	for k, v := range b {
		res[k] = v
	}
	return res
}

func doCommand(orders map[string]string) {
	keys := make([]string, len(orders))
	for k := range orders {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Println(orders[k])
	}
}

func (sq squad) allInHQ(s *state) bool {
	for _, a := range sq.all {
		if s.entities[a].xy.x != 0 {
			return false
		}
	}
	return true
}

func (sq squad) isInSquad(id string) bool {
	if sq.radarman == id {
		return true
	}
	if sq.trapman == id {
		return true
	}
	for _, i := range sq.other {
		if i == id {
			return true
		}
	}
	return false
}

func (sq squad) carryOre(s *state) bool {
	for _, a := range sq.all {
		if s.entities[a].carry == CARRYORE {
			return true
		}
	}
	return false
}

func (sq squad) chooseRadarLocation(s *state) {
	sq.target = coord{rand.Intn(W + 1), rand.Intn(H + 1)}
}

func (sq squad) turn(s *state) map[string]string {
	orders := make(map[string]string, len(sq.all))
	if sq.carryOre(s) {
		for _, a := range sq.all {
			orders[a] = cmdMove(0, sq.target.y)
		}
	} else if sq.allInHQ(s) {
		if s.entities[sq.radarman].carry == CARRYRADAR {
			sq.chooseRadarLocation(s)
			for _, a := range sq.all {
				orders[a] = cmdMove(sq.target.x, sq.target.y)
			}
		} else if s.radarCooldown != 0 || s.trapCooldown != 0 {
			for _, a := range sq.all {
				orders[a] = cmdWait()
			}
		} else {
			orders[sq.radarman] = cmdRequest(RADAR)
			orders[sq.trapman] = cmdRequest(TRAP)
			for _, o := range sq.other {
				orders[o] = cmdWait()
			}
		}
	} else {
		// in the field and not returning ore
		if s.entities[sq.radarman].carry == CARRYRADAR {
			// moving to target
			if isNeighbour(s.entities[sq.radarman].xy, sq.target) {
				orders[sq.radarman] = cmdDig(sq.target.x, sq.target.y)
				orders[sq.trapman] = cmdWait()
				for _, o := range sq.other {
					orders[o] = cmdWait()
				}
			} else {
				for _, a := range sq.all {
					orders[a] = cmdMove(sq.target.x, sq.target.y)
				}
			}
		} else {
			for _, a := range sq.all {
				xy := s.entities[a].xy
				ns := neighbours(xy)
				for _, n := range ns {
					if s.ores[n] > 0 {
						orders[a] = cmdDig(n.x, n.y)
						break
					}
				}
				if _, found := orders[a]; !found {
					orders[a] = cmdWait()
				}
			}
		}

	}
	return orders
}

func isNeighbour(a, b coord) bool {
	if a.x == b.x {
		if a.y == b.y+1 || a.y == b.y-1 {
			return true
		}
	}
	if a.y == b.y {
		if a.x == b.x+1 || a.x == b.x-1 {
			return true
		}
	}
	return false
}

func neighbours(xy coord) []coord {
	res := make([]coord, 0, 4)
	if xy.x-1 >= 0 {
		res = append(res, coord{xy.x - 1, xy.y})
	}
	if xy.x+1 < W {
		res = append(res, coord{xy.x + 1, xy.y})
	}
	if xy.y-1 >= 0 {
		res = append(res, coord{xy.x, xy.y - 1})
	}
	if xy.y+1 < H {
		res = append(res, coord{xy.x, xy.y + 1})
	}
	return res
}

type squadsStragery struct {
	squads []squad
	free   []string
}

func newSquadStrategy() *squadsStragery {
	s := &squadsStragery{
		squads: make([]squad, 0, 5),
		free:   make([]string, 0, 5),
	}
	return s
}

func (stg *squadsStragery) turn(s *state) {
	stg.updateFree(s)
	if len(stg.free) > 0 {
		stg.rebuildSquads()
	}
	var orders map[string]string
	for _, sq := range stg.squads {
		o := sq.turn(s)
		orders = merge(orders, o)
	}
	doCommand(orders)
}

func (stg *squadsStragery) updateFree(s *state) {
	stg.free = make([]string, 0, len(s.myRobots))
	for _, e := range s.myRobots {
		seen := false
		for _, sq := range stg.squads {
			seen = seen || sq.isInSquad(e.id)
		}
		if !seen {
			stg.free = append(stg.free, e.id)
		}
	}
}

func (stg *squadsStragery) rebuildSquads() {
	if len(stg.free) >= 4 {
		stg.squads = make([]squad, 0, 2)
		if len(stg.free) == 4 {
			stg.squads = append(stg.squads, squad{
				radarman: stg.free[0],
				trapman:  stg.free[1],
			})
			stg.squads = append(stg.squads, squad{
				radarman: stg.free[2],
				trapman:  stg.free[3],
			})
		} else {
			stg.squads = append(stg.squads, squad{
				radarman: stg.free[0],
				trapman:  stg.free[1],
				other:    []string{stg.free[2]},
			})
			stg.squads = append(stg.squads, squad{
				radarman: stg.free[3],
				trapman:  stg.free[4],
			})
		}
	} else {
		stg.squads = make([]squad, 0, 1)
		sq := squad{
			radarman: stg.free[0],
		}
		if len(stg.free) >= 2 {
			sq.trapman = stg.free[1]
		}
		if len(stg.free) == 3 {
			sq.other = []string{stg.free[2]}
		}
		stg.squads = []squad{sq}
	}
	stg.free = make([]string, 0)
}

type state struct {
	width, height int

	myScore, enemyScore int

	field    map[coord]cell
	entities map[string]entity

	myRobots    []entity
	enemyRobots []entity

	ores map[coord]int

	radarCooldown, trapCooldown int

	step int
	stg  *squadsStragery

	scanner *bufio.Scanner
}

func newState() *state {
	s := state{}

	s.scanner = bufio.NewScanner(os.Stdin)
	s.scanner.Buffer(make([]byte, 1000000), 1000000)

	s.scanner.Scan()
	fmt.Sscan(s.scanner.Text(), &s.width, &s.height)

	H = s.height
	W = s.width

	s.field = make(map[coord]cell, s.height*s.width)
	s.entities = make(map[string]entity, 0)
	s.stg = newSquadStrategy()
	return &s
}

func (s *state) update() {
	s.field = make(map[coord]cell, s.width*s.height)
	s.entities = make(map[string]entity, len(s.entities))
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
		e := entity{}
		s.scanner.Scan()
		var x, y int
		fmt.Sscan(s.scanner.Text(), &e.id, &e.what, &x, &y, &e.carry)

		e.xy = coord{x, y}
		s.entities[e.id] = e
		if e.what == TYPEMYROBOT {
			s.myRobots = append(s.myRobots, e)
		} else if e.what == TYPEENEMYROBOT {
			s.enemyRobots = append(s.enemyRobots, e)
		}
	}
}

func (s *state) turn() {
	s.stg.turn(s)
}

func main() {
	s := newState()

	for {
		s.update()
		s.turn()
	}
}
