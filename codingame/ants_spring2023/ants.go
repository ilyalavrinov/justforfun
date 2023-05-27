package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func cmdBeacon(cellId int, strength int) string {
	return fmt.Sprintf("BEACON %d %d", cellId, strength)
}

func cmdLine(cellIdFrom, cellIdTo int, strength int) string {
	return fmt.Sprintf("LINE %d %d %d", cellIdFrom, cellIdTo, strength)
}

func cmdWait() string {
	return "WAIT"
}

func cmdMessage(msg string) string {
	return fmt.Sprintf("MESSAGE %s", msg)
}

const (
	RESOURCE_NOTHING int = 0
	RESOURCE_EGG     int = 1
	RESOURCE_CRYSTAL int = 2
)

type Cell struct {
	index         int
	cellType      int
	resourceCount int

	myAnts  int
	oppAnts int

	neighbours []*Cell
}

type Field struct {
	numberOfCells int
	cells         map[int]*Cell

	myBases    []*Cell
	enemyBases []*Cell

	cellsWithCrystals []*Cell
	cellsWithEggs     []*Cell
}

func ScanNewField(scanner *bufio.Scanner) Field {
	var field Field

	var inputs []string

	scanner.Scan()
	fmt.Sscan(scanner.Text(), &field.numberOfCells)

	field.cells = make(map[int]*Cell, field.numberOfCells)
	neighbourLists := make(map[int][]int)

	for i := 0; i < field.numberOfCells; i++ {
		// _type: 0 for empty, 1 for eggs, 2 for crystal
		// initialResources: the initial amount of eggs/crystals on this cell
		// neigh0: the index of the neighbouring cell for each direction
		var _type, initialResources, neigh0, neigh1, neigh2, neigh3, neigh4, neigh5 int
		scanner.Scan()
		fmt.Sscan(scanner.Text(), &_type, &initialResources, &neigh0, &neigh1, &neigh2, &neigh3, &neigh4, &neigh5)

		cell := &Cell{
			index:         i,
			cellType:      _type,
			resourceCount: initialResources,
		}
		field.cells[i] = cell

		if _type == RESOURCE_CRYSTAL {
			field.cellsWithCrystals = append(field.cellsWithCrystals, cell)
		} else if _type == RESOURCE_EGG {
			field.cellsWithEggs = append(field.cellsWithEggs, cell)
		}

		neighbourLists[i] = append(neighbourLists[i], neigh0, neigh1, neigh2, neigh3, neigh4, neigh5)
	}

	for cellIx, neighbours := range neighbourLists {
		cell := field.cells[cellIx]
		for _, neighIx := range neighbours {
			if neighIx == -1 {
				continue
			}
			cell.neighbours = append(cell.neighbours, field.cells[neighIx])
		}
	}

	var numberOfBases int
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &numberOfBases)

	scanner.Scan()
	inputs = strings.Split(scanner.Text(), " ")
	for i := 0; i < numberOfBases; i++ {
		myBaseIndex, _ := strconv.ParseInt(inputs[i], 10, 32)
		field.myBases = append(field.myBases, field.cells[int(myBaseIndex)])
	}
	scanner.Scan()
	inputs = strings.Split(scanner.Text(), " ")
	for i := 0; i < numberOfBases; i++ {
		oppBaseIndex, _ := strconv.ParseInt(inputs[i], 10, 32)
		field.enemyBases = append(field.enemyBases, field.cells[int(oppBaseIndex)])
	}

	return field
}

func (f *Field) ScanNewTurn(scanner *bufio.Scanner) {
	for i := 0; i < f.numberOfCells; i++ {
		cell := f.cells[i]

		scanner.Scan()
		fmt.Sscan(scanner.Text(), &cell.resourceCount, &cell.myAnts, &cell.oppAnts)
	}
}

func bfsFindNearestResources(f *Field) (goals []*Cell, level int) {
	myBase := f.myBases[0]
	nextLevel := myBase.neighbours
	visited := make(map[int]bool, len(f.cells))
	visited[myBase.index] = true
	for _, n := range nextLevel {
		visited[n.index] = true
	}

	goals = make([]*Cell, 0)
	level = 0

	bfs := func(frontier []*Cell) {
		for _, cell := range frontier {
			if (cell.cellType == RESOURCE_CRYSTAL || cell.cellType == RESOURCE_EGG) && cell.resourceCount > 0 {
				goals = append(goals, cell)
			} else {
				for _, n := range cell.neighbours {
					if visited[n.index] {
						continue
					}
					visited[n.index] = true
					nextLevel = append(nextLevel, n)
				}
			}
		}
	}

	for len(goals) == 0 {
		level++
		currentLevel := make([]*Cell, 0, len(nextLevel))
		for _, n := range nextLevel {
			currentLevel = append(currentLevel, n)
		}
		nextLevel = make([]*Cell, 0)
		bfs(currentLevel)
	}

	return
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1000000), 1000000)
	field := ScanNewField(scanner)

	currentGoal := -1
	for {
		field.ScanNewTurn(scanner)
		var cmds []string

		if currentGoal != -1 {
			if cell := field.cells[currentGoal]; cell.resourceCount == 0 {
				currentGoal = -1
			}
		}

		if currentGoal == -1 {
			goals, level := bfsFindNearestResources(&field)
			currentGoal = goals[0].index
			cmds = append(cmds, cmdMessage(fmt.Sprintf("GOALLEVEL %d", level)))
		}

		cmds = append(cmds,
			cmdLine(field.myBases[0].index, currentGoal, 1),
			cmdMessage(fmt.Sprintf("CURGOAL at %d", currentGoal)))
		printCmds(cmds...)
	}
}

func printCmds(cmds ...string) {
	result := make([]string, 0, len(cmds))
	for _, cmd := range cmds {
		result = append(result, cmd)
	}
	fmt.Println(strings.Join(result, ";"))
}
