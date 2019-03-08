package main

import "fmt"
import "math"
import "math/rand"
import "os"
import "strings"
import "bufio"
import "sort"

const (
    logData = false  // log level for data reading and structure preparation
    logTurn = false  // log level for supplimentary turn-related data
    logBattle = true
    logStepSummon = true
    logStepLethal = true
    logStepGuards = true
    logStepAttack = true
)

func log(format string, args... interface{}) {
    format += "\n"
    if len(args) == 0 {
        fmt.Fprintf(os.Stderr, format)
    } else {
        fmt.Fprintf(os.Stderr, format, args...)
    }
}

// PlayerInfo stores some data about a player
type PlayerInfo struct {
    health int
    mana int
    deck int
    runes int
}

// Players contains info about all players
type Players struct {
    me PlayerInfo
    enemy PlayerInfo
}


const (
    maxMana = 12
    maxHandCards = 8
    maxBoardCards = 6
    maxDraftTurns = 30
    maxBattleTurns = 50
)


const (
    typeCreature = 0
    typeItemGreen = 1
    typeItemRed = 2
    typeItemBlue = 3
)

const (
    abilityLethal = "L"
    abilityCharge = "C"
    abilityBreakthrough = "B"
    abilityWard = "W"
    abilityGuard = "G"
    abilityDrain = "D"
)

const (
    locationMyHand = 0
    locationMyBoard = 1
    locationEnemyBoard = -1
)

// Card contains card-related data
type Card struct {
    cardNumber int
    instanceID int
    location int
    cardType int
    cost int
    attack int
    defense int

    abilities string
    isBreakthrough bool
    isCharge bool
    isGuard bool
    isDrain bool
    isLethal bool
    isWard bool

    myHealthChange int
    opponentHealthChange int
    cardDraw int
}

func (c Card) cardPower() float64 {
    var power float64

    power += float64(c.attack)
    power += float64(c.defense)
    if c.isBreakthrough { power += 1 }
    if c.isCharge { power += 2 }
    if c.isWard { power += 1 }
    if c.isGuard { power += float64(c.attack + c.defense) / 2 }
    if c.isDrain { power += float64(c.attack) * 2 }
    if c.isLethal { power += 1 }

    power += float64(c.myHealthChange)
    power += float64(-c.opponentHealthChange)
    power += float64(c.cardDraw) / 2

    return power / float64(c.cost)
}

// CardSet contains a set of cards + some accumulated stats over them
type CardSet struct {
    cards []Card

    cardsWithType map[int]map[int]bool // cardType -> indices of cards (e.g. cardsWithType[cardtypeCreature] = set([5, 7, 12]))
    creaturesWithAbility map[string]map[int]bool // ability -> indices of cards
}

// NewCardSet created a new CardSet using provided slice of raw strings
func NewCardSet(cards []string) *CardSet {
    result := &CardSet{cards: make([]Card, 0, len(cards)),
                       cardsWithType: make(map[int]map[int]bool, len(cards)),
                       creaturesWithAbility: make(map[string]map[int]bool, len(cards))}

    // initializing all internal maps correctly
    result.cardsWithType[typeCreature] = make(map[int]bool, len(cards))
    result.cardsWithType[typeItemRed] = make(map[int]bool, len(cards))
    result.cardsWithType[typeItemBlue] = make(map[int]bool, len(cards))
    result.cardsWithType[typeItemGreen] = make(map[int]bool, len(cards))

    result.creaturesWithAbility[abilityBreakthrough] = make(map[int]bool, len(cards))
    result.creaturesWithAbility[abilityCharge] = make(map[int]bool, len(cards))
    result.creaturesWithAbility[abilityDrain] = make(map[int]bool, len(cards))
    result.creaturesWithAbility[abilityGuard] = make(map[int]bool, len(cards))
    result.creaturesWithAbility[abilityLethal] = make(map[int]bool, len(cards))
    result.creaturesWithAbility[abilityWard] = make(map[int]bool, len(cards))

    if logData { log("Starting generation of a new set using %d cards", len(cards)) }

    for _, c := range cards {
        if logData { log("Processing line: %s", c) }
        var cardNumber, instanceID, location, cardType, cost, attack, defense int
        var abilities string
        var myHealthChange, opponentHealthChange, cardDraw int
        fmt.Sscan(c, &cardNumber, &instanceID, &location, &cardType, &cost, &attack, &defense, &abilities, &myHealthChange, &opponentHealthChange, &cardDraw)
        card := Card{cardNumber: cardNumber,
                     instanceID: instanceID,
                     location: location,
                     cardType: cardType,
                     cost: cost,
                     attack: attack,
                     defense: defense,
                     abilities: abilities,
                     isBreakthrough: strings.Index(abilities, abilityBreakthrough) >= 0,
                     isCharge: strings.Index(abilities, abilityCharge) >= 0,
                     isGuard: strings.Index(abilities, abilityGuard) >= 0,
                     isDrain: strings.Index(abilities, abilityDrain) >= 0,
                     isLethal: strings.Index(abilities, abilityLethal) >= 0,
                     isWard: strings.Index(abilities, abilityWard) >= 0,
                     myHealthChange: myHealthChange,
                     opponentHealthChange: opponentHealthChange,
                     cardDraw: cardDraw}
        if logData { log("Read card: %+v\n", card) }

        result.AddCard(card)

    }
    if logData { log("Done with new set generation, total cards: %d", len(result.cards)) }

    return result
}

