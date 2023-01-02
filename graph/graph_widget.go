package graph

import (
	"image/color"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type graphRenderer struct {
	graph *GraphWidget
}

type GraphWidget struct {
	widget.BaseWidget

	Offset fyne.Position

	// DesiredSize specifies the size which the graph widget should take
	// up, defaults to 800 x 600
	DesiredSize fyne.Size

	Nodes map[string]*GraphNode
	Edges map[string]*GraphEdge

	Menu *fyne.Menu

	LastRightClickPosition fyne.Position
}

func (r *GraphWidget) MouseUp(e *desktop.MouseEvent) {

}

func (r *GraphWidget) MouseDown(e *desktop.MouseEvent) {
	if e.Button == desktop.MouseButtonSecondary && r.Menu != nil {
		log.Println("right click")

		widget.ShowPopUpMenuAtPosition(r.Menu, fyne.CurrentApp().Driver().CanvasForObject(r), e.AbsolutePosition)
	}

	r.LastRightClickPosition = e.Position
}

func (r *graphRenderer) MinSize() fyne.Size {
	return r.graph.DesiredSize
}

func (r *graphRenderer) Layout(size fyne.Size) {
}

func (r *graphRenderer) ApplyTheme(size fyne.Size) {
}

func (r *graphRenderer) Refresh() {
	for _, e := range r.graph.Edges {
		e.Refresh()
	}
	for _, n := range r.graph.Nodes {
		n.Refresh()
	}
}

func (r *graphRenderer) BackgroundColor() color.Color {
	return theme.BackgroundColor()
}

func (r *graphRenderer) Destroy() {
}

func (r *graphRenderer) Objects() []fyne.CanvasObject {
	obj := make([]fyne.CanvasObject, len(r.graph.Edges)+len(r.graph.Nodes))
	for _, n := range r.graph.Nodes {
		obj = append(obj, n)
	}
	for _, e := range r.graph.Edges {
		obj = append(obj, e)
	}

	return obj
}

func (g *GraphWidget) CreateRenderer() fyne.WidgetRenderer {
	r := graphRenderer{
		graph: g,
	}

	return &r
}

func (g *GraphWidget) Cursor() desktop.Cursor {
	return desktop.DefaultCursor
}

func (g *GraphWidget) DragEnd() {
	g.Refresh()
}

func (g *GraphWidget) Dragged(event *fyne.DragEvent) {
	delta := fyne.Position{X: event.Dragged.DX, Y: event.Dragged.DY}
	for _, n := range g.Nodes {
		n.Displace(delta)
	}
	g.Refresh()
}

func (g *GraphWidget) MouseIn(event *desktop.MouseEvent) {
}

func (g *GraphWidget) MouseOut() {
}

func (g *GraphWidget) MouseMoved(event *desktop.MouseEvent) {
}

func NewGraph() *GraphWidget {
	g := &GraphWidget{
		DesiredSize: fyne.Size{Width: 800, Height: 600},
		Offset:      fyne.Position{0, 0},
		Nodes:       map[string]*GraphNode{},
		Edges:       map[string]*GraphEdge{},
	}

	g.ExtendBaseWidget(g)

	return g
}

func (g *GraphWidget) GetEdges(n *GraphNode) []*GraphEdge {
	edges := []*GraphEdge{}

	for _, e := range g.Edges {
		if e.Origin == n {
			edges = append(edges, e)
		} else if e.Target == n {
			edges = append(edges, e)
		}
	}

	return edges
}

// Deletes node and all assocaited edges
func (g *GraphWidget) DeleteNode(n *GraphNode) {

	for id, child := range n.Children {
		delete(child.Target.Parents, id)
		delete(g.Edges, id)
	}

	for id, parent := range n.Parents {
		delete(parent.Target.Children, id)
		delete(g.Edges, id)
	}

	for id, partner := range n.Undirected {
		delete(partner.Target.Children, id)
		delete(partner.Target.Parents, id)
		delete(g.Edges, id)

	}

	delete(g.Nodes, n.Id)
}

func (g *GraphWidget) DeleteAllChildren(n *GraphNode) {

	for id, child := range n.Children {

		g.DeleteAllChildren(child.Target)

		delete(g.Edges, id)
		g.DeleteNode(child.Target)
	}

}
