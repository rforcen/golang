package qh

type VertexList struct {
	head *Vertex
	tail *Vertex
}

func (vl *VertexList) Clear() {
	vl.head, vl.tail = nil, nil
}

func (vl *VertexList) Add(vtx *Vertex) {
	if vl.head == nil {
		vl.head = vtx
	} else {
		vl.tail.next = vtx
	}
	vtx.prev = vl.tail
	vtx.next = nil
	vl.tail = vtx
}

func (vl *VertexList) AddAll(vtx *Vertex) {
	if vl.head == nil {
		vl.head = vtx
	} else {
		vl.tail.next = vtx
	}
	vtx.prev = vl.tail
	for vtx.next != nil {
		vtx = vtx.next
	}
	vl.tail = vtx
}

func (vl *VertexList) Delete(vtx *Vertex) {
	if vtx.prev == nil {
		vl.head = vtx.next
	} else {
		vtx.prev.next = vtx.next
	}
	if vtx.next == nil {
		vl.tail = vtx.prev
	} else {
		vtx.next.prev = vtx.prev
	}
}

func (vl *VertexList) Delete2(vtx1, vtx2 *Vertex) {
	if vtx1.prev == nil {
		vl.head = vtx2.next
	} else {
		vtx1.prev.next = vtx2.next
	}
	if vtx2.next == nil {
		vl.tail = vtx1.prev
	} else {
		vtx2.next.prev = vtx1.prev
	}
}

func (vl *VertexList) InsertBefore(vtx *Vertex, next *Vertex) {
	vtx.prev = next.prev
	if next.prev == nil {
		vl.head = vtx
	} else {
		next.prev.next = vtx
	}
	vtx.next = next
	next.prev = vtx
}

func (vl *VertexList) First() *Vertex {
	return vl.head
}

func (vl *VertexList) IsEmpty() bool {
	return vl.head == nil
}
