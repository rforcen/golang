package qh

import "github.com/chewxy/math32"

// helper func to create a waterman polyhedron returning faces and vertexes
func WatermanPolyhedron(radius float32) (faces [][]int, vertexes []*Point3d) {
	coords := watermanCoords(radius)
	qh := NewQuickHull3D(coords)
	return qh.GetFaces(), qh.GetVertices()
}

func QuickHull3DFromCoords(coords []float32) (faces [][]int, vertexes []*Point3d) {
	qh := NewQuickHull3D(coords)
	return qh.GetFaces(), qh.GetVertices()
}

const (
	CLOCKWISE                 = 0x1
	INDEXED_FROM_ONE          = 0x2
	INDEXED_FROM_ZERO         = 0x4
	POINT_RELATIVE            = 0x8
	AUTOMATIC_TOLERANCE       = -1
	NONCONVEX_WRT_LARGER_FACE = 1
	NONCONVEX                 = 2
)

type QuickHull3D struct {
	charLength         float32
	pointBuffer        []*Vertex
	vertexPointIndices []int
	faces              []*Face
	horizon            []*HalfEdge
	numVertices        int
	numPoints          int
	explicitTolerance  float32
	tolerance          float32
	discardedFaces     []*Face
	maxVtxs            []*Vertex
	minVtxs            []*Vertex
	newFaces           *FaceList
	unclaimed          *VertexList
	claimed            *VertexList
}

// create a new quick hull from a set of points
func NewQuickHull3D(coords []float32) *QuickHull3D {

	qh := &QuickHull3D{
		charLength:         0,
		pointBuffer:        []*Vertex{},
		vertexPointIndices: []int{},
		faces:              []*Face{},
		horizon:            []*HalfEdge{},
		numVertices:        0,
		numPoints:          0,
		discardedFaces:     make([]*Face, 3),
		explicitTolerance:  AUTOMATIC_TOLERANCE,
		maxVtxs:            make([]*Vertex, 3),
		minVtxs:            make([]*Vertex, 3),
		tolerance:          0,
		newFaces:           &FaceList{},
		unclaimed:          &VertexList{},
		claimed:            &VertexList{},
	}
	qh.setCoords(coords)

	qh.buildHull()
	return qh
}

func (qh *QuickHull3D) setCoords(coords []float32) { // pointBuffer = coords, index
	nump := len(coords) / 3 // number of points (x,y,z)

	if nump < 4 || len(coords)%3 != 0 {
		panic("coords length < 4 or not multiple of 3")
	}

	// initBuffers()
	qh.pointBuffer = make([]*Vertex, nump)
	qh.vertexPointIndices = make([]int, nump)
	qh.numPoints = nump

	// setPoints (coords, index)
	for i := range nump {
		qh.pointBuffer[i] = newVertex(coords[i*3], coords[i*3+1], coords[i*3+2], i)
	}
}

func (qh *QuickHull3D) buildHull() {
	var eyeVtx *Vertex

	qh.calcMaxMin()
	qh.createInitialSimplex()
	for eyeVtx = qh.nextPointToAdd(); eyeVtx != nil; eyeVtx = qh.nextPointToAdd() {
		qh.addPointToHull(eyeVtx)
	}
	qh.reindexFacesVertices()

}

func (qh *QuickHull3D) addPointToFace(vtx *Vertex, face *Face) {
	vtx.face = face

	if face.outside == nil {
		qh.claimed.Add(vtx)
	} else {
		qh.claimed.InsertBefore(vtx, face.outside)
	}
	face.outside = vtx
}

func (qh *QuickHull3D) removePointFromFace(vtx *Vertex, face *Face) {
	if vtx == face.outside {
		if vtx.next != nil && vtx.next.face == face {
			face.outside = vtx.next
		} else {
			face.outside = nil
		}
	}
	qh.claimed.Delete(vtx)
}