// AddCard adds a card to a CardSet and calculated nessesary stats
func (s *CardSet) AddCard(card Card) int {
    s.cards = append(s.cards, card)

    cardIx := len(s.cards) - 1
    s.addCardToCaches(card, cardIx)
    return cardIx
}


// RemoveCardByIndex removes a card from a CardSet and purges all associated info
func (s *CardSet) RemoveCardByIndex(ix int) {
    if ix >= len(s.cards) {
        panic("Incorrect index on remove")
    }

    s.cards = append(s.cards[:ix], s.cards[ix+1:]...)
    s.rebuildCaches()
}

func (s *CardSet) addCardToCaches(card Card, ix int) {
    s.cardsWithType[card.cardType][ix] = true

    if card.cardType == typeCreature {
        if card.isGuard {
            s.creaturesWithAbility[abilityGuard][ix] = true
        }
        if card.isCharge {
            s.creaturesWithAbility[abilityCharge][ix] = true
        }
        if card.isBreakthrough {
            s.creaturesWithAbility[abilityBreakthrough][ix] = true
        }
        if card.isWard {
            s.creaturesWithAbility[abilityWard][ix] = true
        }
        if card.isDrain {
            s.creaturesWithAbility[abilityDrain][ix] = true
        }
        if card.isLethal {
            s.creaturesWithAbility[abilityLethal][ix] = true
        }
    }
}

func (s *CardSet) rebuildCaches() {
    // resetting maps
    s.cardsWithType[typeCreature] = make(map[int]bool, len(s.cards))
    s.cardsWithType[typeItemRed] = make(map[int]bool, len(s.cards))
    s.cardsWithType[typeItemBlue] = make(map[int]bool, len(s.cards))
    s.cardsWithType[typeItemGreen] = make(map[int]bool, len(s.cards))

    s.creaturesWithAbility[abilityBreakthrough] = make(map[int]bool, len(s.cards))
    s.creaturesWithAbility[abilityCharge] = make(map[int]bool, len(s.cards))
    s.creaturesWithAbility[abilityDrain] = make(map[int]bool, len(s.cards))
    s.creaturesWithAbility[abilityGuard] = make(map[int]bool, len(s.cards))
    s.creaturesWithAbility[abilityLethal] = make(map[int]bool, len(s.cards))
    s.creaturesWithAbility[abilityWard] = make(map[int]bool, len(s.cards))

    for ix, card := range s.cards {
        s.addCardToCaches(card, ix)
    }
}

// CardsInGame provides all cards visible to a player
type CardsInGame struct {
    myHand CardSet
    myBoard CardSet
    enemyBoard CardSet
}


 func cmdPick(n int) {
     if n < 0 || n > 2 {
         panic("Incorrect ID for card picking")
     }
     fmt.Printf("PICK %d\n", n)
 }


func cmdAttack(myCardID, enemyCardID int) string {
    return fmt.Sprintf("ATTACK %d %d", myCardID, enemyCardID)
}

func cmdSummon(cardID int) string {
    if logStepSummon { log("Summoning id %d", cardID) }
    return fmt.Sprintf("SUMMON %d", cardID)
}

