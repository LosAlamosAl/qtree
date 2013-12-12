package main

import (
	"fmt"
	"math"
    "log"
    "github.com/ajstarks/svgo"
    "net/http"
)

const (NW=iota; NE; SE; SW)
const children = SW+1  // This is possibly lame?
const maxDecomp = 4    // Results in maxDecomp+1 levels in tree

var  TotalCalls int    // Ack! Global!  Used for stats and debugging.
var  TotalLeafNodes int
var  TotalArea float32

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
/*
	head := buildTree(Geom {0.0, 0.0, 1.0, 1.0}, maxDecomp)
	fmt.Println("head: ", head)
	fmt.Println("buildTree TotalCalls: ", TotalCalls)
	TotalCalls = 0
	traverseTree(head)
	fmt.Println("traverseTree TotalCalls: ", TotalCalls)
	fmt.Println("traverseTree TotalLeafNodes: ", TotalLeafNodes)
	fmt.Println("traverseTree TotalArea: ", TotalArea)
*/
	TotalCalls = 0
	aHead := segBoxTree(Segment {0.1, 1.0, 0.9, 0.0}, 
		Geom {0.0, 0.0, 1.0, 1.0}, maxDecomp)
	fmt.Println("segBoxTree TotalCalls: ", TotalCalls)
	TotalCalls = 0
	TotalArea  = 0.0
	traverseTree(aHead)
	fmt.Println("traverseTree TotalCalls: ", TotalCalls)
	fmt.Println("traverseTree TotalLeafNodes: ", TotalLeafNodes)
	fmt.Println("traverseTree TotalArea: ", TotalArea)

    http.Handle("/qtree", http.HandlerFunc(renderSVG))
    err := http.ListenAndServe(":2003", nil)
    if err != nil {
        log.Fatal("ListenAndServe:", err)
    }
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


func renderSVG(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	s := svg.New(w)
	s.Start(1000, 1000)
	drawTree(head)
	s.Grid(0, 0, 1000, 1000, 10, "fill:none;stroke:black")
	s.Circle(250, 250, 125, "fill:none;stroke:black")
	s.End()
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
func traverseTree(nodePtr *Node) {
	TotalCalls++
	if nodePtr == nil {
		return
	} else {
		//  Check for leaf node.  Lame.  Expensive to check all these
		//  pointers on every call--should have leaf flag???
		if nodePtr.child[NW] == nil && nodePtr.child[NE] == nil &&
			nodePtr.child[SE] == nil && nodePtr.child[SW] == nil {
			TotalLeafNodes++
			TotalArea += nodePtr.geom.w * nodePtr.geom.h
		}
		//  Could just do the slice range here--order not important.
		traverseTree(nodePtr.child[NW])
		traverseTree(nodePtr.child[NE])
		traverseTree(nodePtr.child[SE])
		traverseTree(nodePtr.child[SW])
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
