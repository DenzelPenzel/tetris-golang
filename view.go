package main

import (
	"fmt"
	"github.com/gdamore/tcell"
	"strings"
)

func (g *Game) drawMenu() {
	style := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	for y, item := range menu {
		if strings.HasPrefix(item, "Level:") {
			item = fmt.Sprintf(item, 1)
		} else if strings.HasPrefix(item, "Lines:") {
			item = fmt.Sprintf(item, 1)
		} else if strings.HasPrefix(item, "GAME OVER") {
			item = ""
		}
		g.print(gameGridWidth+10, y, item, style)
	}
}

func (g *Game) print(x, y int, msg string, color tcell.Style) {
	for _, m := range msg {
		g.screen.SetContent(x, y, m, nil, color)
		x++
	}
}
