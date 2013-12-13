package main

import (
    "log"
    "github.com/ajstarks/svgo"
    "net/http"
	"fmt"
	"math"
)

const (NW=iota; NE; SE; SW)
const children = SW+1  // This is possibly lame?
const maxDecomp = 8 // Results in maxDecomp+1 levels in tree (0-level is root pointer)

var  TotalCalls int    // Ack! Global!  Used for stats and debugging.
var  TotalLeafNodes int
var  TotalArea float32
var  head   *Node
var  segment Segment

// The next three structs look an awful lot alike!
type Segment struct {
	x0, y0  float32
	x1, y1  float32
}

type Box struct {
	minX, minY  float32
	maxX, maxY  float32
}

type Geom struct {
	x, y   float32
	w, h   float32
}

//  Could we nest Geom in here somehow?
type Node struct {
	geom   Geom
	child  [children]*Node
}



func main() {
	TotalCalls = 0
	segment =  Segment {0.1, 1.0, 0.6, 0.0}
	head = segBoxTree(segment, Geom {0.0, 0.0, 1.0, 1.0}, maxDecomp)
	fmt.Println("segBoxTree TotalCalls: ", TotalCalls)
	TotalCalls = 0
	TotalArea  = 0.0
	incrArea := func(geom Geom) { 
		TotalArea += geom.w * geom.h
		return
	}
	traverseTree(head, incrArea)
	fmt.Println("traverseTree TotalCalls: ", TotalCalls)
	fmt.Println("traverseTree TotalLeafNodes: ", TotalLeafNodes)
	fmt.Println("traverseTree TotalArea: ", TotalArea)
	initWeb()
}


func initWeb() {
    http.Handle("/qtree", http.HandlerFunc(renderSVG))
    err := http.ListenAndServe(":2003", nil)
    if err != nil {
        log.Fatal("ListenAndServe:", err)
    }
}


//  Temporary kludge.
//  Convert geom (in 0.0 -- 1.0 space) to box (in 0 -- 1000 space)
func  svgCoord(g Geom) Geom {
	//fmt.Printf("g: %#v\n", g)
	svgc := Geom {g.x*1000, 1000-(g.y+g.h)*1000, g.w*1000, g.h*1000}
	//fmt.Printf("b: %#v\n\n", svgc)
	return svgc
}


func renderSVG(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	s := svg.New(w)
	s.Start(1000, 1000)
	drawBox := func(geom Geom) { 
		c := svgCoord(geom)
		s.Rect(int(c.x), int(c.y), int(c.w), int(c.h), "fill:none;stroke:grey")
		return
	}
	drawTree(head, drawBox)
	s.Line(int(segment.x0*1000), 1000-int(segment.y0*1000), int(segment.x1*1000), 
			1000-int(segment.y1*1000), "fill:none;stroke:red")
	s.End()
}

//  Recursively build totally bushed-out tree.
func buildTree(geom Geom, level int) *Node {
	TotalCalls++
	// fmt.Println(level, geom)
	var tmp *Node
	if level == 0 {
		return tmp   // Should be nil from initialization
	} else {
		tmp = new(Node)
		tmp.geom = geom
		// Geometry for children.  W and H fixed, new X, Y for each child.
		// 999.9 is bogus value--will change for each child creation
		newGeom := Geom {999.9, 999.9, geom.h/2, geom.w/2}
		// NW child
		newGeom.x = geom.x
		newGeom.y = geom.y + geom.h/2
		tmp.child[NW] = buildTree(newGeom, level-1)
		// NE child
		newGeom.x = geom.x + geom.w/2
		newGeom.y = geom.y + geom.h/2
		tmp.child[NE] = buildTree(newGeom, level-1)
		// SE child
		newGeom.x = geom.x + geom.w/2
		newGeom.y = geom.y
		tmp.child[SE] = buildTree(newGeom, level-1)
		// SW child
		newGeom.x = geom.x
		newGeom.y = geom.y
		tmp.child[SW] = buildTree(newGeom, level-1)
	}
	return tmp
}


