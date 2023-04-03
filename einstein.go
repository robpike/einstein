// Copyright 2023 Rob Pike. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The einstein command produces a simple STL description, suitable
// for 3D printing, of the "einstein hat" monotile. See
// https://arxiv.org/pdf/2303.10798.pdf for details.
// Given the -r flag, it produces output for the reflected (reversed) tile.
package main // import "robpike.io/cmd/einstein"

import (
	"flag"
	"fmt"
	"log"
	"math"
	"strings"
)

var (
	reflect = flag.Bool("r", false, "reflect")

	// These come up a lot.
	rad   = 180 / math.Pi
	sqrt3 = math.Sqrt(3)
	cos30 = math.Cos(30 / rad)
	sin30 = math.Sin(30 / rad)
)

const (
	unit = 12 // mm

	// To create lines around each tile, we generate an inset kite slightly taller.
	// This creates a groove around the kite.
	inset = 0.03 // Fraction of a unit.
)

type kite struct {
	pos [2]float64
	rot int
}

var kites = []kite{
	0: {
		pos: [2]float64{0, 0},
		rot: -60,
	},
	1: {
		pos: [2]float64{0, 0},
		rot: -120,
	},
	2: {
		pos: [2]float64{0, 0},
		rot: -180,
	},
	3: {
		pos: [2]float64{0, 0},
		rot: -240,
	},
	4: {
		pos: [2]float64{0, -2 * sqrt3},
		rot: 60,
	},
	5: {
		pos: [2]float64{0, -2 * sqrt3},
		rot: 120,
	},
	6: {
		pos: [2]float64{-3, -sqrt3},
		rot: -60,
	},
	7: {
		pos: [2]float64{-3, -sqrt3},
		rot: 0,
	},
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("stl: ")
	flag.Parse()
	for i := range kites {
		fmt.Println(render(fmt.Sprintf("kite%d", i), kites[i], 0, *reflect))
		fmt.Println(render(fmt.Sprintf("kite-inset%d", i), kites[i], inset, *reflect))
	}
}

func render(name string, k kite, inset float64, reflect bool) string {
	var b strings.Builder
	fmt.Fprintf(&b, "solid %s\n\n", name)
	stl := Kite(k.pos, k.rot, inset, reflect)
	for _, f := range stl.Facets() {
		fmt.Fprintln(&b, f)
	}
	return b.String()
}

// We draw the lines by insetting a second tile a bit and then lifting
// it, so we get a groove around the outside.
func Kite(loc [2]float64, rotationDegrees int, inset float64, reflect bool) Box {
	// First describe the kite. Then rotate, then translate.
	x0, y0 := 0., 0. // Bottom of kite.
	x1, y1 := 0., sqrt3
	x2, y2 := 1., sqrt3
	x3, y3 := sqrt3*cos30, sqrt3*sin30

	if inset > 0 {
		deflate := 1 - inset
		// Deflate
		x0, y0 = x0*deflate, y0*deflate
		x1, y1 = x1*deflate, y1*deflate
		x2, y2 = x2*deflate, y2*deflate
		x3, y3 = x3*deflate, y3*deflate
		// Translate
		dx, dy := inset*sin30, inset*cos30
		x0, y0 = x0+dx, y0+dy
		x1, y1 = x1+dx, y1+dy
		x2, y2 = x2+dx, y2+dy
		x3, y3 = x3+dx, y3+dy
	}

	// Scale up to unit size.
	x0, y0 = x0*unit, y0*unit
	x1, y1 = x1*unit, y1*unit
	x2, y2 = x2*unit, y2*unit
	x3, y3 = x3*unit, y3*unit

	// Rotate by angle in degrees.
	// x' = x cos 𝛉 -y sin 𝛉
	// y' =x sin 𝛉 +y cos 𝛉
	𝛉 := float64(rotationDegrees) / rad
	sin𝛉 := math.Sin(𝛉)
	cos𝛉 := math.Cos(𝛉)
	x0, y0 = x0*cos𝛉-y0*sin𝛉, x0*sin𝛉+y0*cos𝛉
	x1, y1 = x1*cos𝛉-y1*sin𝛉, x1*sin𝛉+y1*cos𝛉
	x2, y2 = x2*cos𝛉-y2*sin𝛉, x2*sin𝛉+y2*cos𝛉
	x3, y3 = x3*cos𝛉-y3*sin𝛉, x3*sin𝛉+y3*cos𝛉

	// Translate to destination.
	dx, dy := loc[0]*unit, loc[1]*unit
	x0, y0 = x0+dx, y0+dy
	x1, y1 = x1+dx, y1+dy
	x2, y2 = x2+dx, y2+dy
	x3, y3 = x3+dx, y3+dy

	// Reflect?
	if reflect {
		x0, y0, x1, y1, x2, y2, x3, y3 = -x3, y3, -x2, y2, -x1, y1, -x0, y0
	}

	bot := NewQuad(
		x0, y0, 0,
		x1, y1, 0,
		x2, y2, 0,
		x3, y3, 0,
	)
	var height = (0.2 + inset) * unit
	top := NewQuad(
		x0, y0, height,
		x3, y3, height,
		x2, y2, height,
		x1, y1, height,
	)
	return NewBox(bot, top)
}

