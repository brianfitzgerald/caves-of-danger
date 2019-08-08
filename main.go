package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"os"
	"strings"

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
			Value:       2,
			Description: "Look 3 cards into the deck.",
		},
		Card{
			Name:        "Double or Nothing",
			Type:        Item,
			Value:       3,
			Description: "Double value of this card if you escape.",
		},
		Card{
			Name:        "Left Bamboozle",
			Type:        Item,
			Value:       5,
			Description: "The player on your left must draw.",
		},
		Card{
			Name:        "Right Bamboozle",
			Type:        Item,
			Value:       4,
			Description: "The player on your right must draw.",
		},
		Card{
			Name:        "Bliss",
			Type:        Item,
			Value:       6,
			Description: "Ignore the effects of a card you draw.",
		},
	}
	monsterNames = []string{
		"Grid Bug",
		"Hell Beast",
		"Worm King",
		"Lesser Child",
		"Dark Lord",
		"Chosen King",
	}
)

type Card struct {
	Name        string
	Type        CardType
	Value       int
	Description string
}

type CardType int

const (
	Item CardType = iota
	Monster
)

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
)

func main() {
	templateSize := scale * 640
	templateImage := image.NewRGBA(image.Rect(0, 0, templateSize, templateSize))
	for i, card := range items {
		err := drawCard(card, templateImage, i)
		if err != nil {
			fmt.Println(err)
		}
	}
	for i, name := range monsterNames {
		monster := Card{
			Type:  Monster,
			Value: i,
			Name:  name,
		}
		err := drawCard(monster, templateImage, i+len(items))
		if err != nil {
			fmt.Println(err)
		}
	}
	file, err := os.Create("exports/all_cards.png")
	if err != nil {
		fmt.Println(err)
	}
	png.Encode(file, templateImage)

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

	}

	draw.Draw(cardImage, cardImage.Bounds(), &image.Uniform{bgColor}, image.ZP, draw.Src)

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

	drawText(card.Name, robotoBold, cardImage, 10, 15, scale*20)
	drawText(card.Description, robotoRegular, cardImage, 10, 50, scale*18)

	file, err := os.Create(cardFilename)
	if err != nil {
		return err
	}
	png.Encode(file, cardImage)

	y := (i / 10) * 200
	x := (i % 10) * 120

	r := image.Rect(x, y, x+126, y+182)

	draw.Draw(templateImage, r, cardImage, image.ZP, draw.Src)

	return nil
}

func drawText(text string, f *truetype.Font, src *image.RGBA, x, y int, size int) {
	fg := &image.Uniform{color.White}
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