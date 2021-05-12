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

type Command interface {
	Do()
}

type cmdGrow struct {
	ix int
}

func NewCmdGrow(ix int) Command {
	return cmdGrow{
		ix: ix,
	}
}

func (c cmdGrow) Do() {
	fmt.Printf("GROW %d\n", c.ix)
}

type cmdWait struct {
	msg string
}

func NewCmdWaitMsg(msg string) Command {
	return cmdWait{
		msg: msg,
	}
}

func NewCmdWait() Command {
	return NewCmdWaitMsg("")
}

func (c cmdWait) Do() {
	if c.msg != "" {
		fmt.Printf("WAIT %s\n", c.msg)
	}
	fmt.Println("WAIT")
}

type cmdComplete struct {
	ix int
}

func NewCmdComplete(ix int) Command {
	return cmdComplete{
		ix: ix,
	}
}

func (c cmdComplete) Do() {
	fmt.Printf("COMPLETE %d\n", c.ix)
}

type cmdSeed struct {
	from, to int
}

func NewCmdSeed(from, to int) Command {
	return cmdSeed{
		from: from,
		to:   to,
	}
}

func (c cmdSeed) Do() {
	fmt.Printf("SEED %d %d\n", c.from, c.to)
}

func numTreesBySize(state GameState, size int) int {
	n := 0
	for _, t := range state.trees {
		if t.isMine {
			continue
		}
		if t.size != size {
			continue
		}
		n++
	}
	return n
}

func costSeed(state GameState) int {
	return numTreesBySize(state, 0)
}

func costGrow(state GameState, sizeFrom int) int {
	cost := 0
	switch sizeFrom {
	case 0:
		cost = 1
	case 1:
		cost = 3
	case 2:
		cost = 7
	}

	return cost + numTreesBySize(state, sizeFrom+1)
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
	field := Field{}

	scanner.Scan()
	fmt.Sscan(scanner.Text(), &field.numberOfCells)
	cells := make([]Cell, 0, field.numberOfCells)
	for i := 0; i < field.numberOfCells; i++ {
		var index, richness, neigh0, neigh1, neigh2, neigh3, neigh4, neigh5 int
		scanner.Scan()
		fmt.Sscan(scanner.Text(), &index, &richness, &neigh0, &neigh1, &neigh2, &neigh3, &neigh4, &neigh5)
		cells = append(cells, Cell{
			index:    index,
			richness: richness,
		})
	}
	field.cells = cells

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
	state := GameState{
		field: field,
	}

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

	scanner.Scan()
	fmt.Sscan(scanner.Text(), &state.numberOfPossibleMoves)
	for i := 0; i < state.numberOfPossibleMoves; i++ {
		scanner.Scan()
		possibleMove := scanner.Text()
		_ = possibleMove // to avoid unused error
	}

	return state
}

func greedyComplete(state GameState) Command {
	debug("greedy complete start")
	bestIx := state.field.numberOfCells + 1
	bestScore := 0
	for _, t := range state.trees {
		debugf("new tree check: cellIndex %d; mine: %t\n", t.cellIndex, t.isMine)
		if !t.isMine {
			continue
		}
		if t.size != 3 {
			continue
		}
		score := (state.field.cells[t.cellIndex].richness-1)*2 + state.nutrients
		if score > bestScore {
			debugf("New best score %d on cell %d (was score %d cell %d)\n", score, t.cellIndex, bestScore, bestIx)
			bestIx = t.cellIndex
			bestScore = score
		}
	}

	var res Command
	if bestIx < state.field.numberOfCells+1 {
		res = NewCmdComplete(bestIx)
	}
	debugf("greedy complete finish res=%+v", res)
	return res
}

func greedyGrow(state GameState) Command {
	debug("greedy grow start")
	bestIx := state.field.numberOfCells + 1
	bestSize := 0
	for _, t := range state.trees {
		debugf("new tree check: cellIndex %d; mine: %t\n", t.cellIndex, t.isMine)
		if !t.isMine {
			continue
		}
		if t.size == 3 {
			continue
		}
		if t.size > bestSize {
			debugf("New best size %d on cell %d (was score %d cell %d)\n", t.size, t.cellIndex, bestSize, bestIx)
			bestIx = t.cellIndex
			bestSize = t.size
		}
	}

	var res Command
	if bestIx < state.field.numberOfCells+1 {
		res = NewCmdGrow(bestIx)
	}
	debugf("greedy grow finish res=%+v", res)
	return res
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
		c := greedyComplete(state)
		if c == nil {
			c = greedyGrow(state)
			if c == nil {
				c = NewCmdWait()
			}
		}

		c.Do()
		firstScan = false
	}
}
