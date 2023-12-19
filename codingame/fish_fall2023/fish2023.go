package main

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
)

func log(smth ...any) {
	fmt.Fprintln(os.Stderr, smth...)
}

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
	myScansCreatureMap  map[int]struct{}
	myDrones, oppDrones []drone

	scans      []scanInfo
	visible    []visibleCreature
	visibleMap map[int]visibleCreature
	radars     []radarBlip
}

func newState() *gameState {
	var state gameState

	fmt.Scan(&state.creatureCount)
	for i := 0; i < state.creatureCount; i++ {
		var id, color, tipo int
		fmt.Scan(&id, &color, &tipo)

		state.creatures = append(state.creatures, creature{id, color, tipo})
	}
	log("new state")
	return &state
}

func (s *gameState) reset() {
	s.myScans = make([]int, 0, len(s.creatures))
	s.myScansCreatureMap = make(map[int]struct{}, len(s.creatures))
	s.oppScans = make([]int, 0, len(s.creatures))
	s.myDrones = make([]drone, 0, len(s.myDrones))
	s.oppDrones = make([]drone, 0, len(s.oppDrones))
	s.scans = make([]scanInfo, 0, len(s.scans))
	s.visible = make([]visibleCreature, 0, len(s.creatures))
	s.visibleMap = make(map[int]visibleCreature, len(s.creatures))
	s.radars = make([]radarBlip, 0, len(s.radars))
	log("reset done")
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
		s.myScansCreatureMap[id] = struct{}{}
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
		s.visibleMap[vc.id] = vc
	}

	var radarBlipCount int
	fmt.Scan(&radarBlipCount)
	for i := 0; i < radarBlipCount; i++ {
		var blip radarBlip
		fmt.Scan(&blip.droneId, &blip.creatureId, &blip.radar)

		s.radars = append(s.radars, blip)
	}
	log("read turn done")
}

const (
	opMove = iota
	opWait
)

type command struct {
	droneId   int
	operation int
	toX, toY  int
	light     bool
}

func (cmd *command) String() string {
	var result string
	light := 0
	if cmd.light {
		light = 1
	}
	if cmd.operation == opMove {
		log("cmd move", cmd.droneId)
		result = fmt.Sprintf("MOVE %d %d %d", cmd.toX, cmd.toY, light)
	} else {
		log("cmd wait", cmd.droneId)
		result = fmt.Sprintf("WAIT %d", light)
	}
	return result
}

func (s *gameState) cmd() string {
	rawCmds := make([]command, 0, len(s.myDrones))
	alldists := sortVisibleByDroneDistance(s.myDrones, s.visible)
	for droneId, dists := range alldists {
		bestCreatureId := -1
		for _, dist := range dists {
			if _, found := s.myScansCreatureMap[dist.creatureId]; !found {
				bestCreatureId = dist.creatureId
				break
			}
		}

		if bestCreatureId == -1 {
			rawCmds = append(rawCmds, command{
				droneId:   droneId,
				operation: opWait,
			})
		} else {
			rawCmds = append(rawCmds, command{
				droneId:   droneId,
				operation: opMove,
				toX:       s.visibleMap[bestCreatureId].x,
				toY:       s.visibleMap[bestCreatureId].y,
			})
		}
	}

	sort.Slice(rawCmds, func(i, j int) bool {
		return rawCmds[i].droneId < rawCmds[j].droneId
	})
	var cmds []string
	for _, c := range rawCmds {
		cmds = append(cmds, c.String())
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

type creatureDistance struct {
	creatureId int
	dist       float64
}

func sortVisibleByDroneDistance(drones []drone, creatures []visibleCreature) map[int][]creatureDistance {
	result := make(map[int][]creatureDistance)
	for _, d := range drones {
		distances := make([]creatureDistance, 0)
		for _, c := range creatures {
			dist := math.Sqrt((float64(d.x-c.x) + (float64(d.y - c.y))))
			distances = append(distances, creatureDistance{c.id, dist})
		}
		sort.Slice(distances, func(i, j int) bool {
			return (distances[i].dist < distances[j].dist)
		})
		result[d.id] = distances
	}
	log("sort visible distance done", len(result))
	return result
}
