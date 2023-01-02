package graph

import (
	"encoding/json"
	"fmt"
	"image/color"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
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

	NewLinkNode      *GraphNode
	lineToMouse      *canvas.Line
	relationshiptype string

	MousePosition fyne.Position
}

func (r *GraphWidget) MouseUp(e *desktop.MouseEvent) {

}

func (r *GraphWidget) MouseDown(e *desktop.MouseEvent) {

	switch e.Button {
	case desktop.MouseButtonPrimary:
		r.StopLinking()
	case desktop.MouseButtonSecondary:
		log.Println("right click")
		if r.Menu != nil {
			r.LastRightClickPosition = e.Position

			widget.ShowPopUpMenuAtPosition(r.Menu, fyne.CurrentApp().Driver().CanvasForObject(r), e.AbsolutePosition)
		}
	}

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

	if r.graph.NewLinkNode != nil {
		r.graph.lineToMouse.Refresh()
	}
}

func (r *graphRenderer) BackgroundColor() color.Color {
	return theme.BackgroundColor()
}

func (r *graphRenderer) Destroy() {
}

func (r *graphRenderer) Objects() []fyne.CanvasObject {
	obj := make([]fyne.CanvasObject, len(r.graph.Edges)+len(r.graph.Nodes))

	if r.graph.NewLinkNode != nil {
		obj = append(obj, r.graph.lineToMouse)
	}

	for _, e := range r.graph.Edges {
		obj = append(obj, e)
	}

	for _, n := range r.graph.Nodes {
		obj = append(obj, n)
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

func (g *GraphWidget) StartLinking(parent *GraphNode, reltype string) {
	g.NewLinkNode = parent
	g.lineToMouse = canvas.NewLine(selectColor(reltype))
	g.lineToMouse.Position2 = g.MousePosition
	g.lineToMouse.Position1 = g.NewLinkNode.Center()
	g.relationshiptype = reltype
}

func (g *GraphWidget) CompleteLinking(child *GraphNode) {
	if g.NewLinkNode != nil && child != g.NewLinkNode {
		NewGraphEdge(g, fmt.Sprintf("%s->%s", g.NewLinkNode.Id, child.Id), g.relationshiptype, g.NewLinkNode, child)
	}

	g.StopLinking()
	g.Refresh()
}

func (g *GraphWidget) StopLinking() {
	g.NewLinkNode = nil
	g.lineToMouse = nil
	g.Refresh()
}

func (g *GraphWidget) MouseIn(event *desktop.MouseEvent) {
}

func (g *GraphWidget) MouseOut() {
}

func (g *GraphWidget) MouseMoved(event *desktop.MouseEvent) {

	g.MousePosition = event.Position

	if g.NewLinkNode != nil {
		g.lineToMouse.Position2 = event.Position
		g.Refresh()
	}
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

func (g *GraphWidget) ClearGraph() {
	g.StopLinking()
	g.Edges = make(map[string]*GraphEdge)
	g.Nodes = make(map[string]*GraphNode)
}

func (g *GraphWidget) MarshalJSON() ([]byte, error) {

	var sg struct {
		Nodes map[string]*GraphNode
		Edges map[string]*GraphEdge
	}

	sg.Nodes = g.Nodes
	sg.Edges = g.Edges

	return json.Marshal(sg)
}

func (g *GraphWidget) UnmarshalJSON(b []byte) error {

	g.ClearGraph()

	var sg struct {
		Nodes map[string]SerialisedNode
		Edges map[string]SerialisedEdge
	}

	err := json.Unmarshal(b, &sg)
	if err != nil {
		return err
	}

	realNodes := make(map[string]*GraphNode)
	for nodeId, node := range sg.Nodes {

		newNode := NewGraphNode(g, nodeId, widget.NewLabel(nodeId))

		newNode.InnerSize = node.InnerSize
		newNode.Padding = node.Padding
		newNode.BoxStrokeWidth = node.BoxStrokeWidth
		newNode.BoxFillColor = node.BoxFillColor
		newNode.HandleColor = node.HandleColor
		newNode.HandleStroke = node.HandleStroke

		newNode.Resize(node.Size)
		newNode.Move(node.Position)

		realNodes[nodeId] = newNode
	}

	realEdges := make(map[string]*GraphEdge)
	for edgeId, edge := range sg.Edges {
		newEdge := &GraphEdge{
			Id:        edgeId,
			Type:      edge.Type,
			Directed:  edge.Directed,
			EdgeColor: edge.EdgeColor,
			Width:     edge.Width,
		}

		newEdge.Origin = realNodes[edge.Origin]
		newEdge.Target = realNodes[edge.Target]

		if edge.Directed {
			newEdge.Origin.Children[edgeId] = newEdge
			newEdge.Target.Parents[edgeId] = newEdge

		} else {
			newEdge.Origin.Undirected[edgeId] = newEdge
			newEdge.Target.Undirected[edgeId] = newEdge
		}

		newEdge.ExtendBaseWidget(newEdge)

		realEdges[edgeId] = newEdge
	}

	g.Nodes = realNodes
	g.Edges = realEdges

	g.Refresh()

	return nil
}
