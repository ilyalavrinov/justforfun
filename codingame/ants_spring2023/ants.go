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

type Field struct {
	numberOfCells int
	cellTypes     []int

	myBases    []int
	enemyBases []int

	cellsWithCrystals map[int]int
}

func (f *Field) reset() {
	f.cellsWithCrystals = make(map[int]int, f.numberOfCells)
}

func ScanNewField(scanner *bufio.Scanner) Field {
	var field Field

	var inputs []string

	scanner.Scan()
	fmt.Sscan(scanner.Text(), &field.numberOfCells)

	field.cellTypes = make([]int, field.numberOfCells)
	field.reset()

	for i := 0; i < field.numberOfCells; i++ {
		// _type: 0 for empty, 1 for eggs, 2 for crystal
		// initialResources: the initial amount of eggs/crystals on this cell
		// neigh0: the index of the neighbouring cell for each direction
		var _type, initialResources, neigh0, neigh1, neigh2, neigh3, neigh4, neigh5 int
		scanner.Scan()
		fmt.Sscan(scanner.Text(), &_type, &initialResources, &neigh0, &neigh1, &neigh2, &neigh3, &neigh4, &neigh5)

		field.cellTypes[i] = _type
		if _type == RESOURCE_CRYSTAL {
			field.cellsWithCrystals[i] = initialResources
		}
	}
	var numberOfBases int
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &numberOfBases)
	field.myBases = make([]int, 0, numberOfBases)
	field.enemyBases = make([]int, 0, numberOfBases)

	scanner.Scan()
	inputs = strings.Split(scanner.Text(), " ")
	for i := 0; i < numberOfBases; i++ {
		myBaseIndex, _ := strconv.ParseInt(inputs[i], 10, 32)
		field.myBases = append(field.myBases, int(myBaseIndex))
	}
	scanner.Scan()
	inputs = strings.Split(scanner.Text(), " ")
	for i := 0; i < numberOfBases; i++ {
		oppBaseIndex, _ := strconv.ParseInt(inputs[i], 10, 32)
		field.enemyBases = append(field.enemyBases, int(oppBaseIndex))
	}

	return field
}

func (f *Field) ScanNewTurn(scanner *bufio.Scanner) {
	f.reset()
	for i := 0; i < f.numberOfCells; i++ {
		// resources: the current amount of eggs/crystals on this cell
		// myAnts: the amount of your ants on this cell
		// oppAnts: the amount of opponent ants on this cell
		var resources, myAnts, oppAnts int
		scanner.Scan()
		fmt.Sscan(scanner.Text(), &resources, &myAnts, &oppAnts)
		if f.cellTypes[i] == RESOURCE_CRYSTAL {
			f.cellsWithCrystals[i] = resources
		}
	}
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1000000), 1000000)
	field := ScanNewField(scanner)
	for {
		field.ScanNewTurn(scanner)
		maxCrystalCellId := 0
		maxCrystalCount := 0
		for cellId, eggCount := range field.cellsWithCrystals {
			if eggCount > maxCrystalCount {
				maxCrystalCellId = cellId
				maxCrystalCount = eggCount
			}
		}

		printCmds(
			cmdLine(field.myBases[0], maxCrystalCellId, 1),
			cmdMessage(fmt.Sprintf("MAXCNT %d at %d", maxCrystalCount, maxCrystalCellId)),
		)
	}
}

func printCmds(cmds ...string) {
	result := make([]string, 0, len(cmds))
	for _, cmd := range cmds {
		result = append(result, cmd)
	}
	fmt.Println(strings.Join(result, ";"))
}
