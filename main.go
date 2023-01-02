package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"

	"github.com/NHAS/genogram/graph"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
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

			g.Nodes[id] = n

			n.Move(g.LastRightClickPosition)

			g.Refresh()
		}
	})

	addParentsQuickAction := fyne.NewMenuItem("Add Parents (m f)", func() {
		if g != nil {
			log.Println("added new parent nodes")

			maleId := fmt.Sprintf("%d:parent", rand.Int63())
			male := graph.NewGraphNode(g, maleId, widget.NewLabel(maleId))
			g.Nodes[maleId] = male

			male.Move(g.LastRightClickPosition)

			femaleId := fmt.Sprintf("%d:parent", rand.Int63())
			female := graph.NewGraphNode(g, femaleId, widget.NewLabel(femaleId))
			g.Nodes[femaleId] = female

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

	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.DocumentCreateIcon(), func() {
			log.Println("New document")
			g.ClearGraph()
		}),
		widget.NewToolbarAction(theme.DocumentSaveIcon(), func() {
			log.Println("Saving graph")
			b, err := g.MarshalJSON()
			if err != nil {
				log.Println("error saving graph: ", err)
				return
			}

			err = ioutil.WriteFile("saved.json", b, 0644)
			if err != nil {
				log.Println("failed to write file: ", err)
				return
			}

			log.Println("file saved")
		}),
		widget.NewToolbarAction(theme.FolderOpenIcon(), func() {
			log.Println("Open file")

			b, err := ioutil.ReadFile("saved.json")
			if err != nil {
				log.Println("reading file failed: ", err)
				return
			}

			err = g.UnmarshalJSON(b)
			if err != nil {
				log.Println("error unmarshalling graph: ", err)
				return
			}

			log.Println("file opened")

			log.Printf("graph, nodes %d, edges %d", len(g.Nodes), len(g.Edges))
			log.Printf("Nodes: %+v", g.Nodes)
			log.Printf("Edges: %+v", g.Edges)

		}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.ContentCutIcon(), func() {}),
		widget.NewToolbarAction(theme.ContentCopyIcon(), func() {}),
		widget.NewToolbarAction(theme.ContentPasteIcon(), func() {}),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.HelpIcon(), func() {
			log.Println("Display help")
		}),
	)

	content := container.NewBorder(toolbar, nil, nil, nil, g)

	w.SetContent(content)

	w.ShowAndRun()
}