type Point struct {
	x, y, z float64
}

func (p Point) String() string {
	return fmt.Sprintf("%.6e %.6e %.6e", p.x, p.y, p.z)
}

func (p Point) sub(q Point) Point {
	return Point{p.x - q.x, p.y - q.y, p.z - q.z}
}

type Facet [3]Point

func (f Facet) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "facet normal %s\n", f.Normal())
	fmt.Fprintf(&b, "  outer loop\n")
	for i := 0; i < 3; i++ {
		fmt.Fprintf(&b, "    vertex %s\n", f[i])
	}
	fmt.Fprintf(&b, "  endloop\n")
	fmt.Fprintf(&b, "endfacet\n")
	return b.String()
}

func (f Facet) Normal() Point {
	u := f[1].sub(f[0])
	v := f[2].sub(f[0])
	nx := u.y*v.z - u.z*v.y
	ny := u.z*v.x - u.x*v.z
	nz := u.x*v.y - u.y*v.x
	norm := math.Sqrt(nx*nx + ny*ny + nz*nz)
	return Point{nx / norm, ny / norm, nz / norm}
}

type Quad [4]Point

func (q Quad) Facets() (f [2]Facet) {
	f[0] = Facet{q[0], q[1], q[2]}
	f[1] = Facet{q[2], q[3], q[0]}
	return
}

func NewQuad(c0, c1, c2, c3, c4, c5, c6, c7, c8, c9, c10, c11 float64) Quad {
	p0 := Point{c0, c1, c2}
	p1 := Point{c3, c4, c5}
	p2 := Point{c6, c7, c8}
	p3 := Point{c9, c10, c11}
	return Quad{p0, p1, p2, p3}
}

type Box [2]Quad // Top and bottom quads; the rest are created on the fly.

func NewBox(q0, q1 Quad) Box {
	if q0.Facets()[0].Normal().z >= 0 {
		log.Fatal("bad bottom normal")
	}
	if q1.Facets()[0].Normal().z <= 0 {
		log.Fatal("bad top normal")
	}
	b := Box{q0, q1}
	return b
}

func (b Box) Facets() (f [12]Facet) {
	top := b[0]
	bot := b[1]
	tmp := top.Facets()
	f[0] = tmp[0]
	f[1] = tmp[1]

	tmp = bot.Facets()
	f[2] = tmp[0]
	f[3] = tmp[1]

	// The odd order here is to guarantee outward-pointing normals.
	// Box construction guarantees top and bottom are correctly chiral.
	tmp = Quad{bot[0], bot[3], top[1], top[0]}.Facets()
	f[4] = tmp[0]
	f[5] = tmp[1]

	tmp = Quad{bot[3], bot[2], top[2], top[1]}.Facets()
	f[6] = tmp[0]
	f[7] = tmp[1]

	tmp = Quad{bot[2], bot[1], top[3], top[2]}.Facets()
	f[8] = tmp[0]
	f[9] = tmp[1]

	tmp = Quad{bot[1], bot[0], top[0], top[3]}.Facets()
	f[10] = tmp[0]
	f[11] = tmp[1]
	return
}
