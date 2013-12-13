package main

import (
    "log"
    "github.com/ajstarks/svgo"
    "net/http"
)


func init() {
    http.Handle("/qtree", http.HandlerFunc(renderSVG))
    err := http.ListenAndServe(":2003", nil)
    if err != nil {
        log.Fatal("ListenAndServe:", err)
    }
}


func renderSVG(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	s := svg.New(w)
	s.Start(1000, 1000)
	drawBox := func(geom Geom) { 
		s.Grid(0, 0, 1000, 1000, 10, "fill:none;stroke:grey")
		return
	}
	drawTree(head, drawBox)
	s.Line(10, 0, 990, 1000, "fill:none;stroke:red")
	s.Circle(250, 250, 125, "fill:none;stroke:black")
	s.End()
}




