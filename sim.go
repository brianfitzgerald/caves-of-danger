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
	CardsInHand []Card
	name        string
}

var (
	cooperationProbability = 30
)

func SimulateRound() int {
	deck := generateDeck()
	rand.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
	players := []SimPlayer{}
	for i := 0; i < numPlayers; i++ {
		p := SimPlayer{name: fmt.Sprintf("Player %d", i)}
		for i := 0; i < 7; i++ {
			drawnCard := deck[0]
			deck = deck[1:len(deck)]
			p.CardsInHand = append(p.CardsInHand, drawnCard)

		}

		players = append(players, p)
	}
	partyAlive := true
	turnsLasted := 0

	for partyAlive {
		turnsLasted++

		for activePlayerIndex := 0; activePlayerIndex < len(players); activePlayerIndex++ {
			if len(players[activePlayerIndex].CardsInHand) > 0 {
				// use card
			}
			if len(deck) == 0 {
				return turnsLasted
			}
			drawnCard := deck[0]
			deck = deck[1:len(deck)]
			fmt.Printf("Drew %s: %s, %s\n", drawnCard.Type.String(), drawnCard.Name, drawnCard.Description)
			if drawnCard.Type == Monster {
				if !fightMonster(players, drawnCard) {
					fmt.Printf("fighting\n")
					partyAlive = false
				}
			} else if drawnCard.Type == Item || drawnCard.Type == Escape {
				players[activePlayerIndex].CardsInHand = append(players[activePlayerIndex].CardsInHand, drawnCard)
			}
		}
	}
	fmt.Printf("Party lasted for %d turns\n", turnsLasted)
	return turnsLasted
}

func fightMonster(players []SimPlayer, monster Card) bool {
	nonUselessCardsAmount := 0
	allCardsAmount := 0
	goldAmount := 0
	for _, player := range players {
		if rand.Intn(100) <= cooperationProbability {
			for _, card := range player.CardsInHand {
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
