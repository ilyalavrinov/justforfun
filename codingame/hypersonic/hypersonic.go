package main

import (
	"fmt"
	"os"
	"strconv"
)

const (
	symFloor     = '.'
	symWall      = 'X'
	boxEmpty     = 0
	boxItemRange = 1
	boxItemBomb  = 2

	bombTimer = 8

	entityPlayer = 0
	entityBomb   = 1
	entityItem   = 2

	itemRange = 1
	itemBomb  = 2
)

var (
	bombSpan  int = 3
	bombAvail int = 1
)

var (
	width  int
	height int
)

type coord struct {
	x, y int
}

func cmdMove(xy coord) string {
	return fmt.Sprintf("MOVE %d %d", xy.x, xy.y)
}

func cmdBomb(xy coord) string {
	return fmt.Sprintf("BOMB %d %d", xy.x, xy.y)
}

type objectType int

const (
	objMe           objectType = iota
	objEnemy        objectType = iota
	objBoxEmpty     objectType = iota
	objBoxItemRange objectType = iota
	objBoxItemBomb  objectType = iota
	objMyBomb       objectType = iota
	objEnemyBomb    objectType = iota
	objFloor        objectType = iota
	objWall         objectType = iota
	objItemRange    objectType = iota
	objItemBomb     objectType = iota
)

func (o objectType) isBox() bool {
	return o == objBoxEmpty || o == objBoxItemBomb || o == objBoxItemRange
}

type grid [][]objectType

func newGrid(w, h int) grid {
	g := make([][]objectType, w)
	for x := 0; x < w; x++ {
		g[x] = make([]objectType, h)
	}
	return g
}

func newField(defaultValue int) [][]int {
	f := make([][]int, width)
	for x := 0; x < width; x++ {
		f[x] = make([]int, height)
		for y := 0; y < height; y++ {
			f[x][y] = defaultValue
		}
	}
	return f
}

func newHeatmap(g grid, bombRange int) [][]int {
	h := newField(-1)
	calcF := func(xy coord, xMult, yMult int) int {
		for r := 1; r < bombRange; r++ {
			x := xy.x + r*xMult
			y := xy.y + r*yMult
			if x < 0 || x >= width || y < 0 || y >= height {
				return 0
			}
			o := g[x][y]
			//fmt.Fprintf(os.Stderr, "HEATMAP at X %d Y %d check X %d Y %d OBJ %+v\n", xy.x, xy.y, x, y, o)
			if o.isBox() {
				return 1
			}
			if o == objWall || o == objItemRange || o == objItemBomb {
				return 0
			}
		}
		return 0
	}
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			o := g[x][y]
			if o.isBox() || o == objWall {
				continue
			}
			var heatVal int
			heatVal += calcF(coord{x, y}, 1, 0)
			heatVal += calcF(coord{x, y}, -1, 0)
			heatVal += calcF(coord{x, y}, 0, 1)
			heatVal += calcF(coord{x, y}, 0, -1)
			h[x][y] = heatVal
			//fmt.Fprintf(os.Stderr, "heatmap for X %d Y %d VAL %d\n", x, y, heatVal)
		}
	}
	return h
}

func newDistances(g grid, pos coord) [][]int {
	f := newField(-1)
	frontier := []coord{pos}
	f[pos.x][pos.y] = 0
	checkAndAdd := func(x, y, dist int) {
		if x >= 0 && x < width && y >= 0 && y < height && !g[x][y].isBox() && g[x][y] != objWall && f[x][y] == -1 {
			frontier = append(frontier, coord{x, y})
			f[x][y] = dist
			//fmt.Fprintf(os.Stderr, "new frontier X %d Y %d DIST %d\n", x, y, dist)
		}
	}
	for i := 0; i < len(frontier); i++ {
		xy := frontier[i]
		checkAndAdd(xy.x-1, xy.y, f[xy.x][xy.y]+1)
		checkAndAdd(xy.x+1, xy.y, f[xy.x][xy.y]+1)
		checkAndAdd(xy.x, xy.y-1, f[xy.x][xy.y]+1)
		checkAndAdd(xy.x, xy.y+1, f[xy.x][xy.y]+1)
	}

	return f
}

type bomb struct {
	xy        coord
	owner     int
	countdown int
	span      int
}

type bombs struct {
	pos         map[coord]bomb
	byCountdown []bomb
}

type bombNetworks struct {
	networks   []map[coord]struct{}
	membership map[coord]int
}

