package main

import (
	"fmt"
	"strconv"
	"strings"
)

const myinput = `0 3 6 9 12 15
1 3 6 10 15 21
10 13 16 21 30 45`

func main() {
	sum := 0
	for _, line := range strings.Split(myinput, "\n") {
		numbersStr := strings.Split(line, " ")
		numbers := make([]int, 0)
		for _, nStr := range numbersStr {
			number, err := strconv.Atoi(nStr)
			if err != nil {
				panic(err)
			}
			numbers = append(numbers, number)
		}
		for i := 0; i < len(numbers)/2; i++ {
			numbers[i], numbers[len(numbers)-i-1] = numbers[len(numbers)-i-1], numbers[i]
		}
		sum += predictNextNum(numbers)
	}
	fmt.Println("SUM:", sum)
}

func predictNextNum(numbers []int) int {
	all0 := true
	for _, n := range numbers {
		if n != 0 {
			all0 = false
			break
		}
	}
	if all0 {
		return 0
	}

	newNumbers := make([]int, 0, len(numbers)-1)
	for i := 1; i < len(numbers); i++ {
		newNumbers = append(newNumbers, numbers[i]-numbers[i-1])
	}

	return predictNextNum(newNumbers) + numbers[len(numbers)-1]
}