func (qh *QuickHull3D) removeAllPointsFromFace(face *Face) *Vertex {
	if face.outside != nil {
		end := face.outside
		for end.next != nil && end.next.face == face {
			end = end.next
		}
		qh.claimed.Delete2(face.outside, end)
		end.next = nil
		return face.outside
	} else {
		return nil
	}
}

func (qh *QuickHull3D) calcMaxMin() {
	max := Vector3d{}
	min := Vector3d{}

	for i := range 3 {
		qh.maxVtxs[i], qh.minVtxs[i] = qh.pointBuffer[0], qh.pointBuffer[0]
	}
	max.setPoint(qh.pointBuffer[0].pnt)
	min.setPoint(qh.pointBuffer[0].pnt)

	for i := 1; i < qh.numPoints; i++ {
		pnt := qh.pointBuffer[i].pnt
		if pnt.x > max.x {
			max.x = pnt.x
			qh.maxVtxs[0] = qh.pointBuffer[i]
		} else if pnt.x < min.x {
			min.x = pnt.x
			qh.minVtxs[0] = qh.pointBuffer[i]
		}
		if pnt.y > max.y {
			max.y = pnt.y
			qh.maxVtxs[1] = qh.pointBuffer[i]
		} else if pnt.y < min.y {
			min.y = pnt.y
			qh.minVtxs[1] = qh.pointBuffer[i]
		}
		if pnt.z > max.z {
			max.z = pnt.z
			qh.maxVtxs[2] = qh.pointBuffer[i]
		} else if pnt.z < min.z {
			min.z = pnt.z
			qh.minVtxs[2] = qh.pointBuffer[i]
		}
	}

	// this epsilon formula comes from QuickHull, and I'm
	// not about to quibble.
	qh.charLength = math32.Max(max.x-min.x, max.y-min.y)
	qh.charLength = math32.Max(max.z-min.z, qh.charLength)

	if qh.explicitTolerance == AUTOMATIC_TOLERANCE {
		qh.tolerance =
			3 * FLOAT_PREC * (math32.Max(math32.Abs(max.x), math32.Abs(min.x)) +
				math32.Max(math32.Abs(max.y), math32.Abs(min.y)) +
				math32.Max(math32.Abs(max.z), math32.Abs(min.z)))
	} else {
		qh.tolerance = qh.explicitTolerance
	}
}

/**
 * Creates the initial simplex from which the hull will be built.
 */
