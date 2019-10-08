package main

import (
	"bufio"
	"fmt"
	"math"
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

const (
	stepInventoryWait int = iota
	stepRadarDeployed     = iota
)

type squad struct {
	radarman string
	trapman  string
	other    []string
	all      []string
	dig      map[string]coord

	step int

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
		fmt.Print(orders[k])
	}
}

func calcOreDists(ore map[coord]int, xy coord) map[int][]coord {
	res := make(map[int][]coord)
	for c := range ore {
		d := dist(c, xy)
		res[d] = append(res[d], c)
	}
	return res
}

type robot2 struct {
	xy    coord
	toXY  coord
	carry int
}

func (r *robot2) turn() string {
	return cmdWait()
}

func (r *robot2) reset() {
	r.xy = coord{}
	r.toXY = coord{}
	r.carry = CARRYNOTHING
}

type strategy2 struct {
	mybots map[string]*robot2
}

func newStrategy2() *strategy2 {
	return &strategy2{
		mybots: make(map[string]*robot2, 5),
	}
}

func (str *strategy2) turn(s *state) {

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

func (sq squad) allCarryOre(s *state) bool {
	for _, a := range sq.all {
		if s.entities[a].carry != CARRYORE {
			return false
		}
	}
	return true
}

func (sq *squad) chooseRadarLocation(s *state) {
	sq.target = coord{rand.Intn(W), rand.Intn(H)}
}

func (sq *squad) turn(s *state) map[string]string {
	orders := make(map[string]string, len(sq.all))
	if sq.step == stepRadarDeployed {
		log("In stepRadarDeployed")
		atHQ := 0
		for _, a := range sq.all {
			xy := s.entities[a].xy
			if s.entities[a].carry == CARRYORE {
				log("stepRadarDeployed MOVE!")
				orders[a] = cmdMove(0, sq.target.y)
			} else if sq.dig[a] != (coord{}) {
				log("stepRadarDeployed DIG!")
				orders[a] = cmdDig(sq.dig[a].x, sq.dig[a].y)
				if isNeighbour(sq.dig[a], xy) {
					delete(sq.dig, a)
				}
			} else if xy.x == 0 {
				atHQ++
				orders[a] = cmdWait()
			} else {
				log("Searching for ore ")
				var oreCand *coord
				oreCandDist := W + H
				for orexy, cnt := range s.ores {
					if cnt <= 0 {
						continue
					}
					d := dist(xy, orexy)
					if d < oreCandDist {
						oreCand = &orexy
						oreCandDist = d
					}
				}
				if oreCand != nil {
					delete(s.ores, *oreCand)
					orders[a] = cmdDig(oreCand.x, oreCand.y)
					sq.dig[a] = *oreCand
				} else {
					orders[a] = cmdWait()
				}
			}
		}
		if atHQ == len(sq.all) {
			sq.step = stepInventoryWait
			log("setting stepInventoryWait")
		}
	} else if sq.allInHQ(s) {
		log("All in HQ")
		if s.entities[sq.radarman].carry == CARRYRADAR {
			log("Radarman has a radar")
			sq.chooseRadarLocation(s)
			log("Location chosen: ", sq.target)
			for _, a := range sq.all {
				orders[a] = cmdMove(sq.target.x, sq.target.y)
			}
		} else if s.radarCooldown != 0 || s.trapCooldown != 0 {
			log("Cooldown not reached")
			for _, a := range sq.all {
				orders[a] = cmdWait()
			}
		} else {
			log("Requesting stuff")
			orders[sq.radarman] = cmdRequest(RADAR)
			s.radarCooldown = 5
			orders[sq.trapman] = cmdRequest(TRAP)
			s.trapCooldown = 5
			for _, o := range sq.other {
				orders[o] = cmdWait()
			}
		}
	} else {
		// in the field and not returning ore
		log("In the field")
		if s.entities[sq.radarman].carry == CARRYRADAR {
			// moving to target
			log("Moving to target ", sq.target)
			if isNeighbour(s.entities[sq.radarman].xy, sq.target) || s.entities[sq.radarman].xy == sq.target {
				orders[sq.radarman] = cmdDig(sq.target.x, sq.target.y)
				sq.step = stepRadarDeployed
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
			log("UNREACHABLE??")
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

func dist(a, b coord) int {
	return int(math.Abs(float64(a.x-b.x)) + math.Abs(float64(a.y-b.y)))
}

type squadsStragery struct {
	squads []*squad
	free   []string
}

func newSquadStrategy() *squadsStragery {
	s := &squadsStragery{
		squads: make([]*squad, 0, 5),
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
	for i, sq := range stg.squads {
		log("TURN squad ", i)
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
	log("rebuilding squads len ", len(stg.free))
	if len(stg.free) >= 4 {
		stg.squads = make([]*squad, 0, 2)
		if len(stg.free) == 4 {
			stg.squads = append(stg.squads, &squad{
				radarman: stg.free[0],
				trapman:  stg.free[1],
				all:      []string{stg.free[0], stg.free[1]},
			})
			stg.squads = append(stg.squads, &squad{
				radarman: stg.free[2],
				trapman:  stg.free[3],
				all:      []string{stg.free[2], stg.free[3]},
			})
		} else {
			stg.squads = append(stg.squads, &squad{
				radarman: stg.free[0],
				trapman:  stg.free[1],
				other:    []string{stg.free[2]},
				all:      []string{stg.free[0], stg.free[1], stg.free[2]},
			})
			stg.squads = append(stg.squads, &squad{
				radarman: stg.free[3],
				trapman:  stg.free[4],
				all:      []string{stg.free[3], stg.free[4]},
			})
		}
	} else {
		stg.squads = make([]*squad, 0, 1)
		sq := squad{
			radarman: stg.free[0],
			all:      []string{stg.free[0]},
		}
		if len(stg.free) >= 2 {
			sq.trapman = stg.free[1]
			sq.all = append(sq.all, stg.free[1])
		}
		if len(stg.free) == 3 {
			sq.other = []string{stg.free[2]}
			sq.all = append(sq.all, stg.free[2])
		}
		stg.squads = []*squad{&sq}
	}
	stg.free = make([]string, 0)
	for _, s := range stg.squads {
		s.dig = make(map[string]coord, 5)
	}
	log("squads ", len(stg.squads))
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
	s.ores = make(map[coord]int, W*H)
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
				s.ores[coord{x, y}] = c.ore
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
	log("OREs", s.ores)
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
