package main

import (
	"fmt"
	"strconv"
	"strings"
)

var myinput = `two1nine
eightwothree
abcone2threexyz
xtwone3four
4nineeightseven2
zoneight234
7pqrstsixteen`

func main() {
	part2()
}

func part1() {
	sum := 0
	lines := strings.Split(myinput, "\n")
	for _, line := range lines {
		numbers := make([]int, 0, len(lines))
		fmt.Println("Working with line:", line)
		for _, sym := range line {
			if sym < '0' || sym > '9' {
				continue
			}
			num, err := strconv.Atoi(string(sym))
			if err != nil {
				fmt.Println("error converting symbol", sym, " err: %w", err)
				return
			}
			numbers = append(numbers, num)
		}
		lineNumber := 10*numbers[0] + numbers[len(numbers)-1]
		fmt.Println("Got number:", lineNumber)
		sum += lineNumber
	}
	fmt.Println("TOTAL:", sum)
}

var numberMap = map[string]int{
	"one":   1,
	"two":   2,
	"three": 3,
	"four":  4,
	"five":  5,
	"six":   6,
	"seven": 7,
	"eight": 8,
	"nine":  9,
	"1":     1,
	"2":     2,
	"3":     3,
	"4":     4,
	"5":     5,
	"6":     6,
	"7":     7,
	"8":     8,
	"9":     9,
}

func part2() {
	sum := 0
	lines := strings.Split(myinput, "\n")
	for _, line := range lines {
		fmt.Println("Working with line:", line)
		numbers := getNumbersFromLine(line)

		lineNumber := 10*numbers[0] + numbers[len(numbers)-1]
		fmt.Println("Got number:", lineNumber)
		sum += lineNumber
	}
	fmt.Println("TOTAL:", sum)
}

func getNumbersFromLine(line string) []int {
	numbers := make([]int, 0, len(line))
	for i := 0; i < len(line); i++ {
		l2 := line[i:]
		for s, n := range numberMap {
			if strings.HasPrefix(l2, s) {
				numbers = append(numbers, n)
			}
		}
	}
	return numbers
}