func newBombmap(b bombs, g grid) [][]int {
	f := newField(-1)

	findBomb := func(b bomb, xMult, yMult int) *coord {
		for r := 1; r < b.span; r++ {
			x := b.xy.x + r*xMult
			y := b.xy.y + r*yMult
			if x < 0 || x >= width || y < 0 || y >= height {
				return nil
			}
			o := g[x][y]
			if o.isBox() || o == objWall || o == objItemRange || o == objItemBomb {
				return nil
			}
			if o == objMyBomb || o == objEnemyBomb {
				// TODO: can have some more bomb further if range is big enough
				return &coord{x, y}
			}
		}
		return nil
	}

	adjacents := make(map[coord][]coord, 0)
	for _, bmb := range b.pos {
		if xy := findBomb(bmb, 1, 0); xy != nil {
			adjacents[bmb.xy] = append(adjacents[bmb.xy], *xy)
		}
		if xy := findBomb(bmb, -1, 0); xy != nil {
			adjacents[bmb.xy] = append(adjacents[bmb.xy], *xy)
		}
		if xy := findBomb(bmb, 0, 1); xy != nil {
			adjacents[bmb.xy] = append(adjacents[bmb.xy], *xy)
		}
		if xy := findBomb(bmb, 0, -1); xy != nil {
			adjacents[bmb.xy] = append(adjacents[bmb.xy], *xy)
		}
	}

	chains := make([]map[coord]struct{}, 0)
	curChain := make(map[coord]struct{}, 0)
	frontier := make([]coord, 0)

	for len(adjacents) > 0 {
		if len(frontier) > 0 {
			newFrontier := make([]coord, 0)
			for _, xy := range frontier {
				adj := adjacents[xy]
				delete(adjacents, xy)
				curChain[xy] = struct{}{}
				for _, xy2 := range adj {
					if _, found := adjacents[xy2]; found {
						newFrontier = append(newFrontier, xy2)
					}
				}
			}
			frontier = newFrontier
		} else {
			if len(curChain) > 0 {
				chains = append(chains, curChain)
				fmt.Fprintf(os.Stderr, "Added new chain of %d elements to list of chains (now len %d)\n", len(curChain), len(chains))
				curChain = make(map[coord]struct{}, 0)
			}
			var xy coord
			var adj []coord
			for c, ch := range adjacents {
				xy = c
				adj = ch
				break
			}
			delete(adjacents, xy)
			curChain[xy] = struct{}{}
			for _, xy2 := range adj {
				fmt.Fprintf(os.Stderr, "New adjacent for bomb at %v: %v\n", xy, xy2)
				frontier = append(frontier, xy2)
			}
		}

		if len(curChain) > 0 {
			chains = append(chains, curChain)
			fmt.Fprintf(os.Stderr, "Added new chain of %d elements to list of chains (now len %d)\n", len(curChain), len(chains))
		}

	}
	return f
}

func step(g grid, myId int) {
	for y := 0; y < height; y++ {
		var row string
		fmt.Scan(&row)
		for x, c := range row {
			if c == symFloor {
				g[x][y] = objFloor
			} else if c == symWall {
				g[x][y] = objWall
			} else {
				boxN, _ := strconv.Atoi(string(c))
				switch boxN {
				case boxEmpty:
					g[x][y] = objBoxEmpty
				case boxItemRange:
					g[x][y] = objBoxItemRange
				case boxItemBomb:
					g[x][y] = objBoxItemBomb
				}
			}
		}
	}
	var entities int
	fmt.Scan(&entities)

	var myPos coord

	for i := 0; i < entities; i++ {
		var entityType, owner, x, y, param1, param2 int
		fmt.Scan(&entityType, &owner, &x, &y, &param1, &param2)

		switch entityType {
		case entityPlayer:
			if owner == myId {
				g[x][y] = objMe
				myPos = coord{x, y}
				bombAvail = param1
				bombSpan = param2
			} else {
				g[x][y] = objEnemy
			}
		case entityBomb:
			if owner == myId {
				g[x][y] = objMyBomb
			} else {
				g[x][y] = objEnemyBomb
			}
		case entityItem:
			switch param1 {
			case itemRange:
				g[x][y] = objItemRange
			case itemBomb:
				g[x][y] = objItemBomb
			}
		}
	}

	h := newHeatmap(g, bombSpan)
	d := newDistances(g, myPos)
	b := newBombmap(bmb, g)

	fmt.Fprintf(os.Stderr, "HEATMAP\n %v\n DIST\n %v", h, d)

	var bestXY coord
	var bestScore int
	for x, col := range h {
		for y, score := range col {
			if d[x][y] == -1 {
				continue
			}
			if score > bestScore {
				bestXY = coord{x, y}
				bestScore = score
				fmt.Fprintf(os.Stderr, "new best candidate at X %d Y %d SCORE %d\n", x, y, score)
			} else if score == bestScore {
				if d[x][y] < d[bestXY.x][bestXY.y] {
					bestXY = coord{x, y}
					fmt.Fprintf(os.Stderr, "new best candidate at X %d Y %d SAME SCORE %d\n", x, y, score)
				}
			}
		}
	}

	if myPos == bestXY {
		fmt.Println(cmdBomb(myPos))
	} else {
		fmt.Println(cmdMove(bestXY))
	}

}

func main() {
	var myId int
	fmt.Scan(&width, &height, &myId)

	g := newGrid(width, height)

	for {
		step(g, myId)
	}
}
