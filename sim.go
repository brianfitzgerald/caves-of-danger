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
	HasEscaped  bool
}

var (
	cooperationProbability = 30
)

/*
SimulateRound rules

each round:
if player can escape, then do a random chance of whether to escape or not (do a percent chance later based on items they know other players have)
on a players turn draw a card
if it is a monster see if people want to fight it
do a reputation sim later on, based on who attacked who (keep it pretty simple)
if you kill a monster put it in your hand
*/
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
		p.HasEscaped = false
		players = append(players, p)
	}
	partyAlive := true
	turnsLasted := 0

	for partyAlive {
		turnsLasted++

		for i, player := range players {
			if player.HasEscaped == true {
				playersEscapedCount := 0
				for _, p := range players {
					if p.HasEscaped {
						playersEscapedCount++
					}
				}
				if playersEscapedCount >= len(players) {
					partyAlive = false
				}
				continue
			}
			if len(player.CardsInHand) > 0 {
				// use card
			}
			if playerWillEscape(player) {
				players[i].HasEscaped = true
				println("player escaped!")
				continue
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
				} else {
					// put killed monster in players hand
					players[i].CardsInHand = append(players[i].CardsInHand, drawnCard)
				}
			} else if drawnCard.Type == Item || drawnCard.Type == Escape {
				players[i].CardsInHand = append(players[i].CardsInHand, drawnCard)
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
		if player.HasEscaped {
			continue
		}
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

// if a player has an escape, then do a random roll; use reputation and other factors later
// TODO: record history of events that happened to a player
func playerWillEscape(player SimPlayer) bool {
	couldEscape := false
	for _, card := range player.CardsInHand {
		if card.Type == Escape {
			if CanEscape(player, card) {
				couldEscape = true
			}
		}
	}
	if !couldEscape {
		return false
	}
	// TODO: replace this with some sort of averaging system
	return rand.Intn(10) > 5
}

func CanEscape(player SimPlayer, escapeCard Card) bool {
	switch escapeCard.EscapeType {
	case MonstersKilled:
		monstersKilled := 0
		for _, c := range player.CardsInHand {
			if c.Type == Monster {
				monstersKilled++
			}
		}
		if monstersKilled >= escapeCard.EscapeConditionValue {
			return true
		}
		return false
	case CardsInHand:
		if len(player.CardsInHand) >= escapeCard.EscapeConditionValue {
			return true
		}
		return false
	case HandWorthGold:
		handGoldValue := 0
		for _, c := range player.CardsInHand {
			handGoldValue += c.GoldValue
		}
		if handGoldValue >= escapeCard.EscapeConditionValue {
			return true
		}
		return false
	default:
		return false
	}
}