func (qh *QuickHull3D) createInitialSimplex() {
	max := float32(0)
	imax := 0

	for i := range 3 {
		diff := qh.maxVtxs[i].pnt.get(i) - qh.minVtxs[i].pnt.get(i)
		if diff > max {
			max = diff
			imax = i
		}
	}

	if max <= qh.tolerance {
		panic("Input points appear to be coincident")
	}
	vtx := make([]*Vertex, 4)
	// set first two vertices to be those with the greatest
	// one dimensional separation

	vtx[0] = qh.maxVtxs[imax]
	vtx[1] = qh.minVtxs[imax]

	// set third vertex to be the vertex farthest from
	// the line between vtx0 and vtx1
	u01 := Vector3d{}
	diff02 := Vector3d{}
	nrml := Vector3d{}
	xprod := Vector3d{}
	maxSqr := float32(0)
	u01.sub2(vtx[1].pnt, vtx[0].pnt)
	u01.normalize()
	for i := range qh.numPoints {
		diff02.sub2(qh.pointBuffer[i].pnt, vtx[0].pnt)
		xprod.cross(&u01, &diff02)
		lenSqr := xprod.normSquared()
		if lenSqr > maxSqr &&
			qh.pointBuffer[i] != vtx[0] && // paranoid
			qh.pointBuffer[i] != vtx[1] {
			maxSqr = lenSqr
			vtx[2] = qh.pointBuffer[i]
			nrml.setPoint(&xprod)
		}
	}
	if math32.Sqrt(maxSqr) <= 100*qh.tolerance {
		panic("Input points appear to be colinear")
	}
	nrml.normalize()

	maxDist := float32(0)
	d0 := vtx[2].pnt.dot(&nrml)
	for i := range qh.numPoints {
		dist := math32.Abs(qh.pointBuffer[i].pnt.dot(&nrml) - d0)
		if dist > maxDist &&
			qh.pointBuffer[i] != vtx[0] && // paranoid
			qh.pointBuffer[i] != vtx[1] &&
			qh.pointBuffer[i] != vtx[2] {
			maxDist = dist
			vtx[3] = qh.pointBuffer[i]
		}
	}
	if math32.Abs(maxDist) <= 100*qh.tolerance {
		panic("Input points appear to be coplanar")
	}

	tris := make([]*Face, 4)

	if vtx[3].pnt.dot(&nrml)-d0 < 0 {
		tris[0] = createTrig(vtx[0], vtx[1], vtx[2])
		tris[1] = createTrig(vtx[3], vtx[1], vtx[0])
		tris[2] = createTrig(vtx[3], vtx[2], vtx[1])
		tris[3] = createTrig(vtx[3], vtx[0], vtx[2])

		for i := range 3 {
			k := (i + 1) % 3
			tris[i+1].getEdge(1).setOpposite(tris[k+1].getEdge(0))
			tris[i+1].getEdge(2).setOpposite(tris[0].getEdge(k))
		}
	} else {
		tris[0] = createTrig(vtx[0], vtx[2], vtx[1])
		tris[1] = createTrig(vtx[3], vtx[0], vtx[1])
		tris[2] = createTrig(vtx[3], vtx[1], vtx[2])
		tris[3] = createTrig(vtx[3], vtx[2], vtx[0])

		for i := range 3 {
			k := (i + 1) % 3
			tris[i+1].getEdge(0).setOpposite(tris[k+1].getEdge(1))
			tris[i+1].getEdge(2).setOpposite(tris[0].getEdge((3 - i) % 3))
		}
	}

	qh.faces = append(qh.faces, tris...)

	for i := range qh.numPoints {
		v := qh.pointBuffer[i]

		if v == vtx[0] || v == vtx[1] || v == vtx[2] || v == vtx[3] {
			continue
		}

		maxDist := qh.tolerance
		maxFace := (*Face)(nil)
		for k := range 4 {
			dist := tris[k].distanceToPlane(v.pnt)
			if dist > maxDist {
				maxFace = tris[k]
				maxDist = dist
			}
		}
		if maxFace != nil {
			qh.addPointToFace(v, maxFace)
		}
	}
}

func (qh *QuickHull3D) GetNumVertices() int {
	return qh.numVertices
}

func (qh *QuickHull3D) GetVertices() []*Point3d {
	vtxs := make([]*Point3d, qh.numVertices)
	for i := range qh.numVertices {
		vtxs[i] = qh.pointBuffer[qh.vertexPointIndices[i]].pnt
	}
	return vtxs
}

func (qh *QuickHull3D) GetVertexPointIndices() []int {
	indices := make([]int, qh.numVertices)
	copy(indices, qh.vertexPointIndices)
	return indices
}

func (qh *QuickHull3D) GetNumFaces() int {
	return len(qh.faces)
}

func (qh *QuickHull3D) GetFaces() [][]int {
	return qh.getFacesIdxFlag(0)
}

func (qh *QuickHull3D) getFacesIdxFlag(indexFlags int) [][]int {
	allFaces := make([][]int, len(qh.faces))
	for k, face := range qh.faces {
		allFaces[k] = make([]int, face.numVertices())
		qh.getFaceIndices(allFaces[k], face, indexFlags)
	}
	return allFaces
}

func (qh *QuickHull3D) getFaceIndices(indices []int, face *Face, flags int) {
	ccw := ((flags & CLOCKWISE) == 0)
	indexedFromOne := ((flags & INDEXED_FROM_ONE) != 0)
	pointRelative := ((flags & POINT_RELATIVE) != 0)

	hedge := face.he0
	for k := 0; ; {
		idx := hedge.head().index
		if pointRelative {
			idx = qh.vertexPointIndices[idx]
		}
		if indexedFromOne {
			idx++
		}
		indices[k] = idx
		k++

		if ccw {
			hedge = hedge.next
		} else {
			hedge = hedge.prev
		}

		if hedge == face.he0 {
			break
		}
	}
}