func sendCommands(commands []string) {
    cmd := "PASS"
    if len(commands) == 0 {
        if logTurn { log("List of commands is empty, PASS will be sent") }
    } else {
        cmd = strings.Join(commands, ";")
    }
    fmt.Println(cmd)
}

func readPlayers() *Players {

    p := &Players{}

    var playerHealth, playerMana, playerDeck, playerRune int

    fmt.Scan(&playerHealth, &playerMana, &playerDeck, &playerRune)
    me := PlayerInfo{health: playerHealth,
                     mana: playerMana,
                     deck: playerDeck,
                     runes: playerRune}
    p.me = me

    fmt.Scan(&playerHealth, &playerMana, &playerDeck, &playerRune)
    enemy := PlayerInfo{health: playerHealth,
                        mana: playerMana,
                        deck: playerDeck,
                        runes: playerRune}
    p.enemy = enemy


    if logData { log("Read player data: %+v\n",*p) }
    return p
}



func readCards() *CardsInGame {
    var cardCount int
    fmt.Scan(&cardCount)
    if logData { log("Read card count: %d\n", cardCount) }

    myHandCards := make([]string, 0, cardCount)
    myBoardCards := make([]string, 0, cardCount)
    enemyBoardCards := make([]string, 0, cardCount)

    scanner := bufio.NewScanner(os.Stdin)
    cardsRead := 0
    for cardsRead < cardCount && scanner.Scan() {
        line := scanner.Text()
        var i, j, location int
        fmt.Sscan(line, &i, &j, &location)
        if location == locationEnemyBoard {
            enemyBoardCards = append(enemyBoardCards, line)
            if logData { log("Card goes to enemyBoard: %s", line) }
        } else if location == locationMyBoard {
            myBoardCards = append(myBoardCards, line)
            if logData { log("Card goes to myBoard: %s", line) }
        } else if location == locationMyHand {
            myHandCards = append(myHandCards, line)
            if logData { log("Card goes to myHand: %s", line) }
        } else {
            log("PANICKING!")
            panic(fmt.Sprintf("Unknown location: %d", location))
        }
        cardsRead++
    }

    result := &CardsInGame {
                    myHand: *NewCardSet(myHandCards),
                    myBoard: *NewCardSet(myBoardCards),
                    enemyBoard: *NewCardSet(enemyBoardCards)}

    cardSum := len(myHandCards) + len(myBoardCards) + len(enemyBoardCards)
    if cardSum != cardCount {
        log("PANICKING!")
        panic(fmt.Sprintf("Card count mismatch: actual: %d; expected: %d", cardSum, cardCount))
    }

    return result
}


func draftTurn(cards CardSet) {
    // add picking items when item logic is added
    n := rand.Intn(3)
    creatures := make([]int, 0, 3)
    for ix := range cards.cardsWithType[typeCreature] {
        creatures = append(creatures, ix)
    }
    if len(creatures) > 0 {
        // better power - better card!
        sort.Slice(creatures, func(i, j int) bool {
            return cards.cards[i].cardPower() > cards.cards[j].cardPower()
        })
        // best-powered card is here
        n = creatures[0]
    }
    cmdPick(n)
}

func stepSummon(me PlayerInfo, cards *CardsInGame) (cmdNow, cmdAfter []string) {
    if logBattle { log("Starting summoning") }
    cmdNow = make([]string, 0, len(cards.myHand.cards))
    cmdAfter = make([]string, 0, len(cards.myHand.cards))

    summonCards := make([]int, 0, len(cards.myHand.cardsWithType[typeCreature]))
    for ix := range cards.myHand.cardsWithType[typeCreature] {
        summonCards = append(summonCards, ix)
    }
    sort.Slice(summonCards, func (i, j int) bool {
        return cards.myHand.cards[summonCards[i]].cardPower() > cards.myHand.cards[summonCards[j]].cardPower()
    })
    for _, ix := range summonCards {
        c := cards.myHand.cards[ix]
        if c.cost <= me.mana {
            me.mana -= c.cost
            if c.isCharge {
                cards.myBoard.AddCard(c)
                cmdNow = append(cmdNow, cmdSummon(c.instanceID))
            } else {
                cmdAfter = append(cmdAfter, cmdSummon(c.instanceID))
            }
        }
    }

    return
}

