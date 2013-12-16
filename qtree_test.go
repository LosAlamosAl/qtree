// go test -bench=.
package main

import "testing"


func BenchmarkSegBox(b *testing.B) {
    for i := 0; i < b.N; i++ {
		segBox(Segment {0.0, 0.0, 1.0, 1.0}, Box {0.0, 0.3, 1.0, 0.6})
    }
}

func BenchmarkSegBoxTree(b *testing.B) {
    for i := 0; i < b.N; i++ {
		segBoxTree(Segment {0.1, 1.0, 0.6, 0.0}, 
			       Geom {0.0, 0.0, 1.0, 1.0}, maxDecomp)
	}
}
 
