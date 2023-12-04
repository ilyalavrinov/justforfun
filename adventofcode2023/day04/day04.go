package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

const myinput = `Card 1: 41 48 83 86 17 | 83 86  6 31 17  9 48 53
Card 2: 13 32 20 16 61 | 61 30 68 82 17 32 24 19
Card 3:  1 21 53 59 44 | 69 82 63 72 16 21 14  1
Card 4: 41 92 73 84 69 | 59 84 76 51 58  5 54 83
Card 5: 87 83 26 28 32 | 88 30 70 12 93 22 82 36
Card 6: 31 18 13 56 72 | 74 77 10 23 35 67 36 11`

func main() {
	// part1()
	part2()
}

func part1() {
	sum := 0
	for _, line := range strings.Split(myinput, "\n") {
		tmp := strings.Split(line, ":")
		splits := strings.Split(tmp[1], "|")
		winning := splits[0]
		your := splits[1]

		winningMap := make(map[int]bool)
		for _, n := range strings.Split(winning, " ") {
			if n == "" {
				continue
			}
			num, err := strconv.Atoi(n)
			if err != nil {
				panic(err)
			}
			winningMap[num] = true
		}

		pow := -1
		for _, n := range strings.Split(your, " ") {
			if n == "" {
				continue
			}
			num, err := strconv.Atoi(n)
			if err != nil {
				panic(err)
			}
			if winningMap[num] {
				pow++
			}
		}
		if pow != -1 {
			sum += int(math.Pow(2, float64(pow)))
		}
	}
	fmt.Println("SUM:", sum)
}

func part2() {
	wins := originalWins()
	fmt.Println(wins)

	multipliers := make([]int, len(wins))
	for i := 0; i < len(multipliers); i++ {
		multipliers[i] = 1
	}
	for i := 0; i < len(wins); i++ {
		for j := 0; j < multipliers[i]; j++ {
			for k := i + 1; k < i+1+wins[i]; k++ {
				multipliers[k]++
			}
		}
	}

	sum := 0
	for _, val := range multipliers {
		sum += val
	}
	fmt.Println("TOTAL:", sum)
}

func originalWins() []int {
	result := make([]int, 0)
	for _, line := range strings.Split(myinput, "\n") {
		tmp := strings.Split(line, ":")
		splits := strings.Split(tmp[1], "|")
		winning := splits[0]
		your := splits[1]

		winningMap := make(map[int]bool)
		for _, n := range strings.Split(winning, " ") {
			if n == "" {
				continue
			}
			num, err := strconv.Atoi(n)
			if err != nil {
				panic(err)
			}
			winningMap[num] = true
		}

		win := 0
		for _, n := range strings.Split(your, " ") {
			if n == "" {
				continue
			}
			num, err := strconv.Atoi(n)
			if err != nil {
				panic(err)
			}
			if winningMap[num] {
				win++
			}
		}
		result = append(result, win)
	}
	return result
}
