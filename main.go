package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/golang/freetype/truetype"
	wordwrap "github.com/mitchellh/go-wordwrap"
	"golang.org/x/image/font"

	"github.com/golang/freetype"
	colorful "github.com/lucasb-eyer/go-colorful"
)

var (
	items = []Card{
		Card{
			Name:        "Amulet of Sight",
			Type:        Item,
			GoldValue:   3,
			Description: "Look 3 cards into the deck.",
		},
		Card{
			Name:        "Double or Nothing",
			Type:        Item,
			GoldValue:   3,
			Description: "Double the value of this card if you escaped this round.",
		},
		Card{
			Name:        "Midas Touch",
			Type:        Item,
			GoldValue:   5,
			Description: "Turn a monster to Gold.",
		},
		Card{
			Name:        "Dreamwork",
			Type:        Item,
			GoldValue:   3,
			Description: "Double the value of this card if you stayed until the end of this round.",
		},
		Card{
			Name:        "Left Bamboozle",
			Type:        Item,
			GoldValue:   5,
			Description: "Skip your turn, the player on your left keeps this card and has to draw 2 on their turn.",
		},
		Card{
			Name:        "Right Bamboozle",
			Type:        Item,
			GoldValue:   4,
			Description: "Skip your turn, the player on your right keeps this card and has to draw 2 on their turn.",
		},
		Card{
			Name:        "Bliss",
			Type:        Item,
			GoldValue:   6,
			Description: "Ignore the effects of a card you drew.",
		},
		Card{
			Name:        "Automated Buck Passer",
			Type:        Item,
			GoldValue:   4,
			Description: "Skip your turn.",
		},
		Card{
			Name:        "Next Door Over",
			Type:        Item,
			GoldValue:   5,
			Description: "Take the top card on the deck and put it on the bottom.",
		},
		Card{
			Name:         "Zero day",
			Type:         Item,
			GoldValue:    5,
			NumberInDeck: 2,
			Description:  "Steal a card from another player. You get to choose the card.",
		},
		Card{
			Name:         "Walk Softly",
			Type:         Item,
			GoldValue:    5,
			NumberInDeck: 2,
			Description:  "Steal a card from another player. You get to choose the card.",
		},
	}
	monsterNames = []string{
		"Grid Bug",
		"Hell Beast",
		"Worm King",
		"Lesser Child",
		"Dark Lord",
		"Chosen King",
		"Elder Prince",
		"Burrito Dispenser",
		"Libertarian",
		"Comeback Lizard",
		"Night Priest",
		"Big Snake",
		"Cursed Rod",
		"Glass Dragon",
		"Bush Baby",
		"Oaf",
		"Spoon Lord",
	}
	uselessItems = []string{
		"Golden Horn",
		"Decently Used Coat",
		"Working Radio",
		"Spotless Hubcaps",
		"Tylenol",
		"Radio Goggles",
		"Cinnamon Stick",
		"Incense Flume",
		"Confusing Graph",
		"Deodorant Jar",
	}
	escapes = []EscapeDesc{
		EscapeDesc{
			Name:                 "Big Rope",
			Condition:            "Wait until you have killed 3 Monsters this round.",
			EscapeType:           MonstersKilled,
			EscapeConditionValue: 3,
		},
		EscapeDesc{
			Name:                 "Dinner Time",
			Condition:            "Wait until you have 10 cards in your hand.",
			EscapeType:           CardsInHand,
			EscapeConditionValue: 10,
		},
		EscapeDesc{
			Name:       "Dip",
			Condition:  "Wait until your turn is skipped.",
			EscapeType: TurnSkipped,
		},
		EscapeDesc{
			Name:       "Skateboard Away",
			Condition:  "Wait until someone steals a card from you.",
			EscapeType: CardStolen,
		},
		EscapeDesc{
			Name:                 "My Mom's here to pick me up",
			Condition:            "Wait until your hand is worth 15 Gold.",
			EscapeType:           HandWorthGold,
			EscapeConditionValue: 15,
		},
	}
)

type EscapeDesc struct {
	Name                 string
	Condition            string
	EscapeConditionValue int
	EscapeType           EscapeType
}

type Card struct {
	Name                 string
	Type                 CardType
	GoldValue            int
	Description          string
	NumberInDeck         int
	MonsterCombatType    MonsterCombatType
	MonsterCombatValue   int
	EscapeType           EscapeType
	ItemIsUseless        bool
	EscapeConditionValue int
}

type EscapeType int

const (
	MonstersKilled EscapeType = iota
	CardsInHand
	TurnSkipped
	CardStolen
	HandWorthGold
)

type CardType int

const (
	Item CardType = iota
	Monster
	Escape
)

type MonsterCombatType int

const (
	SacrificeAnyItem MonsterCombatType = iota
	SacrificeGoldAmount
	SacrificeNonUselessItem
)

func (f CardType) String() string {
	return [...]string{"Item", "Monster", "Escape"}[f]
}

var (
	red    = ParseHexColor("#C55439")
	blue   = ParseHexColor("#189BC1")
	purple = ParseHexColor("#584976")
)

