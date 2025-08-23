package qh

import (
	"fmt"
	"math"
)

const (
	VISIBLE    = 1
	NON_CONVEX = 2
	DELETED    = 3
)

type Face struct {
	He0         *HalfEdge
	Area        float64
	planeOffset float64
	Index       int
	numVerts    int
	Next        *Face
	Mark        int
	Outside     *Vertex
	normal      *Vector3d
	centroid    *Vector3d
}

func newFace() *Face {
	return &Face{normal: newVector3d(0, 0, 0), centroid: newVector3d(0, 0, 0), Mark: VISIBLE}
}

func createTriangle(v0, v1, v2 *Vertex) *Face {
	return createTriangleMinArea(v0, v1, v2, 0)
}

func createTriangleMinArea(v0, v1, v2 *Vertex, minArea float64) *Face {
	face := newFace()
	he0 := NewHalfEdge(v0, face)
	he1 := NewHalfEdge(v1, face)
	he2 := NewHalfEdge(v2, face)

	he0.prev = he2
	he0.next = he1
	he1.prev = he0
	he1.next = he2
	he2.prev = he1
	he2.next = he0

	face.He0 = he0

	// compute the normal and offset
	face.computeNormalAndCentroidMinArea(minArea)
	return face
}

func create(vtxArray []*Vertex, indices []int) *Face {
	face := newFace()
	hePrev := (*HalfEdge)(nil)
	for _, index1 := range indices {
		he := NewHalfEdge(vtxArray[index1], face)
		if hePrev != nil {
			he.setPrev(hePrev)
			hePrev.setNext(he)
		} else {
			face.He0 = he
		}
		hePrev = he
	}
	face.He0.setPrev(hePrev)
	hePrev.setNext(face.He0)

	// compute the normal and offset
	face.computeNormalAndCentroid()
	return face
}

func (face *Face) computeCentroid(centroid *Vector3d) {
	centroid.setZero()
	he := face.He0
	for {
		centroid.add(he.head().pnt)
		he = he.next

		if he == face.He0 {
			break
		}
	}
	centroid.scale(1 / float64(face.numVerts))
}

func (face *Face) computeNormalMinArea(normal *Vector3d, minArea float64) {
	face.computeNormal(normal)

	if face.Area < minArea {
		fmt.Println("area=", face.Area)
		// make the normal more robust by removing
		// components parallel to the longest edge

		var hedgeMax *HalfEdge = nil
		var lenSqrMax float64 = 0
		he := face.He0
		for {
			lenSqr := he.lengthSquared()
			if lenSqr > lenSqrMax {
				hedgeMax = he
				lenSqrMax = lenSqr
			}
			he = he.next

			if he == face.He0 {
				break
			}
		}

		p2 := hedgeMax.head().pnt
		p1 := hedgeMax.tail().pnt
		lenMax := math.Sqrt(lenSqrMax)
		ux := (p2.x - p1.x) / lenMax
		uy := (p2.y - p1.y) / lenMax
		uz := (p2.z - p1.z) / lenMax
		dot := normal.x*ux + normal.y*uy + normal.z*uz
		normal.x -= dot * ux
		normal.y -= dot * uy
		normal.z -= dot * uz

		normal.normalize()
	}
}

func (face *Face) computeNormal(normal *Vector3d) {
	he1 := face.He0.next
	he2 := he1.next

	p0 := face.He0.head().pnt
	p2 := he1.head().pnt

	d2x := p2.x - p0.x
	d2y := p2.y - p0.y
	d2z := p2.z - p0.z

	normal.setZero()

	face.numVerts = 2

	for he2 != face.He0 {
		d1x := d2x
		d1y := d2y
		d1z := d2z

		p2 := he2.head().pnt
		d2x = p2.x - p0.x
		d2y = p2.y - p0.y
		d2z = p2.z - p0.z

		normal.x += d1y*d2z - d1z*d2y
		normal.y += d1z*d2x - d1x*d2z
		normal.z += d1x*d2y - d1y*d2x

		he1 = he2
		he2 = he2.next
		face.numVerts++
	}
	face.Area = normal.norm()
	normal.scale(1 / face.Area)
}

