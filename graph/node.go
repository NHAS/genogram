package graph

import (
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"math/rand"

	"git.sr.ht/~charles/fynehax/geometry/r2"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	// default inner size
	defaultWidth  float32 = 50
	defaultHeight float32 = 50

	// default padding around the inner object in a node
	defaultPadding float32 = 10
)

type graphNodeRenderer struct {
	node   *GraphNode
	handle *canvas.Line
	box    *canvas.Rectangle
}

// GraphNode represents a node in the graph widget. It contains an inner
// widget, and also draws a border, and a "handle" that can be used to drag it
// around.
type GraphNode struct {
	Id string

	widget.BaseWidget

	Graph *GraphWidget

	// InnerSize stores size that the inner object should have, may not
	// be respected if not large enough for the object.
	InnerSize fyne.Size

	// InnerObject is the canvas object that should be drawn inside of
	// the graph node.
	InnerObject fyne.CanvasObject

	// Padding is the distance between the inner object's drawing area
	// and the box.
	Padding float32

	// BoxStrokeWidth is the stroke width of the box which delineates the
	// node. Defaults to 1.
	BoxStrokeWidth float32

	// BoxFill is the fill color of the node, the inner object will be
	// drawn on top of this. Defaults to the theme.BackgroundColor().
	BoxFillColor color.RGBA

	// BoxStrokeColor is the stroke color of the node rectangle. Defaults
	// to theme.TextColor().
	BoxStrokeColor color.RGBA

	// HandleColor is the color of node handle.
	HandleColor color.RGBA

	// HandleStrokeWidth is the stroke width of the node handle, defaults
	// to 3.
	HandleStroke float32

	Menu *fyne.Menu

	Children map[string]*GraphEdge
	Parents  map[string]*GraphEdge

	Undirected map[string]*GraphEdge
}

type SerialisedNode struct {
	Id string

	Size     fyne.Size
	Position fyne.Position

	InnerSize      fyne.Size
	Padding        float32
	BoxStrokeWidth float32
	BoxFillColor   color.RGBA
	BoxStrokeColor color.RGBA
	HandleColor    color.RGBA
	HandleStroke   float32

	Children   []string
	Parents    []string
	Undirected []string
}

func (r *GraphNode) MarshalJSON() ([]byte, error) {
	out := SerialisedNode{
		Id:             r.Id,
		Size:           r.BaseWidget.Size(),
		Position:       r.BaseWidget.Position(),
		InnerSize:      r.InnerSize,
		Padding:        r.Padding,
		BoxStrokeWidth: r.BoxStrokeWidth,
		BoxFillColor:   r.BoxFillColor,
		BoxStrokeColor: r.BoxStrokeColor,
		HandleColor:    r.HandleColor,
		HandleStroke:   r.HandleStroke,
	}

	for child := range r.Children {
		out.Children = append(out.Children, child)
	}

	for edgeName := range r.Undirected {
		out.Undirected = append(out.Undirected, edgeName)
	}

	for parent := range r.Parents {
		out.Parents = append(out.Parents, parent)
	}

	return json.Marshal(out)
}

func (r *GraphNode) MouseUp(e *desktop.MouseEvent) {
	switch e.Button {
	case desktop.MouseButtonPrimary:
		r.Graph.CompleteLinking(r)
	}
}

func (r *GraphNode) MouseDown(e *desktop.MouseEvent) {
	if e.Button == desktop.MouseButtonSecondary && r.Menu != nil {
		widget.ShowPopUpMenuAtPosition(r.Menu, fyne.CurrentApp().Driver().CanvasForObject(r), e.AbsolutePosition)
	}
}

func (r *graphNodeRenderer) MinSize() fyne.Size {
	// space for the inner widget, plus padding on all sides.
	inner := r.node.effectiveInnerSize()
	return fyne.Size{
		Width:  inner.Width + 2*float32(r.node.Padding),
		Height: inner.Height + 2*float32(r.node.Padding),
	}
}

func (r *graphNodeRenderer) Layout(size fyne.Size) {
	r.node.Resize(r.MinSize())

	r.node.InnerObject.Move(r.node.innerPos())
	r.node.InnerObject.Resize(r.node.effectiveInnerSize())

	r.box.Resize(r.MinSize())

	canvas.Refresh(r.node.InnerObject)
}

func (r *graphNodeRenderer) ApplyTheme(size fyne.Size) {
}

func (r *graphNodeRenderer) Refresh() {
	// move and size the inner object appropriately
	r.node.InnerObject.Move(r.node.innerPos())
	r.node.InnerObject.Resize(r.node.effectiveInnerSize())

	// move the box and update it's colors
	r.box.StrokeWidth = r.node.BoxStrokeWidth
	r.box.FillColor = r.node.BoxFillColor
	r.box.StrokeColor = r.node.BoxStrokeColor
	r.box.Resize(r.MinSize())

	// calculate the handle positions
	r.handle.Position1 = fyne.Position{
		X: float32(r.node.Padding),
		Y: float32(r.node.Padding) / 2,
	}

	r.handle.Position2 = fyne.Position{
		X: r.node.effectiveInnerSize().Width + float32(r.node.Padding),
		Y: float32(r.node.Padding) / 2,
	}

	r.handle.StrokeWidth = r.node.HandleStroke
	r.handle.StrokeColor = r.node.HandleColor

	for _, e := range r.node.Graph.GetEdges(r.node) {
		e.Refresh()
	}

	canvas.Refresh(r.box)
	canvas.Refresh(r.handle)
	canvas.Refresh(r.node.InnerObject)
}