func ParseHexColor(s string) (c color.Color) {
	c, err := colorful.Hex(s)
	if err != nil {
		fmt.Println(err)
	}
	return
}

var (
	robotoRegularSrc = "./resources/roboto/Roboto-Regular.ttf"
	robotoBoldSrc    = "./resources/roboto/Roboto-Bold.ttf"
	hinting          = flag.String("hinting", "none", "none | full")
	headerSize       = float64(24)
	bodyTextSize     = float64(24)
	spacing          = flag.Float64("spacing", 1.5, "line spacing (e.g. 2 means double spaced)")
	wonb             = flag.Bool("whiteonblack", false, "white text on a black background")
	dpi              = flag.Float64("dpi", 22, "screen resolution in Dots Per Inch")
	scale            = 2
	cardsPerRow      = 10
	printType        = TabletopSim
	rowsPerPage      = 6
)

type PrintDocumentType int

const (
	TabletopSim PrintDocumentType = iota
	A4
)

func main() {

	// example commands:
	// go run sim.go main.go gen
	// go run sim.go main.go print A4

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "gen":
			testDeck()
			break
		case "sim":
			roundsLastedResults := []int{}
			for index := 0; index < 100000; index++ {
				roundsLasted := SimulateRound()
				roundsLastedResults = append(roundsLastedResults, roundsLasted)
			}
			fmt.Printf("avg round length: %d\n", average(roundsLastedResults))

			break
		case "print":
			printDeck()
			break
		}
	}

}

func average(xs []int) float64 {
	total := 0
	for _, v := range xs {
		total += v
	}
	return float64(total) / float64(len(xs))
}

func testDeck() {
	cards := generateDeck()
	r := rand.New(rand.NewSource(time.Now().Unix()))
	shuffled := []Card{}
	for _, i := range r.Perm(len(cards)) {
		val := cards[i]
		shuffled = append(shuffled, val)
	}

	for _, card := range shuffled {
		fmt.Println(card.Type.String(), card.Name, card.Description, card.GoldValue)
	}
}

func printDeck() {
	if len(os.Args) > 2 && os.Args[2] == "A4" {
		printType = A4
	}

	templateWidth := scale * 640
	templateHeight := scale * 640

	if printType == A4 {
		templateWidth = 2480 * 1
		templateHeight = 3508 * 1
		scale = 8
	}

	cards := generateDeck()

	cardsPerPage := 40
	numPages := 2

	for i := 0; i < numPages; i++ {
		templateImage := image.NewRGBA(image.Rect(0, 0, templateWidth, templateHeight))
		end := i*cardsPerPage + cardsPerPage
		if i == numPages-1 {
			end = len(cards)
		}
		for i, card := range cards[i*cardsPerPage : end] {
			err := drawCard(card, templateImage, i)
			if err != nil {
				fmt.Println(err)
			}
		}
		file, err := os.Create(fmt.Sprintf("exports/all_cards_page_%d.png", i+1))
		if err != nil {
			fmt.Println(err)
		}
		png.Encode(file, templateImage)

	}

}

var (
	escapesPerRound      = 3
	monstersPerRound     = 6
	itemsPerRound        = 4
	uselessItemsPerRound = 5
	numRounds            = 3
	startingCards        = 3
	numPlayers           = 3
)

func generateDeck() []Card {

	rand.Seed(time.Now().UnixNano())
	cards := []Card{}

	usedItems := map[string]bool{}
	for _, c := range items {
		usedItems[c.Name] = false
	}
	for i := 0; i < itemsPerRound*numRounds+(startingCards*numPlayers); i++ {
		card := items[rand.Intn(len(items))]

		validCards := len(items)
		for _, c := range items {
			if usedItems[c.Name] == true {
				validCards--
			}
		}

		for usedItems[card.Name] == true && validCards > 0 {
			card = items[rand.Intn(len(items))]
		}
		usedItems[card.Name] = true
		num := card.NumberInDeck
		if card.NumberInDeck == 0 {
			num = 1
		}
		for i := 0; i < num; i++ {
			cards = append(cards, card)
		}
	}
	for i := 0; i < monstersPerRound*numRounds; i++ {
		name := monsterNames[rand.Intn(len(monsterNames))]
		desc := "Sacrifice any 2 Items to defeat."
		val := rand.Intn(7) + 2
		combatType := SacrificeAnyItem
		combatValue := 2
		if rand.Intn(5) >= 4 {
			combatType = SacrificeNonUselessItem
			desc = "Sacrifice any non-Useless Item to defeat."
			val = 6
			combatValue = 0
		}
		if rand.Intn(5) >= 4 {
			combatType = SacrificeGoldAmount
			desc = "Sacrifice 10 Gold worth of Items to defeat."
			val = 8
			combatValue = 10
		}
		if rand.Intn(5) >= 4 {
			desc = "Sacrifice any 3 Items to defeat."
			val = 10
			combatValue = 3
			combatType = SacrificeAnyItem
		}
		if rand.Intn(5) >= 4 {
			desc = "Sacrifice a Useless Item to defeat."
			val = 6
			combatType = SacrificeNonUselessItem
		}
		if rand.Intn(10) >= 9 {
			name = "Rogue Genie"
			desc = "Sacrifice 10 Gold worth of items to defeat."
			val = 12
			combatValue = 10
			combatType = SacrificeGoldAmount
		}
		monster := Card{
			Type:               Monster,
			GoldValue:          val,
			Name:               name,
			Description:        desc,
			MonsterCombatType:  combatType,
			MonsterCombatValue: combatValue,
		}
		cards = append(cards, monster)
	}
	for i := 0; i < escapesPerRound*numRounds; i++ {
		e := escapes[rand.Intn(len(escapes))]
		escape := Card{
			Type:                 Escape,
			EscapeConditionValue: e.EscapeConditionValue,
			EscapeType:           e.EscapeType,
			GoldValue:            i,
			Description:          e.Condition,
			Name:                 e.Name,
		}
		cards = append(cards, escape)
	}

	// generate useless cards
	for i := 0; i < uselessItemsPerRound*numRounds; i++ {
		e := uselessItems[rand.Intn(len(uselessItems))]
		item := Card{
			Type:          Item,
			GoldValue:     i,
			Name:          e,
			ItemIsUseless: true,
			Description:   "This card is useless! But it is worth some coin.",
		}
		cards = append(cards, item)
	}

	return cards

}

