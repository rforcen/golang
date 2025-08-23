package qh

type Vertex struct {
	pnt   *Vector3d
	index int
	prev  *Vertex
	next  *Vertex
	face  *Face
}

func NewVertex(x, y, z float64, index int) *Vertex {
	return &Vertex{newVector3d(x, y, z), index, nil, nil, nil}
}
func NewVertexFromPoint(p *Point3d, index int) *Vertex {
	return &Vertex{p, index, nil, nil, nil}
}
