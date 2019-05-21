package main

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"
)

/**
 * Help the Christmas elves fetch presents in a magical labyrinth!
 **/

const (
	turnPush = 0
	turnMove = 1

	maxHeight = 7
	maxWidth  = 7

	playerMe    = 0
	playerEnemy = 1
)

func log(format string, args ...interface{}) {
	format += "\n"
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, format)
	} else {
		fmt.Fprintf(os.Stderr, format, args...)
	}
}

func cmdTurnPush(arg int, dir string) string {

	return fmt.Sprintf("PUSH %d %s", arg, dir)
}

func cmdTurnMove() string {
	return "PASS"
}

type tile struct {
	up, right, down, left bool
}

func newTile(directions string) tile {
	t := tile{}
	if directions[0] == '1' {
		t.up = true
	}
	if directions[1] == '1' {
		t.right = true
	}
	if directions[2] == '1' {
		t.down = true
	}
	if directions[3] == '1' {
		t.left = true
	}
	return t
}

type coord struct {
	col, row int
}

type gameState struct {
	tiles    map[coord]tile
	turnType int

	me, enemy struct {
		coord
		totalQuests int
		extraTile   tile
	}

	myItems    map[string]coord
	enemyItems map[string]coord

	myQuests    []string
	enemyQuests []string

	edges map[coord][]coord
}

func newGameState() gameState {
	game := gameState{tiles: make(map[coord]tile, maxHeight*maxWidth)}

	game.init()
	game.findEdges()

	return game
}

func (g *gameState) clone() *gameState {
	g2 := *g
	g2.tiles = make(map[coord]tile, len(g.tiles))
	for k, v := range g.tiles {
		g2.tiles[k] = v
	}

	g2.myItems = make(map[string]coord, len(g.myItems))
	for k, v := range g.myItems {
		g2.myItems[k] = v
	}

	g2.enemyItems = make(map[string]coord, len(g.enemyItems))
	for k, v := range g.enemyItems {
		g2.enemyItems[k] = v
	}
	copy(g2.myQuests, g.myQuests)
	copy(g2.enemyQuests, g.enemyQuests)

	return &g2
}

func (g *gameState) init() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &g.turnType)

	for i := 0; i < maxHeight; i++ {
		scanner.Scan()
		inputs := strings.Split(scanner.Text(), " ")
		for j := 0; j < maxWidth; j++ {
			directions := inputs[j]
			t := newTile(directions)
			g.tiles[coord{col: j, row: i}] = t
		}
	}

	// my info
	scanner.Scan()
	var extraTileDirs string
	fmt.Sscan(scanner.Text(), &g.me.totalQuests, &g.me.col, &g.me.row, &extraTileDirs)
	g.me.extraTile = newTile(extraTileDirs)

	// enemy info
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &g.enemy.totalQuests, &g.enemy.col, &g.enemy.row, &extraTileDirs)
	g.enemy.extraTile = newTile(extraTileDirs)

	// numItems: the total number of items available on board and on player tiles
	var numItems int
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &numItems)

	g.myItems = make(map[string]coord, numItems)
	g.enemyItems = make(map[string]coord, numItems)
	for i := 0; i < numItems; i++ {
		var itemName string
		var itemX, itemY, itemPlayerID int
		scanner.Scan()
		fmt.Sscan(scanner.Text(), &itemName, &itemX, &itemY, &itemPlayerID)
		if itemPlayerID == playerMe {
			g.myItems[itemName] = coord{itemX, itemY}
		} else if itemPlayerID == playerEnemy {
			g.enemyItems[itemName] = coord{itemX, itemY}
		} else {
			panic(fmt.Sprintf("Unknown item player ID %d", itemPlayerID))
		}
	}
	// numQuests: the total number of revealed quests for both players
	var numQuests int
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &numQuests)

	for i := 0; i < numQuests; i++ {
		var questItemName string
		var questPlayerID int
		scanner.Scan()
		fmt.Sscan(scanner.Text(), &questItemName, &questPlayerID)
		if questPlayerID == playerMe {
			g.myQuests = append(g.myQuests, questItemName)
		} else if questPlayerID == playerEnemy {
			g.enemyQuests = append(g.enemyQuests, questItemName)
		} else {
			panic(fmt.Sprintf("Unknown quest player ID %d", questPlayerID))
		}
	}
}

