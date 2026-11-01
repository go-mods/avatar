package avatar

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// point is a position in a shape outline: the outline fits between -1 and 1 on both axes, around the
// center of the svg, with y going down
type point struct {
	x, y float64
}

// arc is a clockwise circular arc from the previous point of an outline
type arc struct {
	radius float64
	to     point
}

// octagonCorner is where the corners of a regular octagon cut the sides of its bounding square
const octagonCorner = math.Sqrt2 - 1

var (
	// diamondPoints are the corners of a diamond
	diamondPoints = []point{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}

	// octagonPoints are the corners of a regular octagon
	octagonPoints = []point{
		{-octagonCorner, -1}, {octagonCorner, -1}, {1, -octagonCorner}, {1, octagonCorner},
		{octagonCorner, 1}, {-octagonCorner, 1}, {-1, octagonCorner}, {-1, -octagonCorner},
	}

	// heartStart is the bottom tip of a heart
	heartStart = point{0, 1}

	// heartCurves are the cubic Bézier curves of a heart, three points for each: two control points
	// and an end point
	heartCurves = []point{
		{-0.35, 0.72}, {-1, 0.3}, {-1, -0.35},
		{-1, -0.75}, {-0.72, -0.95}, {-0.45, -0.95},
		{-0.25, -0.95}, {-0.08, -0.85}, {0, -0.62},
		{0.08, -0.85}, {0.25, -0.95}, {0.45, -0.95},
		{0.72, -0.95}, {1, -0.75}, {1, -0.35},
		{1, 0.3}, {0.35, 0.72}, {0, 1},
	}

	// cloudStart is the left end of the flat base of a cloud
	cloudStart = point{-0.62, 0.65}

	// cloudArcs are the bumps of a cloud, from left to right
	cloudArcs = []arc{
		{0.38, point{-0.72, -0.09}},
		{0.32, point{-0.27, -0.5}},
		{0.42, point{0.42, -0.35}},
		{0.25, point{0.72, -0.09}},
		{0.38, point{0.62, 0.65}},
	}
)

// shapeElement returns the svg element of the shape. All the shapes but the rectangle are centered and
// fit in the smallest side, without the padding and the border.
func shapeElement(opts avatarGenerator) string {
	inset := opts.padding + opts.borderWidth
	centerX, centerY := opts.width/2, opts.height/2
	radius := min(opts.width, opts.height)/2 - inset
	paint := fmt.Sprintf(`fill="%s" stroke="%s" stroke-width="%d" stroke-dasharray="%s" />`,
		opts.backgroundColor, opts.borderColor, opts.borderWidth, opts.borderDash)

	// at returns the position of a point of an outline in the svg
	at := func(p point) string {
		return formatCoordinate(float64(centerX)+p.x*float64(radius)) + "," +
			formatCoordinate(float64(centerY)+p.y*float64(radius))
	}

	switch opts.shape {
	case shapeRect:
		return fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" rx="%d" ry="%d" `,
			inset, inset, opts.width-inset*2, opts.height-inset*2, opts.borderRadius, opts.borderRadius) + paint
	case shapeDiamond:
		return polygonElement(diamondPoints, at) + paint
	case shapeOctagon:
		return polygonElement(octagonPoints, at) + paint
	case shapeHeart:
		return pathElement(heartData(at)) + paint
	case shapeCloud:
		return pathElement(cloudData(at, float64(radius))) + paint
	default:
		return fmt.Sprintf(`<circle cx="%d" cy="%d" r="%d" `, centerX, centerY, radius) + paint
	}
}

// polygonElement starts a polygon element, up to its paint
func polygonElement(points []point, at func(point) string) string {
	positions := make([]string, len(points))
	for i, p := range points {
		positions[i] = at(p)
	}
	return fmt.Sprintf(`<polygon points="%s" `, strings.Join(positions, " "))
}

// pathElement starts a path element, up to its paint. The joins are round: the sharp hollows of a
// curved outline would otherwise draw a spike of the border color inside the shape.
func pathElement(data string) string {
	return fmt.Sprintf(`<path d="%s" stroke-linejoin="round" `, data)
}

// heartData returns the path data of a heart
func heartData(at func(point) string) string {
	data := "M" + at(heartStart)
	for i := 0; i < len(heartCurves); i += 3 {
		data += " C" + at(heartCurves[i]) + " " + at(heartCurves[i+1]) + " " + at(heartCurves[i+2])
	}
	return data + " Z"
}

// cloudData returns the path data of a cloud
func cloudData(at func(point) string, scale float64) string {
	data := "M" + at(cloudStart)
	for _, a := range cloudArcs {
		radius := formatCoordinate(a.radius * scale)
		data += " A" + radius + "," + radius + " 0 0 1 " + at(a.to)
	}
	return data + " Z"
}

// formatCoordinate writes a coordinate with at most two decimals, far below what a screen can show
func formatCoordinate(value float64) string {
	return strconv.FormatFloat(math.Round(value*100)/100, 'f', -1, 64)
}