func battleTurn(players Players, cards CardsInGame) {
    if logBattle { log("Starting battle turn. On my board: %d cards. On enemy board: %d cards", len(cards.myBoard.cards), len(cards.enemyBoard.cards)) }
    commands := make([]string, 0, 0)
    summonNow, summonAfter := stepSummon(players.me, &cards)
    commands = append(commands, summonNow...)

    myArmy := NewArmy(players.me, false)
    for _, c := range cards.myBoard.cards {
        myArmy.AddCardUnit(c, true)
    }

    enemyArmy := NewArmy(players.enemy, true)
    for _, c := range cards.enemyBoard.cards {
        enemyArmy.AddCardUnit(c, true)
    }

    attacks := calcAttacksFrontierGreedy(*myArmy, *enemyArmy)
    for _, a := range attacks {
        commands = append(commands, a.Cmd())
    }
    commands = append(commands, summonAfter...)

    sendCommands(commands)
}

/*
Frontier-based strategy
*/
type InstanceID int

type Abilities map[string]bool

func NewAbilities(abilities string) Abilities {
    a := make(Abilities, 6)

    a[abilityBreakthrough] = (strings.Index(abilities, abilityBreakthrough) >= 0)
    a[abilityCharge] = (strings.Index(abilities, abilityCharge) >= 0)
    a[abilityGuard] = (strings.Index(abilities, abilityGuard) >= 0)
    a[abilityDrain] = (strings.Index(abilities, abilityDrain) >= 0)
    a[abilityLethal] = (strings.Index(abilities, abilityLethal) >= 0)
    a[abilityWard] = (strings.Index(abilities, abilityWard) >= 0)

    return a
}

type Unit struct {
    instanceID InstanceID
    attack int
    defense int
    blocks map[InstanceID]bool
    blockedBy map[InstanceID]bool

    abilities Abilities
    isKing bool
}

type Army struct {
    units map[InstanceID]Unit
}

func NewArmy(p PlayerInfo, createKing bool) *Army {
    a := Army {units: make(map[InstanceID]Unit, maxBoardCards * 2 + 1)}
    if createKing {
        king := Unit {  instanceID: math.MaxInt32,
                        attack: 0,
                        defense: p.health,
                        blocks: make(map[InstanceID]bool, 0),
                        blockedBy: make(map[InstanceID]bool, maxBoardCards),
                        abilities: NewAbilities(""),
                        isKing: true}
        a.units[king.instanceID] = king
    }
    return &a
}

func (army *Army) Frontier() []Unit {
    frontier := make([]Unit, 0, len(army.units))
    for _, u := range army.units {
        if len(u.blockedBy) > 0 {
            continue
        }
        frontier = append(frontier, u)
    }
    if logStepAttack { log("Requested frontier contains %d units", len(frontier)) }
    return frontier
}

func (army *Army) AddCardUnit(card Card, createWardUnit bool) {
    u := Unit {instanceID: InstanceID(card.instanceID),
               attack: card.attack,
               defense: card.defense,
               blocks: make(map[InstanceID]bool, 0),
               blockedBy: make(map[InstanceID]bool, 0),
               abilities: NewAbilities(card.abilities),
               isKing: false}

    // TODO: slightly incorrect as also modifies blocks warded creatures, but should be fine
    for id, other := range army.units {
        if u.abilities[abilityGuard] && !other.abilities[abilityGuard] {
            u.blocks[other.instanceID] = true
            army.units[id].blockedBy[u.instanceID] = true
        } else if !u.abilities[abilityGuard] && other.abilities[abilityGuard] {
            u.blockedBy[other.instanceID] = true
            army.units[id].blocks[other.instanceID] = true
        }
    }

    if createWardUnit && card.isWard {
        ward := u
        ward.instanceID = -u.instanceID
        ward.defense = 1
        ward.blocks[u.instanceID] = true
        ward.blockedBy = u.blockedBy
        u.blockedBy = make(map[InstanceID]bool, 1)
        u.blockedBy[ward.instanceID] = true
        for id, _ := range ward.blockedBy {
            delete(army.units[id].blocks, u.instanceID)
            army.units[id].blocks[ward.instanceID] = true
        }
        army.units[ward.instanceID] = ward
    }

    army.units[u.instanceID] = u
}