func (g *gameState) findEdges() {
	edges := make(map[coord][]coord)
	for pos, t := range g.tiles {
		if t.up && pos.row > 0 {
			t2pos := coord{pos.col, pos.row - 1}
			if t2 := g.tiles[t2pos]; t2.down {
				edges[pos] = append(edges[pos], t2pos)
			}
		}
		if t.down && pos.row < maxHeight-1 {
			t2pos := coord{pos.col, pos.row + 1}
			if t2 := g.tiles[t2pos]; t2.up {
				edges[pos] = append(edges[pos], t2pos)
			}
		}
		if t.left && pos.col > 0 {
			t2pos := coord{pos.col - 1, pos.row}
			if t2 := g.tiles[t2pos]; t2.right {
				edges[pos] = append(edges[pos], t2pos)
			}
		}
		if t.right && pos.col < maxWidth-1 {
			t2pos := coord{pos.col + 1, pos.row}
			if t2 := g.tiles[t2pos]; t2.left {
				edges[pos] = append(edges[pos], t2pos)
			}
		}
	}

	g.edges = edges
}

func (g *gameState) score() float64 {
	var score float64
	for _, name := range g.myQuests {
		item := g.myItems[name]
		path := g.path(g.me.coord, item)
		if len(path) > 0 {
			score += 1
		}
	}
	longest := g.longestMove()
	score += float64(len(longest)) / (maxWidth * maxHeight)
	return score
}

var ROOTCOORD = coord{-1, -1}

func (g *gameState) path(from, to coord) []coord {
	//log("Entered pathfinding from '%v' to '%v'", from, to)
	frontier := []coord{from}
	visited := map[coord]coord{from: ROOTCOORD}

	g._path(to, frontier, visited)
	result := make([]coord, 0, maxHeight*maxWidth)
	if _, found := visited[to]; found {
		target := to
		for target != ROOTCOORD {
			//log("New path part: '%v'", target)
			result = append(result, target)
			target = visited[target]
		}
	}
	//log("Path after pathfinding: '%v'", result)
	for i := 0; i < len(result)/2; i++ {
		j := len(result) - 1 - i
		result[i], result[j] = result[j], result[i]
	}
	//log("Path after reverse: '%v'", result)

	return result
}

func (g *gameState) _path(target coord, frontier []coord, visited map[coord]coord) {
	//log("Path iteration frontier '%v'; visited '%v'", frontier, visited)
	if len(frontier) == 0 {
		return
	}

	f2 := make([]coord, 0, maxHeight*maxWidth)
	for _, f := range frontier {
		if f == target {
			//log("Pos '%v' from frontier is our target '%v'", f, target)
			return
		}

		neighbors := g.edges[f]
		//log("Neighbors for '%v' are: %v", f, neighbors)
		for _, n := range neighbors {
			if _, found := visited[n]; found {
				continue
			}
			visited[n] = f
			f2 = append(f2, n)
		}

	}
	g._path(target, f2, visited)
}

func (g *gameState) longestMove() []coord {
	from := g.me.coord

	path := []coord{from}
	visited := map[coord]bool{from: true}
	neighbors := g.edges[from]

	for len(neighbors) > 0 {
		next := ROOTCOORD
		for _, n := range neighbors {
			if visited[n] {
				continue
			}
			next = n
		}
		if next != ROOTCOORD {
			path = append(path, next)
			visited[next] = true
			neighbors = g.edges[next]
		} else {
			neighbors = nil
		}
	}

	return path[:len(path)]
}

func (g *gameState) shiftRowLeft(row int) {
	if row < 0 || row >= maxWidth {
		panic(fmt.Sprintf("shiftRowLeft row %d out of range", row))
	}

	extraTile := g.tiles[coord{0, row}]
	for i := 1; i < maxWidth; i++ {
		g.tiles[coord{i - 1, row}] = g.tiles[coord{i, row}]
	}
	g.tiles[coord{maxWidth - 1, row}] = g.me.extraTile
	g.me.extraTile = extraTile

	for name, itemPos := range g.myItems {
		if itemPos.row != row {
			continue
		}
		if itemPos.col == 0 {
			g.myItems[name] = coord{-1, -1}
		} else if itemPos.col == -1 {
			g.myItems[name] = coord{maxWidth - 1, row}
		} else {
			itemPos.col -= 1
			g.myItems[name] = itemPos
		}
	}

	if g.me.row == row {
		g.me.col -= 1
		if g.me.col < 0 {
			g.me.col = maxWidth - 1
		}
	}

	// TODO: enemy items

	g.findEdges()
}

