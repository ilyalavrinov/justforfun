package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

var myinput = `467..114..
...*......
..35..633.
......#...
617*......
.....+.58.
..592.....
......755.
...$.*....
.664.598..`

func main() {
	// part1()
	part2()
}

type coord struct {
	line, pos int
}

func markAdjuscentToSymbols() map[int]map[int]bool {
	markedLocs := make([]coord, 0)
	lines := strings.Split(myinput, "\n")
	for lineNo, line := range lines {
		for pos, sym := range line {
			if sym != '.' && !unicode.IsNumber(sym) {
				markedLocs = append(markedLocs, coord{lineNo - 1, pos - 1})
				markedLocs = append(markedLocs, coord{lineNo - 1, pos})
				markedLocs = append(markedLocs, coord{lineNo - 1, pos + 1})

				markedLocs = append(markedLocs, coord{lineNo, pos - 1})
				markedLocs = append(markedLocs, coord{lineNo, pos + 1})

				markedLocs = append(markedLocs, coord{lineNo + 1, pos - 1})
				markedLocs = append(markedLocs, coord{lineNo + 1, pos})
				markedLocs = append(markedLocs, coord{lineNo + 1, pos + 1})
			}
		}
	}

	result := make(map[int]map[int]bool)
	for _, c := range markedLocs {
		if c.line == -1 || c.line == len(lines) || c.pos == -1 || c.pos == len(lines[0]) {
			continue
		}

		lineMarks := result[c.line]
		if lineMarks == nil {
			lineMarks = make(map[int]bool)
			result[c.line] = lineMarks
		}
		lineMarks[c.pos] = true
	}
	return result
}

func findApplicableNumbers(applicableCells map[int]map[int]bool) []string {
	result := make([]string, 0)
	for lineNo, line := range strings.Split(myinput, "\n") {
		if applicableCells[lineNo] == nil {
			continue
		}
		isApplicable := false
		curNumber := ""
		for pos, symbol := range line {
			if unicode.IsNumber(symbol) {
				curNumber += string(symbol)
				if applicableCells[lineNo][pos] {
					isApplicable = true
				}
			} else {
				if curNumber != "" {
					if isApplicable {
						result = append(result, curNumber)
						isApplicable = false
						curNumber = ""
					} else {
						curNumber = ""
					}
				}
			}
		}
		if curNumber != "" && isApplicable {
			result = append(result, curNumber)
		}
	}
	return result
}

func part1() {
	sum := 0
	applicableCells := markAdjuscentToSymbols()
	numbersStr := findApplicableNumbers(applicableCells)
	for _, numberStr := range numbersStr {
		num, err := strconv.Atoi(numberStr)
		if err != nil {
			panic(err)
		}
		sum += num
	}
	fmt.Println("SUM:", sum)
}

type coordWithGear struct {
	line, pos int
	gearPos   coord
}

func markAdjuscentToSymbols2() map[int]map[int][]coord {
	markedLocs := make([]coordWithGear, 0)
	lines := strings.Split(myinput, "\n")
	for lineNo, line := range lines {
		for pos, sym := range line {
			if sym == '*' {
				markedLocs = append(markedLocs, coordWithGear{lineNo - 1, pos - 1, coord{lineNo, pos}})
				markedLocs = append(markedLocs, coordWithGear{lineNo - 1, pos, coord{lineNo, pos}})
				markedLocs = append(markedLocs, coordWithGear{lineNo - 1, pos + 1, coord{lineNo, pos}})

				markedLocs = append(markedLocs, coordWithGear{lineNo, pos - 1, coord{lineNo, pos}})
				markedLocs = append(markedLocs, coordWithGear{lineNo, pos + 1, coord{lineNo, pos}})

				markedLocs = append(markedLocs, coordWithGear{lineNo + 1, pos - 1, coord{lineNo, pos}})
				markedLocs = append(markedLocs, coordWithGear{lineNo + 1, pos, coord{lineNo, pos}})
				markedLocs = append(markedLocs, coordWithGear{lineNo + 1, pos + 1, coord{lineNo, pos}})
			}
		}
	}

	result := make(map[int]map[int][]coord)
	for _, c := range markedLocs {
		if c.line == -1 || c.line == len(lines) || c.pos == -1 || c.pos == len(lines[0]) {
			continue
		}

		lineMarks := result[c.line]
		if lineMarks == nil {
			lineMarks = make(map[int][]coord)
			result[c.line] = lineMarks
		}
		lineMarks[c.pos] = append(lineMarks[c.pos], c.gearPos)
	}
	return result
}

type numberWithGears struct {
	str   string
	gears map[coord]bool
}

func findApplicableNumbers2(applicableCells map[int]map[int][]coord) []numberWithGears {
	result := make([]numberWithGears, 0)
	for lineNo, line := range strings.Split(myinput, "\n") {
		if applicableCells[lineNo] == nil {
			continue
		}

		seenGears := make(map[coord]bool)
		curNumber := ""
		for pos, symbol := range line {
			if unicode.IsNumber(symbol) {
				curNumber += string(symbol)
				if len(applicableCells[lineNo][pos]) > 0 {
					for _, gear := range applicableCells[lineNo][pos] {
						seenGears[gear] = true
					}
				}
			} else {
				if curNumber != "" {
					result = append(result, numberWithGears{str: curNumber, gears: seenGears})
				}
				curNumber = ""
				seenGears = make(map[coord]bool)
			}
		}
		if curNumber != "" {
			result = append(result, numberWithGears{str: curNumber, gears: seenGears})
		}
	}
	return result
}

func part2() {
	sum := 0
	applicableCells := markAdjuscentToSymbols2()
	numbers := findApplicableNumbers2(applicableCells)
	gearsToNumbers := make(map[coord][]int)
	for _, number := range numbers {
		num, err := strconv.Atoi(number.str)
		if err != nil {
			panic(err)
		}
		for c := range number.gears {
			gearsToNumbers[c] = append(gearsToNumbers[c], num)
		}
	}
	for _, nums := range gearsToNumbers {
		if len(nums) != 2 {
			continue
		}
		sum += nums[0] * nums[1]
	}
	fmt.Println("SUM:", sum)
}