func (face *Face) computeNormalAndCentroid() {
	face.computeNormal(face.normal)
	face.computeCentroid(face.centroid)

	face.planeOffset = face.normal.dot(face.centroid)
	var numv int = 0
	var he *HalfEdge = face.He0

	for {
		numv++
		he = he.next

		if he == face.He0 {
			break
		}
	}
	if numv != face.numVerts {
		fmt.Printf("face %v numVerts=%d should be %d\n", face, face.numVerts, numv)
		panic("face numVerts mismatch")
	}
}

func (face *Face) computeNormalAndCentroidMinArea(minArea float64) {
	face.computeNormalMinArea(face.normal, minArea)
	face.computeCentroid(face.centroid)
	face.planeOffset = face.normal.dot(face.centroid)
}

func (face *Face) getEdge(i int) *HalfEdge {
	he := face.He0
	for i > 0 {
		he = he.next
		i--
	}
	for i < 0 {
		he = he.prev
		i++
	}
	return he
}

func (face *Face) getFirstEdge() *HalfEdge {
	return face.He0
}

func (face *Face) findEdge(vt, vh *Vertex) *HalfEdge {
	var he *HalfEdge = face.He0
	for {
		if he.head() == vh && he.tail() == vt {
			return he
		}
		he = he.next

		if he == face.He0 {
			break
		}
	}
	return nil
}

func (face *Face) distanceToPlane(p *Vector3d) float64 {
	return face.normal.x*p.x + face.normal.y*p.y + face.normal.z*p.z - face.planeOffset
}

func (face *Face) getNormal() *Vector3d {
	return face.normal
}

func (face *Face) getCentroid() *Vector3d {
	return face.centroid
}

func (face *Face) numVertices() int {
	return face.numVerts
}

func (face *Face) getVertexIndices(idxs []int) {
	he := face.He0
	i := 0
	for {
		idxs[i] = he.head().index
		i++
		he = he.next

		if he == face.He0 {
			break
		}
	}
}

func (face *Face) connectHalfEdges(hedgePrev, hedge *HalfEdge) *Face {
	discardedFace := (*Face)(nil)

	if hedgePrev.oppositeFace() == hedge.oppositeFace() { // then there is a redundant edge that we can get rid off

		oppFace := hedge.oppositeFace()
		var hedgeOpp *HalfEdge

		if hedgePrev == face.He0 {
			face.He0 = hedge
		}
		if oppFace.numVertices() == 3 { // then we can get rid of the opposite face altogether
			hedgeOpp = hedge.getOpposite().prev.getOpposite()

			oppFace.Mark = DELETED
			discardedFace = oppFace
		} else {
			hedgeOpp = hedge.getOpposite().next

			if oppFace.He0 == hedgeOpp.prev {
				oppFace.He0 = hedgeOpp
			}
			hedgeOpp.prev = hedgeOpp.prev.prev
			hedgeOpp.prev.next = hedgeOpp
		}
		hedge.prev = hedgePrev.prev
		hedge.prev.next = hedge

		hedge.opposite = hedgeOpp
		hedgeOpp.opposite = hedge

		// oppFace was modified, so need to recompute
		oppFace.computeNormalAndCentroid()
	} else {
		hedgePrev.next = hedge
		hedge.prev = hedgePrev
	}
	return discardedFace
}

func (face *Face) checkConsistency() {
	// do a sanity check on the face
	hedge := face.He0
	maxd := 0.0
	numv := 0

	if face.numVerts < 3 {
		panic("degenerate face: ")
	}
	for {
		hedgeOpp := hedge.getOpposite()
		if hedgeOpp == nil {
			panic("unreflected half edge")
		} else if hedgeOpp.getOpposite() != hedge {
			panic("opposite half edge has opposite")
		}
		if hedgeOpp.head() != hedge.tail() ||
			hedge.head() != hedgeOpp.tail() {
			panic("half edge reflected by")
		}
		oppFace := hedgeOpp.face
		if oppFace == nil {
			panic("no face on half edge")
		} else if oppFace.Mark == DELETED {
			panic("opposite face not on hull")
		}
		d := math.Abs(face.distanceToPlane(hedge.head().pnt))
		if d > maxd {
			maxd = d
		}
		numv++
		hedge = hedge.next

		if hedge == face.He0 {
			break
		}
	}
	if numv != face.numVerts {
		panic(fmt.Sprintf("face, numVerts=%d should be %d", face.numVerts, numv))
	}
}