func (g *gameState) shiftRowRight(row int) {
	if row < 0 || row >= maxWidth {
		panic(fmt.Sprintf("shiftRowRight row %d out of range", row))
	}

	extraTile := g.tiles[coord{maxWidth - 1, row}]
	for i := 0; i < maxWidth-1; i++ {
		g.tiles[coord{maxWidth - i, row}] = g.tiles[coord{maxWidth - i - 1, row}]
	}
	g.tiles[coord{0, row}] = g.me.extraTile
	g.me.extraTile = extraTile

	for name, itemPos := range g.myItems {
		if itemPos.row != row {
			continue
		}
		if itemPos.col == maxWidth-1 {
			g.myItems[name] = coord{-1, -1}
		} else if itemPos.col == -1 {
			g.myItems[name] = coord{0, row}
		} else {
			itemPos.col += 1
			g.myItems[name] = itemPos
		}
	}

	if g.me.row == row {
		g.me.col += 1
		if g.me.col > 6 {
			g.me.col = 0
		}
	}

	// TODO: enemy items

	g.findEdges()
}

func (g *gameState) shiftColUp(col int) {
	if col < 0 || col >= maxHeight {
		panic(fmt.Sprintf("shiftColUp col %d out of range", col))
	}

	extraTile := g.tiles[coord{col, 0}]
	for i := 1; i < maxHeight; i++ {
		g.tiles[coord{col, i - 1}] = g.tiles[coord{col, i}]
	}
	g.tiles[coord{col, maxHeight - 1}] = g.me.extraTile
	g.me.extraTile = extraTile

	for name, itemPos := range g.myItems {
		if itemPos.col != col {
			continue
		}
		if itemPos.row == 0 {
			g.myItems[name] = coord{-1, -1}
		} else if itemPos.row == -1 {
			g.myItems[name] = coord{col, maxHeight - 1}
		} else {
			itemPos.row -= 1
			g.myItems[name] = itemPos
		}
	}

	if g.me.col == col {
		g.me.row -= 1
		if g.me.col < 0 {
			g.me.col = maxHeight - 1
		}
	}

	// TODO: enemy items

	g.findEdges()
}

func (g *gameState) shiftColDown(col int) {
	if col < 0 || col >= maxHeight {
		panic(fmt.Sprintf("shiftColDown col %d out of range", col))
	}

	extraTile := g.tiles[coord{col, maxHeight - 1}]
	for i := 1; i < maxHeight; i++ {
		g.tiles[coord{col, maxHeight - i}] = g.tiles[coord{col, maxHeight - i - 1}]
	}
	g.tiles[coord{col, 0}] = g.me.extraTile
	g.me.extraTile = extraTile

	for name, itemPos := range g.myItems {
		if itemPos.col != col {
			continue
		}
		if itemPos.row == maxHeight-1 {
			g.myItems[name] = coord{-1, -1}
		} else if itemPos.row == -1 {
			g.myItems[name] = coord{col, 0}
		} else {
			itemPos.row += 1
			g.myItems[name] = itemPos
		}
	}

	if g.me.col == col {
		g.me.row += 1
		if g.me.col > 6 {
			g.me.col = 0
		}
	}

	// TODO: enemy items

	g.findEdges()
}