func (qh *QuickHull3D) resolveUnclaimedPoints(newFaces *FaceList) {
	vtxNext := qh.unclaimed.First()
	for vtx := vtxNext; vtx != nil; vtx = vtxNext {
		vtxNext = vtx.next

		maxDist := qh.tolerance
		maxFace := (*Face)(nil)
		for newFace := newFaces.first(); newFace != nil; newFace = newFace.next {
			if newFace.mark == VISIBLE {
				dist := newFace.distanceToPlane(vtx.pnt)
				if dist > maxDist {
					maxDist = dist
					maxFace = newFace
				}
				if maxDist > 1000*qh.tolerance {
					break
				}
			}
		}
		if maxFace != nil {
			qh.addPointToFace(vtx, maxFace)
		}
	}
}

func (qh *QuickHull3D) deleteFacePoints(face *Face, absorbingFace *Face) {
	faceVtxs := qh.removeAllPointsFromFace(face)
	if faceVtxs != nil {
		if absorbingFace == nil {
			qh.unclaimed.AddAll(faceVtxs)
		} else {
			vtxNext := faceVtxs
			for vtx := vtxNext; vtx != nil; vtx = vtxNext {
				vtxNext = vtx.next
				dist := absorbingFace.distanceToPlane(vtx.pnt)
				if dist > qh.tolerance {
					qh.addPointToFace(vtx, absorbingFace)
				} else {
					qh.unclaimed.Add(vtx)
				}
			}
		}
	}
}

func (qh *QuickHull3D) oppFaceDistance(he *HalfEdge) float32 {
	return he.face.distanceToPlane(he.opposite.face.getCentroid())
}

func (qh *QuickHull3D) doAdjacentMerge(face *Face, mergeType int) bool {
	hedge := face.he0
	convex := true
	for {
		oppFace := hedge.oppositeFace()
		merge := false
		dist1 := qh.oppFaceDistance(hedge)
		dist2 := qh.oppFaceDistance(hedge.opposite)

		if mergeType == NONCONVEX { // then merge faces if they are definitively non-convex
			if dist1 > -qh.tolerance ||
				dist2 > -qh.tolerance {
				merge = true
			}
		} else // mergeType == NONCONVEX_WRT_LARGER_FACE
		{      // merge faces if they are parallel or non-convex
			// wrt to the larger face; otherwise, just mark
			// the face non-convex for the second pass.
			if face.area > oppFace.area {
				if dist1 > -qh.tolerance {
					merge = true
				} else if dist2 > -qh.tolerance {
					convex = false
				}
			} else {
				if dist2 > -qh.tolerance {
					merge = true
				} else if dist1 > -qh.tolerance {
					convex = false
				}
			}
		}

		if merge {
			numd := face.mergeAdjacentFace(hedge, qh.discardedFaces)
			for i := range numd {
				qh.deleteFacePoints(qh.discardedFaces[i], face)
			}
			return true
		}
		hedge = hedge.next
		if hedge == face.he0 {
			break
		}
	}
	if !convex {
		face.mark = NON_CONVEX
	}
	return false
}

func (qh *QuickHull3D) calculateHorizon(eyePnt *Point3d, edge0 *HalfEdge, face *Face) {

	qh.deleteFacePoints(face, nil)
	face.mark = DELETED

	edge := &HalfEdge{}

	if edge0 == nil {
		edge0 = face.getEdge(0)
		edge = edge0
	} else {
		edge = edge0.getNext()
	}
	for {
		oppFace := edge.oppositeFace()
		if oppFace.mark == VISIBLE {
			if oppFace.distanceToPlane(eyePnt) > qh.tolerance {
				qh.calculateHorizon(eyePnt, edge.getOpposite(), oppFace)
			} else {
				qh.horizon = append(qh.horizon, edge)
			}
		}
		edge = edge.getNext()
		if edge == edge0 {
			break
		}
	}
}

