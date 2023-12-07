package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const myinput = `32T3K 765
T55J5 684
KK677 28
KTJJT 220
QQQJA 483`

func main() {
	part1()
}

func part1() {
	lines := strings.Split(myinput, "\n")
	cards := make([]card, 0)
	for _, l := range lines {
		cards = append(cards, newCard(l))
	}

	sort.Slice(cards, func(i, j int) bool {
		return cards[i].worseThan(cards[j])
	})

	sum := 0
	for i, c := range cards {
		sum += c.bid * (i + 1)
	}
	fmt.Println("SUM:", sum)
}

const (
	highCard = iota
	onePair
	twoPair
	threeOfAKind
	fullHouse
	fourOfAKind
	fiveOfAKind
)

type card struct {
	hand string
	bid  int

	handType int
}

func newCard(line string) card {
	parts := strings.Split(line, " ")
	bid, err := strconv.Atoi(parts[1])
	if err != nil {
		panic(err)
	}

	return card{
		hand:     parts[0],
		bid:      bid,
		handType: determineHandType(parts[0]),
	}
}

func determineHandType(hand string) int {
	handMap := make(map[rune]int)
	for _, s := range hand {
		handMap[s] = handMap[s] + 1
	}

	replaceJokers(handMap)

	if len(handMap) == 1 {
		return fiveOfAKind
	}
	if len(handMap) == 5 {
		return highCard
	}
	if len(handMap) == 4 {
		return onePair
	}

	maxSameCards := 0
	for _, v := range handMap {
		if v > maxSameCards {
			maxSameCards = v
		}
	}

	if maxSameCards == 4 {
		return fourOfAKind
	}

	if maxSameCards == 2 {
		return twoPair
	}

	if len(handMap) == 2 {
		return fullHouse
	}

	return threeOfAKind
}

var cardVal = map[rune]int{
	'J': 0,
	'1': 1,
	'2': 2,
	'3': 3,
	'4': 4,
	'5': 5,
	'6': 6,
	'7': 7,
	'8': 8,
	'9': 9,
	'T': 10,
	//	'J': 11,
	'Q': 12,
	'K': 13,
	'A': 14,
}

func (c card) worseThan(other card) bool {
	if c.handType != other.handType {
		return c.handType < other.handType
	}

	for i := 0; i < len(c.hand); i++ {
		thisCard := c.hand[i]
		otherCard := other.hand[i]

		if thisCard == otherCard {
			continue
		}
		return cardVal[rune(thisCard)] < cardVal[rune(otherCard)]
	}

	return false
}

func replaceJokers(handMap map[rune]int) {
	jokersNum := handMap['J']
	if jokersNum == 0 || jokersNum == 5 {
		return
	}

	delete(handMap, 'J')
	maxValNow := 0
	var maxCardNow rune
	for card, val := range handMap {
		if val > maxValNow {
			maxValNow = val
			maxCardNow = card
		}
	}

	handMap[maxCardNow] += jokersNum
}