func drawCard(card Card, templateImage *image.RGBA, i int) error {

	// draw individual card
	cardFilename := "exports/cards/"
	cardFilename += strings.ReplaceAll(card.Name, " ", "_")
	cardFilename = strings.ToLower(cardFilename)
	cardFilename += ".png"
	cardImage := image.NewRGBA(image.Rect(0, 0, scale*63, scale*91))

	bgColor := blue
	switch card.Type {
	case Monster:
		bgColor = red
	case Item:
		bgColor = blue
	case Escape:
		bgColor = purple
	}

	bodyTextColor := bgColor
	headerTextColor := color.White

	innerBounds := cardImage.Bounds()
	innerBounds = innerBounds.Inset(2)

	// draw outline

	draw.Draw(cardImage, cardImage.Bounds(), &image.Uniform{color.Black}, image.ZP, draw.Src)
	draw.Draw(cardImage, innerBounds, &image.Uniform{color.White}, image.ZP, draw.Src)

	headerRect := image.Rect(0, 0, cardImage.Bounds().Dx(), 240)
	draw.Draw(cardImage, headerRect, &image.Uniform{bgColor}, image.ZP, draw.Src)

	// load fonts

	fontBytes, err := ioutil.ReadFile(robotoBoldSrc)
	if err != nil {
		return err
	}

	robotoBold, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return err
	}

	fontBytes, err = ioutil.ReadFile(robotoRegularSrc)
	if err != nil {
		return err
	}

	robotoRegular, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return err
	}

	// draw text

	leftMargin := 5

	drawText(card.Name, robotoBold, cardImage, leftMargin, 10, 20, headerTextColor)
	drawText(card.Description, robotoRegular, cardImage, leftMargin, 40, 18, bodyTextColor)
	valueString := fmt.Sprintf("%d", card.GoldValue)
	drawText(valueString, robotoRegular, cardImage, leftMargin+40, 25, 18, headerTextColor)
	drawText(card.Type.String(), robotoRegular, cardImage, leftMargin, 25, 18, headerTextColor)

	// draw money icon

	moneyIcon, _ := os.Open("resources/money_icon.png")
	moneyIconImg, _, _ := image.Decode(moneyIcon)
	iconBounds := moneyIconImg.Bounds()
	iconPos := image.Pt(280, 160)
	iconPosRect := image.Rect(iconPos.X, iconPos.Y, iconBounds.Dx()+iconPos.X, iconBounds.Dy()+iconPos.Y)
	draw.Draw(cardImage, iconPosRect, moneyIconImg, image.Pt(0, 0), draw.Over)

	// draw individual card

	file, err := os.Create(cardFilename)
	if err != nil {
		return err
	}
	png.Encode(file, cardImage)

	// draw card to full page template image

	y := (i / cardsPerRow) * 182 * scale / 2
	x := (i % cardsPerRow) * 126 * scale / 2

	r := image.Rect(x, y, x+(126*scale), y+(182*scale))

	draw.Draw(templateImage, r, cardImage, image.ZP, draw.Src)

	return nil
}

func drawText(text string, f *truetype.Font, src *image.RGBA, x, y int, size int, textColor color.Color) {

	y = y * scale
	x = x * scale
	size = size * scale

	fg := &image.Uniform{textColor}
	c := freetype.NewContext()
	c.SetDPI(*dpi)
	c.SetFont(f)
	c.SetFontSize(float64(size))
	c.SetClip(src.Bounds())
	c.SetDst(src)
	c.SetSrc(fg)
	c.SetHinting(font.HintingFull)

	lineWidth := 15

	wrapped := wordwrap.WrapString(text, uint(lineWidth))
	splitText := strings.Split(wrapped, "\n")

	for i, s := range splitText {
		pt := freetype.Pt(x, y+(int(c.PointToFixed(float64(size))>>6)*i))
		c.DrawString(s, pt)
	}

}
