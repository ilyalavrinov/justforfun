package main

import (
	"fmt"
	"strconv"
	"strings"
)

/*
const myinput = `Time:      7  15   30
Distance:  9  40  200`
*/

/*
const myinput = `Time:        44     89     96     91
Distance:   277   1136   1890   1768`
*/

const myinput = `Time:        44899691
Distance:   277113618901768`

func main() {
	part1()
}

func part1() {
	lines := strings.Split(myinput, "\n")
	lineTime := lines[0]
	lineDist := lines[1]
	lineTimes := strings.Split(lineTime, ":")[1]
	lineDists := strings.Split(lineDist, ":")[1]

	timesStr := strings.Split(lineTimes, " ")
	distsStr := strings.Split(lineDists, " ")

	times := make([]int, 0)
	dists := make([]int, 0)

	for _, val := range timesStr {
		val := strings.TrimSpace(val)
		if val == "" {
			continue
		}
		res, err := strconv.Atoi(val)
		if err != nil {
			panic(err)
		}
		times = append(times, res)
	}

	for _, val := range distsStr {
		val := strings.TrimSpace(val)
		if val == "" {
			continue
		}
		res, err := strconv.Atoi(val)
		if err != nil {
			panic(err)
		}
		dists = append(dists, res)
	}

	res := 1
	for i := 0; i < len(times); i++ {
		res *= calcNumOfWins(times[i], dists[i])
	}
	fmt.Println("RESULT:", res)
}

func calcNumOfWins(t, d int) int {
	fmt.Println("CALC:", t, d)
	wins := 0
	for i := 0; i <= t; i++ {
		speed := i
		distTraveled := (t - i) * speed
		if distTraveled > d {
			wins++
		}
	}
	fmt.Println("CALC WINS:", wins)
	return wins
}