func (qh *QuickHull3D) addAdjoiningFace(eyeVtx *Vertex, he *HalfEdge) *HalfEdge {
	face := createTrig(eyeVtx, he.tail(), he.head())
	qh.faces = append(qh.faces, face)
	face.getEdge(-1).setOpposite(he.getOpposite())
	return face.getEdge(0)
}

func (qh *QuickHull3D) addNewFaces(newFaces *FaceList, eyeVtx *Vertex) {
	newFaces.clear()

	var hedgeSidePrev *HalfEdge = nil
	var hedgeSideBegin *HalfEdge = nil

	for _, o := range qh.horizon {
		horizonHe := o
		hedgeSide := qh.addAdjoiningFace(eyeVtx, horizonHe)
		if hedgeSidePrev != nil {
			hedgeSide.next.setOpposite(hedgeSidePrev)
		} else {
			hedgeSideBegin = hedgeSide
		}
		newFaces.add(hedgeSide.getFace())
		hedgeSidePrev = hedgeSide
	}
	hedgeSideBegin.next.setOpposite(hedgeSidePrev)
}

func (qh *QuickHull3D) nextPointToAdd() *Vertex {
	if !qh.claimed.IsEmpty() {
		eyeFace := qh.claimed.First().face
		var eyeVtx *Vertex = nil
		maxDist := float32(0)
		for vtx := eyeFace.outside; vtx != nil && vtx.face == eyeFace; vtx = vtx.next {
			dist := eyeFace.distanceToPlane(vtx.pnt)
			if dist > maxDist {
				maxDist = dist
				eyeVtx = vtx
			}
		}
		return eyeVtx
	} else {
		return nil
	}
}

func (qh *QuickHull3D) addPointToHull(eyeVtx *Vertex) {
	qh.horizon = make([]*HalfEdge, 0)
	qh.unclaimed = &VertexList{}

	qh.removePointFromFace(eyeVtx, eyeVtx.face)
	qh.calculateHorizon(eyeVtx.pnt, nil, eyeVtx.face)
	qh.newFaces.clear()
	qh.addNewFaces(qh.newFaces, eyeVtx)

	// first merge pass ... merge faces which are non-convex
	// as determined by the larger face

	for face := qh.newFaces.first(); face != nil; face = face.next {
		if face.mark == VISIBLE {
			for qh.doAdjacentMerge(face, NONCONVEX_WRT_LARGER_FACE) {
			}
		}
	}
	// second merge pass ... merge faces which are non-convex
	// wrt either face
	for face := qh.newFaces.first(); face != nil; face = face.next {
		if face.mark == NON_CONVEX {
			face.mark = VISIBLE
			for qh.doAdjacentMerge(face, NONCONVEX) {
			}
		}
	}
	qh.resolveUnclaimedPoints(qh.newFaces)
}

func (qh *QuickHull3D) markFaceVertices(face *Face, mark int) {
	he0 := face.getFirstEdge()

	for he := he0; ; {
		he.head().index = mark
		he = he.next

		if he == he0 {
			break
		}
	}
}

func (qh *QuickHull3D) reindexFacesVertices() {
	for i := range qh.numPoints {
		qh.pointBuffer[i].index = -1
	}
	// remove inactive faces and mark active vertices
	newFaces := make([]*Face, 0, len(qh.faces))
	for _, face := range qh.faces {
		if face.mark == VISIBLE {
			newFaces = append(newFaces, face)

			qh.markFaceVertices(face, 0)
		}
	}
	qh.faces = newFaces

	// reindex vertices
	qh.numVertices = 0
	for i := range qh.numPoints {
		vtx := qh.pointBuffer[i]
		if vtx.index == 0 {
			qh.vertexPointIndices[qh.numVertices] = i
			vtx.index = qh.numVertices
			qh.numVertices++
		}
	}
}