func (r *graphNodeRenderer) BackgroundColor() color.Color {
	return theme.BackgroundColor()
}

func (r *graphNodeRenderer) Destroy() {
}

func (r *graphNodeRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{
		r.box,
		r.handle,
		r.node.InnerObject,
	}
}

func (n *GraphNode) CreateRenderer() fyne.WidgetRenderer {
	r := graphNodeRenderer{
		node:   n,
		handle: canvas.NewLine(n.HandleColor),
		box:    canvas.NewRectangle(n.BoxStrokeColor),
	}

	r.handle.StrokeWidth = n.HandleStroke
	r.box.StrokeWidth = n.BoxStrokeWidth
	r.box.FillColor = n.BoxFillColor

	(&r).Refresh()

	return &r
}

func NewGraphNode(graph *GraphWidget, id string, obj fyne.CanvasObject) *GraphNode {
	w := &GraphNode{
		Id:             id,
		Graph:          graph,
		InnerSize:      fyne.Size{Width: defaultWidth, Height: defaultHeight},
		InnerObject:    obj,
		Padding:        defaultPadding,
		BoxStrokeWidth: 1,

		HandleStroke: 3,
		Children:     make(map[string]*GraphEdge),
		Parents:      make(map[string]*GraphEdge),
		Undirected:   make(map[string]*GraphEdge),
	}

	r, g, b, a := theme.BackgroundColor().RGBA()
	w.BoxFillColor = color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}

	r, g, b, a = theme.ForegroundColor().RGBA()
	w.BoxStrokeColor = color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}

	r, g, b, a = theme.ForegroundColor().RGBA()
	w.HandleColor = color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}

	addChild := fyne.NewMenuItem("Create Child", func() {
		if w != nil {

			id := fmt.Sprintf("%d:random", rand.Int63())

			newNode := NewGraphNode(graph, id, widget.NewLabel(id))
			graph.Nodes[id] = newNode

			log.Println("New child node: ", id)
			newPos := w.Position()

			newPos.Y += w.Size().Height + 50
			newNode.Move(newPos)

			NewGraphEdge(graph, fmt.Sprintf("%s->%s", w.Id, newNode.Id), ChildRel, w, newNode)

			graph.Refresh()
		}
	})

	addLink := fyne.NewMenuItem("Link Child", func() {
		if w != nil {
			graph.StartLinking(w, ChildRel)
			graph.Refresh()
		}
	})

	addHostileLink := fyne.NewMenuItem("Link Hostile", func() {
		if w != nil {
			graph.StartLinking(w, HostileRel)
			graph.Refresh()
		}
	})

	deleteNode := fyne.NewMenuItem("Remove (single)", func() {
		if w != nil {
			graph.DeleteNode(w)
			graph.Refresh()
		}
	})

	deleteChildren := fyne.NewMenuItem("Remove (children)", func() {
		if w != nil {
			graph.DeleteAllChildren(w)
			graph.Refresh()
		}
	})

	deleteAll := fyne.NewMenuItem("Remove (person + all children)", func() {
		if w != nil {

			graph.DeleteAllChildren(w)
			graph.DeleteNode(w)
			graph.Refresh()
		}
	})

	w.Menu = fyne.NewMenu("", addChild, addHostileLink, addLink, fyne.NewMenuItemSeparator(), deleteNode, deleteChildren, deleteAll)

	w.ExtendBaseWidget(w)

	return w
}

func (n *GraphNode) innerPos() fyne.Position {
	return fyne.Position{
		X: n.Padding,
		Y: n.Padding,
	}
}

func (n *GraphNode) effectiveInnerSize() fyne.Size {
	return n.InnerSize.Max(n.InnerObject.MinSize())
}

func (n *GraphNode) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}

func (n *GraphNode) DragEnd() {
	n.Refresh()
}

func (n *GraphNode) Dragged(event *fyne.DragEvent) {
	delta := fyne.Position{X: event.Dragged.DX, Y: event.Dragged.DY}
	n.Displace(delta)
	n.Refresh()
}

func (n *GraphNode) MouseIn(event *desktop.MouseEvent) {

	r, g, b, a := theme.FocusColor().RGBA()
	n.HandleColor = color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}

	n.Refresh()
}

func (n *GraphNode) MouseOut() {

	// For marshalling
	r, g, b, a := theme.ForegroundColor().RGBA()
	n.HandleColor = color.RGBA{uint8(r), uint8(g), uint8(b), uint8(a)}

	n.Refresh()
}

func (n *GraphNode) MouseMoved(event *desktop.MouseEvent) {
}

func (n *GraphNode) Displace(delta fyne.Position) {
	n.Move(n.Position().Add(delta))
}

func (n *GraphNode) R2Position() r2.Vec2 {
	return r2.V2(float64(n.Position().X), float64(n.Position().Y))
}

func (n *GraphNode) R2Box() r2.Box {
	inner := n.effectiveInnerSize()
	s := r2.V2(
		float64(inner.Width+2*n.Padding),
		float64(inner.Height+2*n.Padding),
	)

	return r2.MakeBox(n.R2Position(), s)
}

func (n *GraphNode) R2Center() r2.Vec2 {
	return n.R2Box().Center()
}

func (n *GraphNode) Center() fyne.Position {
	return fyne.Position{float32(n.R2Center().X), float32(n.R2Center().Y)}
}
