package main

import (
	"fmt"
	"sort"
)

const (
	BREW = "BREW"
	WAIT = "WAIT"
	REST = "REST"
	CAST = "CAST"
)

type inventory struct {
	inv0, inv1, inv2, inv3, score int
}

func (i *inventory) scan() {
	fmt.Scan(&i.inv0, &i.inv1, &i.inv2, &i.inv3, &i.score)
}

type action struct {
	actionId                                                   int
	actionType                                                 string
	delta0, delta1, delta2, delta3, price, tomeIndex, taxCount int
	castable, repeatable                                       bool
}

func (a *action) scan() {
	var _castable, _repeatable int
	fmt.Scan(&a.actionId, &a.actionType, &a.delta0, &a.delta1, &a.delta2, &a.delta3, &a.price, &a.tomeIndex, &a.taxCount, &_castable, &_repeatable)

	a.castable = _castable != 0
	a.repeatable = _repeatable != 0
}

func main() {
	for {
		turn()
	}
}

func turn() {
	var actionCount int
	fmt.Scan(&actionCount)

	actions := make([]action, 0, actionCount)

	for i := 0; i < actionCount; i++ {
		a := action{}
		a.scan()
		actions = append(actions, a)
	}

	invMe := inventory{}
	invMe.scan()

	invEnemy := inventory{}
	invEnemy.scan()

	brewable := make([]action, 0, len(actions))
	castableNow := make([]action, 0, len(actions))
	for _, a := range actions {
		if canBrew(invMe, a) {
			brewable = append(brewable, a)
		} else if canCast(invMe, a) {
			castableNow = append(castableNow, a)
		}
	}

	if len(brewable) > 0 {
		sort.Slice(brewable, func(i, j int) bool {
			return brewable[i].price < brewable[j].price
		})
		best := brewable[len(brewable)-1]
		fmt.Printf("%s %d\n", BREW, best.actionId)
	} else if len(castableNow) > 0 {
		sort.Slice(castableNow, func(i, j int) bool {
			iNet := castableNow[i].delta0 + castableNow[i].delta1 + castableNow[i].delta2 + castableNow[i].delta3
			jNet := castableNow[j].delta0 + castableNow[j].delta1 + castableNow[j].delta2 + castableNow[j].delta3
			return iNet < jNet
		})
		best := castableNow[len(castableNow)-1]
		fmt.Printf("%s %d\n", CAST, best.actionId)
	} else {
		fmt.Println(REST)
	}
}

func canBrew(i inventory, a action) bool {
	if a.actionType != BREW {
		return false
	}

	if i.inv0+a.delta0 < 0 {
		return false
	} else if i.inv1+a.delta1 < 0 {
		return false
	} else if i.inv2+a.delta2 < 0 {
		return false
	} else if i.inv3+a.delta3 < 0 {
		return false
	}
	return true
}

func canCast(i inventory, a action) bool {
	if a.actionType != CAST {
		return false
	}

	if !a.castable {
		return false
	}

	if i.inv0+a.delta0 < 0 {
		return false
	} else if i.inv1+a.delta1 < 0 {
		return false
	} else if i.inv2+a.delta2 < 0 {
		return false
	} else if i.inv3+a.delta3 < 0 {
		return false
	}
	return true
}
