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
			Value:       3,
			Description: "Look 3 cards into the deck.",
		},
		Card{
			Name:        "Double or Nothing",
			Type:        Item,
			Value:       3,
			Description: "Double the value of this card if you escaped this round.",
		},
		Card{
			Name:        "Left Bamboozle",
			Type:        Item,
			Value:       5,
			Description: "Skip your turn, the player on your left keeps this card and has to draw 2 on their turn.",
		},
		Card{
			Name:        "Right Bamboozle",
			Type:        Item,
			Value:       4,
			Description: "Skip your turn, the player on your right keeps this card and has to draw 2 on their turn.",
		},
		Card{
			Name:        "Bliss",
			Type:        Item,
			Value:       6,
			Description: "Ignore the effects of a card you drew.",
		},
		Card{
			Name:        "Automated Buck Passer",
			Type:        Item,
			Value:       4,
			Description: "Skip your turn.",
		},
		Card{
			Name:        "Next Door Over",
			Type:        Item,
			Value:       5,
			Description: "Take the top card on the deck and put it on the bottom.",
		},
		Card{
			Name:         "Next Door Over",
			Type:         Item,
			Value:        5,
			NumberInDeck: 2,
			Description:  "Take the top card on the deck and put it on the bottom.",
		},
		Card{
			Name:         "Swipe",
			Type:         Item,
			Value:        5,
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
		"Deodorant Jar",
	}
	escapes = []EscapeDesc{
		EscapeDesc{
			Name:      "Rope",
			Condition: "Wait until 3 Monsters are killed this round.",
		},
		EscapeDesc{
			Name:      "Dinner Time",
			Condition: "Wait until you have 5 cards in your hand.",
		},
		EscapeDesc{
			Name:      "Dip",
			Condition: "Wait until your turn is skipped.",
		},
		EscapeDesc{
			Name:      "Skateboard Away",
			Condition: "Wait until the deck is shuffled.",
		},
		EscapeDesc{
			Name:      "My Mom's here to pick me up",
			Condition: "Wait until your hand is worth 10 Gold.",
		},
	}
)

type EscapeDesc struct {
	Name      string
	Condition string
}

type Card struct {
	Name         string
	Type         CardType
	Value        int
	Description  string
	NumberInDeck int
}

type CardType int

const (
	Item CardType = iota
	Monster
	Escape
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

	if len(os.Args) > 1 {
		if os.Args[1] == "deck" {
			testDeck()
		}
		if os.Args[1] == "print" {
			printDeck()
		}
	}

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
		fmt.Println(card.Type.String(), card.Name, card.Description, card.Value)
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

	rand.Seed(time.Now().Unix())

	cards := []Card{}

	for i := 0; i < itemsPerRound*numRounds+(startingCards*numPlayers); i++ {
		card := items[rand.Intn(len(items))]
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
		if rand.Intn(5) >= 4 {
			desc = "Sacrifice any non-Useless Item to defeat."
		}
		if rand.Intn(5) >= 4 {
			desc = "Sacrifice 10 Gold worth of Items to defeat."
		}
		if rand.Intn(5) >= 4 {
			desc = "Sacrifice any 3 Items to defeat."
		}
		if rand.Intn(5) >= 4 {
			desc = "Sacrifice any 3 Items to defeat."
		}
		if rand.Intn(5) >= 4 {
			desc = "Sacrifice a Useless Item to defeat."
		}
		monster := Card{
			Type:        Monster,
			Value:       i * 2,
			Name:        name,
			Description: desc,
		}
		cards = append(cards, monster)
	}
	for i := 0; i < escapesPerRound*numRounds; i++ {
		e := escapes[rand.Intn(len(escapes))]
		escape := Card{
			Type:        Escape,
			Value:       i,
			Description: e.Condition,
			Name:        e.Name,
		}
		cards = append(cards, escape)
	}
	for i := 0; i < uselessItemsPerRound*numRounds; i++ {
		e := uselessItems[rand.Intn(len(uselessItems))]
		item := Card{
			Type:        Item,
			Value:       i,
			Name:        e,
			Description: "This card is useless! But it is worth some coin.",
		}
		cards = append(cards, item)
	}

	return cards

}

func drawCard(card Card, templateImage *image.RGBA, i int) error {
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

	textColor := color.Color(color.White)

	if printType == A4 {
		textColor = bgColor
		bgColor = color.White
	}

	innerBounds := cardImage.Bounds()
	innerBounds = innerBounds.Inset(2)

	draw.Draw(cardImage, cardImage.Bounds(), &image.Uniform{color.Black}, image.ZP, draw.Src)
	draw.Draw(cardImage, innerBounds, &image.Uniform{bgColor}, image.ZP, draw.Src)

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

	leftMargin := 5

	drawText(card.Name, robotoBold, cardImage, leftMargin, 10, 20, textColor)
	drawText(card.Description, robotoRegular, cardImage, leftMargin, 40, 18, textColor)
	valueString := fmt.Sprintf("Worth %d Gold", card.Value)
	drawText(valueString, robotoRegular, cardImage, leftMargin, 82, 16, textColor)
	drawText(card.Type.String(), robotoRegular, cardImage, leftMargin, 25, 18, textColor)

	file, err := os.Create(cardFilename)
	if err != nil {
		return err
	}
	png.Encode(file, cardImage)

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
