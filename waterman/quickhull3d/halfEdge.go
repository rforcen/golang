package qh

type HalfEdge struct {
	vertex   *Vertex
	face     *Face
	next     *HalfEdge
	prev     *HalfEdge
	opposite *HalfEdge
}

func NewHalfEdge(vertex *Vertex, face *Face) *HalfEdge {
	return &HalfEdge{vertex, face, nil, nil, nil}
}

func (he *HalfEdge) getNext() *HalfEdge {
	return he.next
}
func (he *HalfEdge) setNext(next *HalfEdge) {
	he.next = next
}
func (he *HalfEdge) getPrev() *HalfEdge {
	return he.prev
}
func (he *HalfEdge) setPrev(prev *HalfEdge) {
	he.prev = prev
}

func (he *HalfEdge) getVertex() *Vertex {
	return he.vertex
}

func (he *HalfEdge) getFace() *Face {
	return he.face
}

func (he *HalfEdge) getOpposite() *HalfEdge {
	return he.opposite
}

func (he *HalfEdge) setOpposite(opposite *HalfEdge) {
	he.opposite = opposite
	opposite.opposite = he
}

func (he *HalfEdge) head() *Vertex {
	return he.vertex
}

func (he *HalfEdge) tail() *Vertex {
	if he.prev != nil {
		return he.prev.vertex
	}
	return nil
}

func (he *HalfEdge) oppositeFace() *Face {
	if he.opposite != nil {
		return he.opposite.face
	}
	return nil
}

func (he *HalfEdge) length() float64 {
	if he.tail() != nil {
		return he.head().pnt.distance(he.tail().pnt)
	}
	return -1
}

func (he *HalfEdge) lengthSquared() float64 {
	if he.tail() != nil {
		return he.head().pnt.distanceSquared(he.tail().pnt)
	}
	return -1
}