func (face *Face) mergeAdjacentFace(hedgeAdj *HalfEdge, discarded []*Face) int {
	oppFace := hedgeAdj.oppositeFace()
	var numDiscarded int = 0
	discarded[numDiscarded] = oppFace
	numDiscarded++
	oppFace.Mark = DELETED

	hedgeOpp := hedgeAdj.getOpposite()

	hedgeAdjPrev := hedgeAdj.prev
	hedgeAdjNext := hedgeAdj.next
	hedgeOppPrev := hedgeOpp.prev
	hedgeOppNext := hedgeOpp.next

	for hedgeAdjPrev.oppositeFace() == oppFace {
		hedgeAdjPrev = hedgeAdjPrev.prev
		hedgeOppNext = hedgeOppNext.next
	}

	for hedgeAdjNext.oppositeFace() == oppFace {
		hedgeOppPrev = hedgeOppPrev.prev
		hedgeAdjNext = hedgeAdjNext.next
	}

	var hedge *HalfEdge

	for hedge = hedgeOppNext; hedge != hedgeOppPrev.next; hedge = hedge.next {
		hedge.face = face
	}

	if hedgeAdj == face.He0 {
		face.He0 = hedgeAdjNext
	}

	// handle the half edges at the head
	discardedFace := face.connectHalfEdges(hedgeOppPrev, hedgeAdjNext)

	if discardedFace != nil {
		discarded[numDiscarded] = discardedFace
		numDiscarded++
	}

	// handle the half edges at the tail
	discardedFace = face.connectHalfEdges(hedgeAdjPrev, hedgeOppNext)
	if discardedFace != nil {
		discarded[numDiscarded] = discardedFace
		numDiscarded++
	}

	face.computeNormalAndCentroid()
	face.checkConsistency()

	return numDiscarded
}

func (face *Face) areaSquared(hedge0, hedge1 *HalfEdge) float64 {
	// return the squared area of the triangle defined
	// by the half edge hedge0 and the point at the
	// head of hedge1.

	p0 := hedge0.tail().pnt
	p1 := hedge0.head().pnt
	p2 := hedge1.head().pnt

	dx1 := p1.x - p0.x
	dy1 := p1.y - p0.y
	dz1 := p1.z - p0.z

	dx2 := p2.x - p0.x
	dy2 := p2.y - p0.y
	dz2 := p2.z - p0.z

	x := dy1*dz2 - dz1*dy2
	y := dz1*dx2 - dx1*dz2
	z := dx1*dy2 - dy1*dx2

	return x*x + y*y + z*z
}

func (face *Face) triangulate(newFaces *FaceList, minArea float64) {
	if face.numVertices() < 4 {
		return
	}

	v0 := face.He0.head()
	hedge := face.He0.next
	oppPrev := hedge.opposite
	var face0 *Face = nil

	for hedge = hedge.next; hedge != face.He0.prev; hedge = hedge.next {
		face := createTriangleMinArea(v0, hedge.prev.head(), hedge.head(), minArea)
		face.He0.next.setOpposite(oppPrev)
		face.He0.prev.setOpposite(hedge.opposite)
		oppPrev = face.He0
		newFaces.add(face)
		if face0 == nil {
			face0 = face
		}
	}
	hedge = NewHalfEdge(face.He0.prev.prev.head(), face)
	hedge.setOpposite(oppPrev)

	hedge.prev = face.He0
	hedge.prev.next = hedge

	hedge.next = face.He0.prev
	hedge.next.prev = hedge

	face.computeNormalAndCentroidMinArea(minArea)
	face.checkConsistency()

	for face := face0; face != nil; face = face.Next {
		face.checkConsistency()
	}

}