//  Recursively build segment-box intersect tree
func segBoxTree(seg Segment, geom Geom, level int) *Node {
	// fmt.Println(level, geom)
	TotalCalls++
	if level == 0 {
		return nil
	}
	tmp := new(Node)
	tmp.geom = geom
	if ! segBox(seg, Box {geom.x, geom.y, geom.x+geom.w, geom.x+geom.h}) {
		return tmp
	}
	// Geometry for children.  W and H fixed, new X, Y for each child.
	// 999.9 is bogus value--will change for each child creation
	newGeom := Geom {999.9, 999.9, geom.h/2, geom.w/2}
	// NW child
	newGeom.x = geom.x
	newGeom.y = geom.y + geom.h/2
	tmp.child[NW] = segBoxTree(seg, newGeom, level-1)
	// NE child
	newGeom.x = geom.x + geom.w/2
	newGeom.y = geom.y + geom.h/2
	tmp.child[NE] = segBoxTree(seg, newGeom, level-1)
	// SE child
	newGeom.x = geom.x + geom.w/2
	newGeom.y = geom.y
	tmp.child[SE] = segBoxTree(seg, newGeom, level-1)
	// SW child
	newGeom.x = geom.x
	newGeom.y = geom.y
	tmp.child[SW] = segBoxTree(seg, newGeom, level-1)
	return tmp
}


//  Should have pluggable functions for...
//  - what to do when at a leaf node (draw it?)
//
//  Think there is a logic error in leaf node check. May miss
//  root-level nodes.  Check it out.
func traverseTree(nodePtr *Node,  leafFunc func(g Geom)) {
	TotalCalls++
	if nodePtr == nil {
		return
	} else {
		//  Check for leaf node.  Lame.  Expensive to check all these
		//  pointers on every call--should have leaf flag???
		if nodePtr.child[NW] == nil && nodePtr.child[NE] == nil &&
			nodePtr.child[SE] == nil && nodePtr.child[SW] == nil {
			TotalLeafNodes++
			leafFunc(nodePtr.geom)
		}
		//  Could just do the slice range here--order not important.
		traverseTree(nodePtr.child[NW], leafFunc)
		traverseTree(nodePtr.child[NE], leafFunc)
		traverseTree(nodePtr.child[SE], leafFunc)
		traverseTree(nodePtr.child[SW], leafFunc)
	}
	return
}


func drawTree(nodePtr *Node, leafFunc func(g Geom)) {
	if nodePtr == nil {
		return
	} else {
		//  Check for leaf node.  Lame.  Expensive to check all these
		//  pointers on every call--should have leaf flag???
		if nodePtr.child[NW] == nil && nodePtr.child[NE] == nil &&
			nodePtr.child[SE] == nil && nodePtr.child[SW] == nil {
			leafFunc(nodePtr.geom)
		}
		//  Could just do the slice range here--order not important.
		drawTree(nodePtr.child[NW], leafFunc)
		drawTree(nodePtr.child[NE], leafFunc)
		drawTree(nodePtr.child[SE], leafFunc)
		drawTree(nodePtr.child[SW], leafFunc)
	}
	return
}


//  Line segment and box intersecton test
func segBox(seg Segment, box Box) bool {
    // Find min and max X for the segment
	minX := seg.x0
	maxX := seg.x1
    if seg.x0 > seg.x1 {
		minX = seg.x1;
		maxX = seg.x0;
    }

    // Find the intersection of the segment's and box's x-projections
    if maxX > box.maxX {
		maxX = box.maxX;
    }
    if minX < box.minX {
		minX = box.minX;
    }
    if minX > maxX {   // If their projections do not intersect return false
		return false;
    }

    // Find corresponding min and max Y for min and max X we found before
	minY := seg.y0
	maxY := seg.y1
    dx   := seg.x1 - seg.x0

    if math.Abs(float64(dx)) > 0.0000001 {
		a := (seg.y1 - seg.y0) / dx
		b := seg.y0 - a * seg.x0
		minY = a * minX + b;
		maxY = a * maxX + b;
    }

    if minY > maxY {  // Swap??
		tmp := maxY;
		maxY = minY;
		minY = tmp;
    }

    // Find the intersection of the segment's and rectangle's y-projections
    if maxY > box.maxY {
		maxY = box.maxY;
    }
    if minY < box.minY {
		minY = box.minY;
    }
    if minY > maxY {   // If Y-projections do not intersect return false
		return false;
    }

    return true;
}
