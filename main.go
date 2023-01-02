package main

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/NHAS/genogram/graph"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

func main() {
	app := app.New()
	w := app.NewWindow("Graph Demo")

	w.SetMaster()

	g := graph.NewGraph()

	newPerson := fyne.NewMenuItem("New", func() {
		if g != nil {
			log.Println("added new node")

			id := fmt.Sprintf("%d:random", rand.Int63())
			n := graph.NewGraphNode(g, id, widget.NewLabel(id))

			n.Move(g.LastRightClickPosition)

			g.Refresh()
		}
	})

	addParentsQuickAction := fyne.NewMenuItem("Add Parents (m f)", func() {
		if g != nil {
			log.Println("added new parent nodes")

			maleId := fmt.Sprintf("%d:parent", rand.Int63())
			male := graph.NewGraphNode(g, maleId, widget.NewLabel(maleId))

			male.Move(g.LastRightClickPosition)

			femaleId := fmt.Sprintf("%d:parent", rand.Int63())
			female := graph.NewGraphNode(g, femaleId, widget.NewLabel(femaleId))

			femalePos := g.LastRightClickPosition
			femalePos.X += male.Size().Width + 270
			female.Move(femalePos)

			graph.NewGraphEdge(g, fmt.Sprintf("%s->%s", maleId, femaleId), graph.MarriedRel, male, female)
			g.Refresh()
		}
	})

	addParentsMenu := fyne.NewMenuItem("Add Parents (menu)", func() {
		if g != nil {

		}
	})

	g.Menu = fyne.NewMenu("", newPerson, addParentsQuickAction, addParentsMenu, fyne.NewMenuItemSeparator())

	w.SetContent(g)

	w.ShowAndRun()
}
