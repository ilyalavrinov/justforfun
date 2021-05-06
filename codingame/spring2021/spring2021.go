package main

import (
	"bufio"
	"fmt"
	"os"
)

const (
	priceCompleteSun       = 4
	priceCompleteNutrients = 1
)

func debug(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
}

func debugf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
}

func cmdGrow(ix int) {
	fmt.Printf("GROW %d\n", ix)
}

func cmdWait() {
	cmdWaitMsg("")
}

func cmdWaitMsg(msg string) {
	if msg != "" {
		fmt.Printf("WAIT %s\n", msg)
	}
	fmt.Println("WAIT")
}

func cmdComplete(ix int) {
	fmt.Printf("COMPLETE %d\n", ix)
}

func cmdSeed(from, to int) {
	fmt.Printf("SEED %d %d\n", from, to)
}

type Cell struct {
	index    int
	richness int
}

type Field struct {
	numberOfCells int
	cells         []Cell
}

func NewField(scanner *bufio.Scanner) Field {
	debug("new field started")
	field := Field{}

	scanner.Scan()
	fmt.Sscan(scanner.Text(), &field.numberOfCells)
	cells := make([]Cell, 0, field.numberOfCells)
	for i := 0; i < field.numberOfCells; i++ {
		debugf("new field: reading cell %d of %d\n", i, field.numberOfCells)
		var index, richness, neigh0, neigh1, neigh2, neigh3, neigh4, neigh5 int
		scanner.Scan()
		fmt.Sscan(scanner.Text(), &index, &richness, &neigh0, &neigh1, &neigh2, &neigh3, &neigh4, &neigh5)
		cells = append(cells, Cell{
			index:    index,
			richness: richness,
		})
	}
	field.cells = cells

	debug("new field finished")
	return field
}

type Tree struct {
	cellIndex int
	size      int
	isMine    bool
	isDormant bool
}

type GameState struct {
	field Field

	day       int
	nutrients int

	mySun, myScore   int
	oppSun, oppScore int
	oppIsWaiting     bool

	numberOfTrees int
	trees         []Tree

	numberOfPossibleMoves int
}

func NewGameState(field Field, scanner *bufio.Scanner) GameState {
	debug("new state started")
	state := GameState{
		field: field,
	}

	debug("new state: scan day")
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &state.day)

	debug("new state: scan nutrients")
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &state.nutrients)

	debug("new state: scan my sun & score")
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &state.mySun, &state.myScore)

	debug("new state: scan opp sun & score")
	var _oppIsWaiting int
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &state.oppSun, &state.oppScore, &_oppIsWaiting)
	state.oppIsWaiting = _oppIsWaiting != 0

	debug("new state: scan number of trees")
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &state.numberOfTrees)
	trees := make([]Tree, 0, state.numberOfTrees)
	for i := 0; i < state.numberOfTrees; i++ {
		debugf("new state: reading tree %d of %d\n", i, state.numberOfTrees)
		var cellIndex, size int
		var isMine, isDormant bool
		var _isMine, _isDormant int
		scanner.Scan()
		fmt.Sscan(scanner.Text(), &cellIndex, &size, &_isMine, &_isDormant)
		isMine = _isMine != 0
		isDormant = _isDormant != 0
		trees = append(trees, Tree{
			cellIndex: cellIndex,
			size:      size,
			isMine:    isMine,
			isDormant: isDormant,
		})
	}
	state.trees = trees

	debug("new state: scan possible moves")
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &state.numberOfPossibleMoves)
	for i := 0; i < state.numberOfPossibleMoves; i++ {
		scanner.Scan()
		possibleMove := scanner.Text()
		_ = possibleMove // to avoid unused error
	}

	debug("new state finished")
	return state
}

func greedyComplete(state GameState) {
	debug("greedy start")
	bestIx := state.field.numberOfCells + 1
	bestScore := 0
	for _, t := range state.trees {
		debugf("new tree check: cellIndex %d; mine: %t\n", t.cellIndex, t.isMine)
		if !t.isMine {
			continue
		}
		score := (state.field.cells[t.cellIndex].richness-1)*2 + state.nutrients
		if score > bestScore {
			debug()
			bestIx = t.cellIndex
			bestScore = score
		}
	}

	if bestIx < state.field.numberOfCells+1 {
		cmdComplete(bestIx)
	} else {
		cmdWait()
	}
	debug("greedy finish")
}

func main() {
	firstScan := true
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1000000), 1000000)

	field := NewField(scanner)
	for {
		if !firstScan {
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Buffer(make([]byte, 1000000), 1000000)
		}
		state := NewGameState(field, scanner)
		debugf("new day; cells: %d, trees: %d\n", len(state.field.cells), len(state.trees))
		greedyComplete(state)

		firstScan = false
	}
}
