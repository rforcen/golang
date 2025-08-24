package qh

type Vertex struct {
	pnt   *Vector3d
	index int
	prev  *Vertex
	next  *Vertex
	face  *Face
}

func newVertex(x, y, z float32, index int) *Vertex {
	return &Vertex{newVector3d(x, y, z), index, nil, nil, nil}
}
func newVertexFromPoint(p *Point3d, index int) *Vertex {
	return &Vertex{p, index, nil, nil, nil}
}
