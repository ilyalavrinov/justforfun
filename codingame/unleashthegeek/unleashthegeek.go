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

func cmdRequest(i item) {
	fmt.Printf("%s %s\n", REQUEST, i)
}

func cmdMove(x, y int) {
	fmt.Printf("%s %d %d\n", MOVE, x, y)
}

func cmdDig(xy coord) {
	fmt.Printf("%s %d %d\n", DIG, xy.x, xy.y)
}

func cmdWait() {
	fmt.Println(WAIT)
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

type strategy interface {
	turn(*state)
}

type scanningTask struct {
	y           int
	toHQ        bool
	rowFinished bool
}

type scanningStrategy struct {
	visitedY      map[int]struct{}
	assignedTasks map[string]*scanningTask
}

func newScanningStrategy() *scanningStrategy {
	stg := &scanningStrategy{
		visitedY:      make(map[int]struct{}),
		assignedTasks: make(map[string]*scanningTask),
	}
	return stg
}

func (stg *scanningStrategy) turn(s *state) {
	log("len entities ", len(s.entities))
	for _, e := range s.entities {
		log("ET ", e.what, " id ", e.id)
		if e.what != TYPEMYROBOT {
			continue
		}
		xy := e.xy
		task, found := stg.assignedTasks[e.id]
		if !found {
			log("need new task")
			task = stg.newTask(s, e, xy, nil)
		}
		if task.toHQ {
			log("toHQ!")
			if xy.x == 0 {
				task.toHQ = false
				if task.rowFinished {
					log("row finished; need new task")
					task = stg.newTask(s, e, xy, task)
				}
			} else {
				cmdMove(0, xy.y)
				continue
			}
		}
		if xy.y != task.y {
			log("change y")
			cmdMove(xy.x, task.y)
			continue
		}
		if e.carry == CARRYORE {
			log("carry ore, fallback")
			cmdMove(0, xy.y)
			continue
		}
		c := s.field[xy]
		if c.ore > 0 {
			log("digging ore")
			cmdDig(xy)
			continue
		}
		if !c.hole && xy.x != 0 {
			log("digging test hole")
			cmdDig(xy)
			continue
		}
		if xy.x == s.width-1 {
			log("row finished, back to HQ")
			task.toHQ = true
			task.rowFinished = true
			cmdMove(0, xy.y)
			continue
		}
		log("next step")
		cmdMove(xy.x+1, xy.y)
	}
}

func (stg *scanningStrategy) newTask(s *state, myRobot entity, xy coord, prevTask *scanningTask) *scanningTask {
	yToCheck := []int{xy.y}
	inc := 1
	for xy.y-inc >= 0 || xy.y+inc < s.height {
		if xy.y-inc >= 0 {
			yToCheck = append(yToCheck, xy.y-inc)
		}
		if xy.y+inc < s.height {
			yToCheck = append(yToCheck, xy.y+inc)
		}
		inc++
	}
	var task *scanningTask
	for _, y := range yToCheck {
		log("check Y: ", y)
		if _, found := stg.visitedY[y]; !found {
			task = &scanningTask{
				y:    y,
				toHQ: true,
			}
			stg.visitedY[y] = struct{}{}
			stg.assignedTasks[myRobot.id] = task
			break
		}
	}
	log("new task: ", task)
	return task
}

var _ strategy = (*scanningStrategy)(nil)

type state struct {
	width, height int

	myScore, enemyScore int

	field    map[coord]cell
	entities []entity

	radarCooldown, trapCooldown int

	step int
	stg  strategy

	scanner *bufio.Scanner
}

func newState() *state {
	s := state{}
	s.field = make(map[coord]cell, s.height*s.width)
	s.entities = make([]entity, 0)
	s.stg = newScanningStrategy()

	s.scanner = bufio.NewScanner(os.Stdin)
	s.scanner.Buffer(make([]byte, 1000000), 1000000)

	s.scanner.Scan()
	fmt.Sscan(s.scanner.Text(), &s.width, &s.height)

	return &s
}

func (s *state) update() {
	s.field = make(map[coord]cell, s.width*s.height)
	s.entities = make([]entity, 0, len(s.entities))
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
		s.entities = append(s.entities, e)
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
