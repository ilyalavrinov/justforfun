package main

import (
	"bufio"
	"fmt"
	"os"
)

func debug(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
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

type Tree struct {
	cellIndex int
	size      int
	isMine    bool
	isDormant bool
}

type GameState struct {
	numberOfCells int
	cells         []Cell
	day           int
	nutrients     int

	mySun, myScore   int
	oppSun, oppScore int
	oppIsWaiting     bool

	numberOfTrees int
	trees         []Tree

	numberOfPossibleMoves int
}

func NewGameState() GameState {
	state := GameState{}

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1000000), 1000000)

	scanner.Scan()
	fmt.Sscan(scanner.Text(), &state.numberOfCells)
	cells := make([]Cell, 0, state.numberOfCells)
	for i := 0; i < state.numberOfCells; i++ {
		// index: 0 is the center cell, the next cells spiral outwards
		// richness: 0 if the cell is unusable, 1-3 for usable cells
		// neigh0: the index of the neighbouring cell for each direction
		var index, richness, neigh0, neigh1, neigh2, neigh3, neigh4, neigh5 int
		scanner.Scan()
		fmt.Sscan(scanner.Text(), &index, &richness, &neigh0, &neigh1, &neigh2, &neigh3, &neigh4, &neigh5)
		cells = append(cells, Cell{
			index:    index,
			richness: richness,
		})
	}
	state.cells = cells

	scanner.Scan()
	fmt.Sscan(scanner.Text(), &state.day)

	scanner.Scan()
	fmt.Sscan(scanner.Text(), &state.nutrients)

	scanner.Scan()
	fmt.Sscan(scanner.Text(), &state.mySun, &state.myScore)

	var _oppIsWaiting int
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &state.oppSun, &state.oppScore, &_oppIsWaiting)
	state.oppIsWaiting = _oppIsWaiting != 0

	scanner.Scan()
	fmt.Sscan(scanner.Text(), &state.numberOfTrees)
	trees := make([]Tree, 0, state.numberOfTrees)
	for i := 0; i < state.numberOfTrees; i++ {
		// cellIndex: location of this tree
		// size: size of this tree: 0-3
		// isMine: 1 if this is your tree
		// isDormant: 1 if this tree is dormant
		var cellIndex, size int
		var isMine, isDormant bool
		var _isMine, _isDormant int
		scanner.Scan()
		fmt.Sscan(scanner.Text(), &cellIndex, &size, &_isMine, &_isDormant)
		isMine = _isMine != 0
		isDormant = _isDormant != 0
		state.trees = append(trees, Tree{
			cellIndex: cellIndex,
			size:      size,
			isMine:    isMine,
			isDormant: isDormant,
		})
	}
	state.trees = trees

	fmt.Sscan(scanner.Text(), &state.numberOfPossibleMoves)
	for i := 0; i < state.numberOfPossibleMoves; i++ {
		scanner.Scan()
		possibleMove := scanner.Text()
		_ = possibleMove // to avoid unused error
	}

	return state
}

func main() {
	for {
		_ := NewGameState()
		cmdWaitMsg("Hello")
	}
}
