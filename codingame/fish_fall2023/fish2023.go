package main

import (
	"fmt"
	"strings"
)

type creature struct {
	id    int
	color int
	tipo  int
}

type visibleCreature struct {
	id     int
	x, y   int
	vx, vy int
}

type drone struct {
	id        int
	x, y      int
	emergency int
	battery   int
}

type scanInfo struct {
	droneId    int
	creatureId int
}

type radarBlip struct {
	droneId    int
	creatureId int
	radar      string
}

type gameState struct {
	creatureCount int
	creatures     []creature

	myScore, oppScore int

	myScans, oppScans   []int
	myDrones, oppDrones []drone

	scans   []scanInfo
	visible []visibleCreature
	radars  []radarBlip
}

func newState() *gameState {
	var state gameState

	fmt.Scan(&state.creatureCount)
	for i := 0; i < state.creatureCount; i++ {
		var id, color, tipo int
		fmt.Scan(&id, &color, &tipo)

		state.creatures = append(state.creatures, creature{id, color, tipo})
	}
	return &state
}

func (s *gameState) reset() {
	s.myScans = make([]int, 0, len(s.creatures))
	s.oppScans = make([]int, 0, len(s.creatures))
	s.myDrones = make([]drone, 0, len(s.myDrones))
	s.oppDrones = make([]drone, 0, len(s.oppDrones))
	s.scans = make([]scanInfo, 0, len(s.scans))
	s.visible = make([]visibleCreature, 0, len(s.creatures))
	s.radars = make([]radarBlip, 0, len(s.radars))
}

func (s *gameState) readTurn() {
	s.reset()

	fmt.Scan(&s.myScore, &s.oppScore)

	var scanCount int
	// my
	fmt.Scan(&scanCount)
	for i := 0; i < scanCount; i++ {
		var id int
		fmt.Scan(&id)
		s.myScans = append(s.myScans, id)
	}
	// opp
	fmt.Scan(&scanCount)
	for i := 0; i < scanCount; i++ {
		var id int
		fmt.Scan(&id)
		s.oppScans = append(s.oppScans, id)
	}

	var droneCount int
	// my
	fmt.Scan(&droneCount)
	for i := 0; i < droneCount; i++ {
		var d drone
		fmt.Scan(&d.id, &d.x, &d.y, &d.emergency, &d.battery)
		s.myDrones = append(s.myDrones, d)
	}
	// opp
	fmt.Scan(&droneCount)
	for i := 0; i < droneCount; i++ {
		var d drone
		fmt.Scan(&d.id, &d.x, &d.y, &d.emergency, &d.battery)
		s.oppDrones = append(s.oppDrones, d)
	}

	fmt.Scan(&scanCount)
	for i := 0; i < scanCount; i++ {
		var scan scanInfo
		fmt.Scan(&scan.droneId, &scan.creatureId)
		s.scans = append(s.scans, scan)
	}

	var visibleCreatureCount int
	fmt.Scan(&visibleCreatureCount)
	for i := 0; i < visibleCreatureCount; i++ {
		var vc visibleCreature
		fmt.Scan(&vc.id, &vc.x, &vc.y, &vc.vx, &vc.vy)
		s.visible = append(s.visible, vc)
	}

	var radarBlipCount int
	fmt.Scan(&radarBlipCount)
	for i := 0; i < radarBlipCount; i++ {
		var blip radarBlip
		fmt.Scan(&blip.droneId, &blip.creatureId, &blip.radar)

		s.radars = append(s.radars, blip)
	}
}

func (s *gameState) cmd() string {
	var cmds []string
	for i := 0; i < len(s.myDrones); i++ {
		// fmt.Fprintln(os.Stderr, "Debug messages...")
		cmds = append(cmds, "WAIT 0") // MOVE <x> <y> <light (1|0)> | WAIT <light (1|0)>
	}
	return strings.Join(cmds, " | ")
}

func main() {
	state := newState()
	for {
		state.readTurn()
		cmd := state.cmd()
		fmt.Println(cmd)
	}
}
