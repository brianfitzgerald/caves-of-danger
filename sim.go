package main

// generate deck
// each player draws and if they have an item they use it next round
// what is the group probability of getting through a floor at various levels of cooperation?
// model cooperation probability

import (
	"fmt"
	"math/rand"
)

type SimPlayer struct {
	cardsInHand []Card
	name        string
}

var (
	cooperationProbability = 30
)

func SimulateRound() {
	deck := generateDeck()
	rand.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
	players := []SimPlayer{}
	for i := 0; i < numPlayers; i++ {
		p := SimPlayer{name: fmt.Sprintf("Player %d", i)}
		players = append(players, p)
	}
	partyAlive := true
	turnsLasted := 0
	for partyAlive {
		turnsLasted++
		for activePlayerIndex := 0; activePlayerIndex < len(players); activePlayerIndex++ {
			if len(players[activePlayerIndex].cardsInHand) > 0 {
				// use card
			}
			drawnCard := deck[0]
			deck = deck[1:len(deck)]
			fmt.Printf("Card drawn: %+v \n", drawnCard)
			if drawnCard.Type == Monster {
				if !fightMonster(players, drawnCard) {
					partyAlive = false
				}
			} else if drawnCard.Type == Item || drawnCard.Type == Escape {
				players[activePlayerIndex].cardsInHand = append(players[activePlayerIndex].cardsInHand, drawnCard)
			}
		}
	}
	fmt.Printf("Party lasted for %d turns", turnsLasted)
}

func fightMonster(players []SimPlayer, monster Card) bool {
	nonUselessCardsAmount := 0
	allCardsAmount := 0
	goldAmount := 0
	for _, player := range players {
		if rand.Intn(100) <= cooperationProbability {
			for _, card := range player.cardsInHand {
				if card.Type == Item {
					allCardsAmount++
					if !card.ItemIsUseless {
						nonUselessCardsAmount++
					}
					goldAmount += card.GoldValue
				}
			}
		}
	}
	switch monster.MonsterCombatType {
	case SacrificeAnyItem:
		if allCardsAmount > monster.MonsterCombatValue {
			return true
		}
	case SacrificeGoldAmount:
		if goldAmount > monster.MonsterCombatValue {
			return true
		}
	case SacrificeNonUselessItem:
		if nonUselessCardsAmount > monster.MonsterCombatValue {
			return true
		}
	}
	return false
}
