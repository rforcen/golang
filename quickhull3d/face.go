package qh

import (
	"fmt"

	"github.com/chewxy/math32"
)

const (
	VISIBLE    = 1
	NON_CONVEX = 2
	DELETED    = 3
)

type Face struct {
	he0         *HalfEdge
	area        float32
	planeOffset float32
	numVerts    int
	next        *Face
	mark        int
	outside     *Vertex
	normal      *Vector3d
	centroid    *Vector3d
}

func newFace() *Face {
	return &Face{normal: newVector3d(0, 0, 0), centroid: newVector3d(0, 0, 0), mark: VISIBLE}
}

func createTrig(v0, v1, v2 *Vertex) *Face {
	face := newFace()
	he0 := newHalfEdge(v0, face)
	he1 := newHalfEdge(v1, face)
	he2 := newHalfEdge(v2, face)

	he0.prev = he2 // thread
	he0.next = he1
	he1.prev = he0
	he1.next = he2
	he2.prev = he1
	he2.next = he0

	face.he0 = he0

	// compute the normal and offset
	face.computeNormalCentroid()
	return face
}

func create(vtxArray []*Vertex, indices []int) *Face {
	face := newFace()
	hePrev := (*HalfEdge)(nil)
	for _, index := range indices {
		he := newHalfEdge(vtxArray[index], face)
		if hePrev != nil {
			he.setPrev(hePrev)
			hePrev.setNext(he)
		} else {
			face.he0 = he
		}
		hePrev = he
	}
	face.he0.setPrev(hePrev)
	hePrev.setNext(face.he0)

	// compute the normal and offset
	face.computeNormalCentroid()
	return face
}

func (face *Face) computeCentroid(centroid *Vector3d) {
	centroid.setZero()

	for he := face.he0; ; {
		centroid.accumulate(he.head().pnt)
		he = he.next

		if he == face.he0 {
			break
		}
	}
	centroid.scale(1 / float32(face.numVerts))
}

func (face *Face) computeNormal(normal *Vector3d) {
	he1 := face.he0.next
	he2 := he1.next

	p0 := face.he0.head().pnt
	p2 := he1.head().pnt

	d2x := p2.x - p0.x
	d2y := p2.y - p0.y
	d2z := p2.z - p0.z

	normal.setZero()

	face.numVerts = 2

	for he2 != face.he0 {
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
	face.area = normal.norm()
	normal.scale(1 / face.area)
}

func (face *Face) computeNormalCentroid() {
	face.computeNormal(face.normal)
	face.computeCentroid(face.centroid)

	face.planeOffset = face.normal.dot(face.centroid)
	var numv int = 0

	for he := face.he0; ; {
		numv++
		he = he.next

		if he == face.he0 {
			break
		}
	}
	if numv != face.numVerts {
		fmt.Printf("face %v numVerts=%d should be %d\n", face, face.numVerts, numv)
		panic("face numVerts mismatch")
	}
}

func (face *Face) getEdge(i int) *HalfEdge {
	he := face.he0
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
	return face.he0
}

func (face *Face) distanceToPlane(p *Vector3d) float32 {
	return face.normal.x*p.x + face.normal.y*p.y + face.normal.z*p.z - face.planeOffset
}

func (face *Face) getCentroid() *Vector3d {
	return face.centroid
}

func (face *Face) numVertices() int {
	return face.numVerts
}

func (face *Face) connectHalfEdges(hedgePrev, hedge *HalfEdge) *Face {
	discardedFace := (*Face)(nil)

	if hedgePrev.oppositeFace() == hedge.oppositeFace() { // then there is a redundant edge that we can get rid off

		oppFace := hedge.oppositeFace()
		var hedgeOpp *HalfEdge

		if hedgePrev == face.he0 {
			face.he0 = hedge
		}
		if oppFace.numVertices() == 3 { // then we can get rid of the opposite face altogether
			hedgeOpp = hedge.getOpposite().prev.getOpposite()

			oppFace.mark = DELETED
			discardedFace = oppFace
		} else {
			hedgeOpp = hedge.getOpposite().next

			if oppFace.he0 == hedgeOpp.prev {
				oppFace.he0 = hedgeOpp
			}
			hedgeOpp.prev = hedgeOpp.prev.prev
			hedgeOpp.prev.next = hedgeOpp
		}
		hedge.prev = hedgePrev.prev
		hedge.prev.next = hedge

		hedge.opposite = hedgeOpp
		hedgeOpp.opposite = hedge

		// oppFace was modified, so need to recompute
		oppFace.computeNormalCentroid()
	} else {
		hedgePrev.next = hedge
		hedge.prev = hedgePrev
	}
	return discardedFace
}

func (face *Face) checkConsistency() {
	// do a sanity check on the face
	hedge := face.he0
	maxd := float32(0)
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
		} else if oppFace.mark == DELETED {
			panic("opposite face not on hull")
		}
		d := math32.Abs(face.distanceToPlane(hedge.head().pnt))
		if d > maxd {
			maxd = d
		}
		numv++
		hedge = hedge.next

		if hedge == face.he0 {
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
	oppFace.mark = DELETED

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

	if hedgeAdj == face.he0 {
		face.he0 = hedgeAdjNext
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

	face.computeNormalCentroid()
	face.checkConsistency()

	return numDiscarded
}