func (g *gameState) turnPush() string {
	// estimating pushes
	scores := make(map[string]float64, maxHeight*2+maxWidth*2)
	for i := 0; i < maxWidth; i++ {
		g2 := g.clone()
		g2.shiftColUp(i)
		score := g2.score()
		cmd := cmdTurnPush(i, "UP")
		scores[cmd] = score
		//log("Score for '%s' = %f", cmd, score)

		g3 := g.clone()
		g3.shiftColDown(i)
		score = g3.score()
		cmd = cmdTurnPush(i, "DOWN")
		scores[cmd] = score
		//log("Score for '%s' = %f", cmd, score)
	}
	for i := 0; i < maxHeight; i++ {
		g2 := g.clone()
		g2.shiftRowLeft(i)
		score := g2.score()
		cmd := cmdTurnPush(i, "LEFT")
		scores[cmd] = score
		//log("Score for '%s' = %f", cmd, score)

		g3 := g.clone()
		g3.shiftRowRight(i)
		score = g3.score()
		cmd = cmdTurnPush(i, "RIGHT")
		scores[cmd] = score
		//log("Score for '%s' = %f", cmd, score)
	}

	var bestScore float64
	turns := map[int]string{0: "UP", 1: "DOWN", 2: "LEFT", 3: "RIGHT"}
	bestMove := cmdTurnPush(rand.Intn(maxHeight), turns[rand.Intn(4)])
	for move, score := range scores {
		if score > bestScore {
			//log("Found better move with score %f: '%s'", score, move)
			bestMove = move
			bestScore = score
		}
	}

	return bestMove

}

func (g *gameState) turnMove() string {
	paths := make([][]coord, 0, len(g.myQuests))
	for _, q := range g.myQuests {
		itemPos := g.myItems[q]
		path := g.path(g.me.coord, itemPos)
		if len(path) > 0 {
			paths = append(paths, path)
		}
	}
	//log("Path to items: '%v'", path)

	var path []coord
	if len(paths) == 0 {
		path = g.longestMove()
		path = path[:len(path)/2]
	} else if len(paths) == 1 {
		path = paths[0]
	} else {
		reachableItems := make([]coord, 0, len(paths))
		for _, p := range paths {
			reachableItems = append(reachableItems, p[len(p)-1])
		}
		itemPaths := make(map[coord]map[coord][]coord)
		for i := 0; i < len(reachableItems); i++ {
			for j := i + 1; j < len(reachableItems); j++ {
				i1 := reachableItems[i]
				i2 := reachableItems[j]
				path := g.path(i1, i2)
				if _, found := itemPaths[i1]; !found {
					itemPaths[i1] = make(map[coord][]coord)
				}
				itemPaths[i1][i2] = path

				// + reversed path
				for ii := 0; ii < len(path)/2; ii++ {
					jj := len(path) - 1 - ii
					path[ii], path[jj] = path[jj], path[ii]
				}
				if _, found := itemPaths[i2]; !found {
					itemPaths[i2] = make(map[coord][]coord)
				}
				itemPaths[i2][i1] = path
			}
		}

		// greedy pathfinding
		minLen := math.MaxInt32
		var finalPath []coord
		for _, p := range paths {
			if len(p) < minLen {
				minLen = len(p)
				finalPath = p
			}
		}

		for len(finalPath) <= 20 && len(itemPaths) > 0 {
			end := finalPath[len(finalPath)-1]
			minLen = math.MaxInt32
			var addPath []coord
			for _, p := range itemPaths[end] {
				if len(p) < minLen {
					minLen = len(p)
					addPath = p
				}
			}
			c1, c2 := addPath[0], addPath[len(addPath)-1]
			delete(itemPaths[c1], c2)
			if len(itemPaths[c1]) == 0 {
				delete(itemPaths, c1)
			}
			delete(itemPaths[c2], c1)
			if len(itemPaths[c2]) == 0 {
				delete(itemPaths, c2)
			}
			finalPath = append(finalPath, addPath...)
		}

		path = finalPath
	}
	//log("Path after random move: '%v'", path)

	if len(path) < 2 {
		return "PASS"
	}

	pathStr := "MOVE"
	for i := 0; i < len(path)-1 && i < 20; i++ {
		cur := path[i]
		next := path[i+1]
		if next.col > cur.col {
			pathStr += " RIGHT"
		} else if next.col < cur.col {
			pathStr += " LEFT"
		} else if next.row > cur.row {
			pathStr += " DOWN"
		} else if next.row < cur.row {
			pathStr += " UP"
		}
	}
	return pathStr
}

func (g *gameState) turn() string {
	var cmd string
	switch g.turnType {
	case turnPush:
		cmd = g.turnPush()
	case turnMove:
		cmd = g.turnMove()
	}
	return cmd
}

func main() {
	for {
		t1 := time.Now()
		game := newGameState()
		fmt.Println(game.turn())
		log("TDiff: %d ms", time.Now().Sub(t1).Nanoseconds()/1000000)
	}
}