func (army *Army) RemoveUnit(id InstanceID) {
    unit := army.units[id]
    for blocksId, _ := range unit.blocks {
        for blockedById, _ := range unit.blockedBy {
            army.units[blockedById].blocks[blocksId] = true
        }

        delete(army.units[blocksId].blockedBy, id)
    }

    for blockedById, _ := range unit.blockedBy {
        delete(army.units[blockedById].blocks, id)
    }
    delete(army.units, id)
}




type AttackKey struct {
    attacker, defender InstanceID
}

func NewAttackKey(a, b InstanceID) AttackKey {
    return AttackKey{attacker: a,
                     defender: b}
}


func (k AttackKey) Cmd() string {
    a, b := k.attacker, k.defender
    if b < 0 {
        // in case of Ward
        b = -b
    } else if b == math.MaxInt32 { // king
        b = -1
    }
    return cmdAttack(int(a), int(b))
}


type AttackResult struct {
    willKill bool
    willDie bool
    willAttackKing bool
    efficiency float64
}

func calcAttackResult(u1, u2 Unit) AttackResult {
    willKill := false
    if u1.attack >= u2.defense || u1.abilities[abilityLethal] {
        willKill = true
    }

    willDie := false
    if u1.defense <= u2.attack || u2.abilities[abilityLethal] {
        willDie = true
    }
    if u1.abilities[abilityWard] {
        willDie = false
    }

    willAttackKing := false
    efficiency := 1 / (1 + float64(u1.attack - u2.defense))
    if u2.isKing {
        willAttackKing = true
        efficiency = float64(u1.attack)
    }
    return AttackResult {willKill: willKill,
                         willDie: willDie,
                         willAttackKing: willAttackKing,
                         efficiency: efficiency}
}

type AttackResults map[AttackKey]AttackResult


func calcAttacksFrontierGreedy(my, enemy Army) []AttackKey {
    if logStepAttack { log("New greedy frontier iteration. My units: %d; enemy units: %d", len(my.units), len(enemy.units)) }
    results := make(AttackResults, 0)
    for _, e := range enemy.Frontier() {
        for _, u := range my.units {
            results[NewAttackKey(u.instanceID, e.instanceID)] = calcAttackResult(u, e)
        }
    }
    bestKey := NewAttackKey(-1, -1)
    bestResult := AttackResult{willKill: false,
                               willDie: true,
                               willAttackKing: false,
                               efficiency: -math.MaxFloat64}

    for key, result := range results {
        isBetterResult := false
        /*
        if result.willAttackKing && result.efficiency > bestResult.efficiency {
            isBetterResult = true
        } else
        */
        if result.willKill == bestResult.willKill &&
                  result.willDie == bestResult.willDie &&
                  result.efficiency > bestResult.efficiency {
            isBetterResult = true
        } else if result.willKill == true && bestResult.willKill == false {
            isBetterResult = true
        } else if result.willKill == bestResult.willKill &&
                  bestResult.willDie == true &&
                  result.willDie == false {
            isBetterResult = true
        }

        if isBetterResult {
            if logStepAttack { log("Attack %+v (%+v) is better than previous %+v (%+v)", result, key, bestResult, bestKey) }
            bestResult = result
            bestKey = key
        }
    }

    if logStepAttack { log("Best attack: %d -> %d", bestKey.attacker, bestKey.defender) }

    attacks := make([]AttackKey, 0, len(my.units))
    if bestKey.attacker == -1 {
        // ending calculations
        if logStepAttack { log("No best attack - no more attacks") }
        return attacks
    }

    attacks = append(attacks, bestKey)

    if bestResult.willKill {
        enemy.RemoveUnit(bestKey.defender)
    } else {
        unit := enemy.units[bestKey.defender]
        unit.defense -= my.units[bestKey.attacker].attack
        enemy.units[bestKey.defender] = unit
    }
    my.RemoveUnit(bestKey.attacker) // no second attack - removing from army

    attacks = append(attacks, calcAttacksFrontierGreedy(my, enemy)...)
    return attacks
}




func main() {
    turnIx := 0
    for {
        players := readPlayers()

        var opponentHand int
        fmt.Scan(&opponentHand)
        fmt.Fprintf(os.Stderr, "Read opponent hand: %d\n", opponentHand)

        cards := readCards()

        if turnIx < maxDraftTurns {
            draftTurn(cards.myHand)
        } else {
            battleTurn(*players, *cards)
        }

        turnIx++

    }
}
