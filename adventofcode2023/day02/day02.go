package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var myinput = `Game 1: 3 blue, 4 red; 1 red, 2 green, 6 blue; 2 green
Game 2: 1 blue, 2 green; 3 green, 4 blue, 1 red; 1 green, 1 blue
Game 3: 8 green, 6 blue, 20 red; 5 blue, 4 red, 13 green; 5 green, 1 red
Game 4: 1 green, 3 red, 6 blue; 3 green, 6 red; 3 green, 15 blue, 14 red
Game 5: 6 red, 1 blue, 3 green; 2 blue, 1 red, 2 green`

func main() {
	// part1()
	part2()
}

func part1() {
	maxCubes := map[string]int{
		"red":   12,
		"green": 13,
		"blue":  14,
	}
	reGame := regexp.MustCompile(`Game ([0-9]*): (.*)`)
	reColor := regexp.MustCompile(`([0-9]*) ([a-z]*)`)

	goodGamesSum := 0
	lines := strings.Split(myinput, "\n")
	for _, line := range lines {
		matches := reGame.FindAllStringSubmatch(line, -1)
		gameIx, err := strconv.Atoi(matches[0][1])
		if err != nil {
			panic(fmt.Sprintf("Cannot convert game %q to number; err: %w", matches[0][1], err))
		}

		fmt.Println("Looking at game:", gameIx)

		gameOK := true

		allParts := matches[0][2]
		parts := strings.Split(allParts, ";")
		for _, part := range parts {
			colors := strings.Split(part, ",")
			for _, color := range colors {
				color = strings.TrimSpace(color)
				colorMatches := reColor.FindAllStringSubmatch(color, -1)
				colorNum, err := strconv.Atoi(colorMatches[0][1])
				if err != nil {
					panic(fmt.Sprintf("Cannot convert color %q to number; err: %w", colorMatches[0][1], err))
				}
				fmt.Println("Colors: number:", colorNum, "; color:", colorMatches[0][2])
				if colorNum > maxCubes[colorMatches[0][2]] {
					gameOK = false
				}
			}
		}

		if gameOK {
			goodGamesSum += gameIx
		}
	}

	fmt.Println("SUM", goodGamesSum)
}

func part2() {
	reGame := regexp.MustCompile(`Game ([0-9]*): (.*)`)
	reColor := regexp.MustCompile(`([0-9]*) ([a-z]*)`)

	powerSum := 0
	lines := strings.Split(myinput, "\n")
	for _, line := range lines {
		minRequired := make(map[string]int, 0)
		matches := reGame.FindAllStringSubmatch(line, -1)
		allParts := matches[0][2]
		parts := strings.Split(allParts, ";")
		for _, part := range parts {
			colors := strings.Split(part, ",")
			for _, color := range colors {
				color = strings.TrimSpace(color)
				colorMatches := reColor.FindAllStringSubmatch(color, -1)
				colorNum, err := strconv.Atoi(colorMatches[0][1])
				if err != nil {
					panic(fmt.Sprintf("Cannot convert color %q to number; err: %w", colorMatches[0][1], err))
				}
				minForColor := minRequired[colorMatches[0][2]]
				if minForColor == 0 {
					minRequired[colorMatches[0][2]] = colorNum
				} else if colorNum > minForColor {
					minRequired[colorMatches[0][2]] = colorNum
				}
			}
		}
		power := 1
		for _, val := range minRequired {
			power *= val
		}
		powerSum += power
	}

	fmt.Println("SUM", powerSum)
}
