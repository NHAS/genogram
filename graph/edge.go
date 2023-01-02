package graph

import (
	"image/color"
	"log"

	"git.sr.ht/~charles/fynehax/geometry/r2"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	MarriedRel       = "married"
	ChildRel         = "child"
	SeparationRel    = "separated"
	CohabatiationRel = "cohabatiation"
	DistantRel       = "distant"
	FriendRel        = "friend"
	HostileRel       = "hostile"
	AbuseRel         = "abuse"
)

type graphEdgeRenderer struct {
	edge *GraphEdge
	line *canvas.Line

	circle *canvas.Circle
}

type GraphEdge struct {
	widget.BaseWidget

	Graph *GraphWidget

	EdgeColor color.Color

	Width float32

	Origin *GraphNode
	Target *GraphNode

	Type string

	Directed bool
}

func (r *graphEdgeRenderer) MinSize() fyne.Size {
	xdelta := r.edge.Origin.Position().X - r.edge.Target.Position().X
	if xdelta < 0 {
		xdelta *= -1
	}

	ydelta := r.edge.Origin.Position().Y - r.edge.Target.Position().Y
	if ydelta < 0 {
		ydelta *= -1
	}

	return fyne.Size{Width: xdelta, Height: ydelta}
}

func (r *graphEdgeRenderer) Layout(size fyne.Size) {
	r.circle.Resize(fyne.NewSize(10, 10))
}

func (r *graphEdgeRenderer) ApplyTheme(size fyne.Size) {
}

func (r *graphEdgeRenderer) Refresh() {
	l := r.edge.R2Line()
	b1 := r.edge.Origin.R2Box()
	b2 := r.edge.Target.R2Box()

	p1, _ := b1.Intersect(l)
	p2, _ := b2.Intersect(l)

	r.line.Position1 = fyne.Position{
		X: float32(p1.X),
		Y: float32(p1.Y),
	}

	r.line.Position2 = fyne.Position{
		X: float32(p2.X),
		Y: float32(p2.Y),
	}

	r.line.StrokeWidth = r.edge.Width

	if r.edge.Directed {
		r.circle.FillColor = color.Gray{0x99}
		r.circle.StrokeColor = color.Gray{0x99}
		r.circle.StrokeWidth = 1.5

		cirPos := r.line.Position2
		cirPos.X -= 5
		cirPos.Y -= 5

		r.circle.Move(cirPos)

		canvas.Refresh(r.circle)
	}
	canvas.Refresh(r.line)

}

func (r *graphEdgeRenderer) BackgroundColor() color.Color {
	return theme.BackgroundColor()
}

func (r *graphEdgeRenderer) Destroy() {
}

func (r *graphEdgeRenderer) Objects() []fyne.CanvasObject {

	return []fyne.CanvasObject{r.line, r.circle}
}

func (e *GraphEdge) CreateRenderer() fyne.WidgetRenderer {
	r := graphEdgeRenderer{
		edge:   e,
		line:   canvas.NewLine(e.EdgeColor),
		circle: canvas.NewCircle(e.EdgeColor),
	}

	if !e.Directed {
		r.circle.Hide()
	}

	(&r).Refresh()

	return &r
}

func (e *GraphEdge) R2Line() r2.Line {
	return r2.MakeLineFromEndpoints(e.Origin.R2Center(), e.Target.R2Center())
}

func NewGraphEdge(g *GraphWidget, id, relationship string, from, to *GraphNode) *GraphEdge {
	e := &GraphEdge{
		Type:      relationship,
		Graph:     g,
		EdgeColor: theme.ForegroundColor(),
		Width:     2,
		Origin:    from,
		Target:    to,
	}

	g.Edges[id] = e

	log.Println("edge: ", id)

	switch e.Type {

	case SeparationRel:
		e.EdgeColor = color.RGBA{80, 1, 1, 255}
	case CohabatiationRel:
		e.EdgeColor = color.RGBA{8, 6, 151, 255}
	case FriendRel:
		e.EdgeColor = color.RGBA{15, 91, 5, 255}
	case HostileRel:
		e.EdgeColor = color.RGBA{215, 38, 6, 255}
	case AbuseRel:
		e.EdgeColor = color.RGBA{6, 215, 183, 255}
	}

	switch relationship {
	case ChildRel:
		e.Directed = true
		from.Children[id] = e
		to.Parents[id] = e
	default:
		from.Undirected[id] = e
		to.Undirected[id] = e
	}

	e.ExtendBaseWidget(e)

	return e
}
