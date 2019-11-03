package main

import (
	"fmt"
	"os"
	"strconv"
)

const (
	symFloor     = '.'
	boxEmpty     = 0
	boxItemRange = 1
	boxItemBomb  = 2

	bombTimer = 8
	bombSpan  = 3
	bombMax   = 1

	entityPlayer = 0
	entityBomb   = 1
	entityItem   = 2

	itemRange = 1
	itemBomb  = 2
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
			//fmt.Fprintf(os.Stderr, "HEATMAP check X %d Y %d OBJ %v\n", x, y, o)
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

func step(g grid, myId int) {
	for y := 0; y < height; y++ {
		var row string
		fmt.Scan(&row)
		for x, c := range row {
			if c == symFloor {
				g[x][y] = objFloor
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

	fmt.Fprintf(os.Stderr, "HEATMAP\n %v\n DIST\n %v", h, d)

	var bestXY coord
	var bestScore int
	for x, col := range h {
		for y, score := range col {
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
