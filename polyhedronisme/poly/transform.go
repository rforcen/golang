package poly

import "fmt"

func (p *Polyhedron) Kiss_n(n int, apex_dist float32) *Polyhedron {
	flag := NewFlag()

	p.CalcNormals()
	p.CalcCenters()

	f_ := ToInt("f")

	for nface, face := range p.Faces {
		v1 := face[len(face)-1]
		fname := Int4_2(f_, uint32(nface))

		for _, v2 := range face {
			iv2 := Int4_1(v2)
			flag.AddVertex(iv2, p.Vertexes[v2])

			if len(face) == n || n == 0 {
				flag.AddVertex(fname, *p.Centers[nface].Add(p.Normals[nface].Mulc(apex_dist)))
				flag.AddFaceVect([]Int4{Int4_1(v1), iv2, fname})
			} else {
				flag.AddFace(Int4_1(uint32(nface)), Int4_1(v1), iv2)
			}
			v1 = v2
		}
	}

	return flag.CreatePoly("k", p)
}

func (p *Polyhedron) Ambo() *Polyhedron {
	flag := NewFlag()

	dual_ := ToInt("dual")
	orig_ := ToInt("orig")

	for iface, face := range p.Faces {
		v1 := face[len(face)-2]
		v2 := face[len(face)-1]

		f_orig := make([]Int4, 0, len(face))

		for _, v3 := range face {
			m12 := I4_min(v1, v2)
			m23 := I4_min(v2, v3)

			if v1 < v2 {
				flag.AddVertex(m12, *p.Vertexes[v1].Add(&p.Vertexes[v2]).Mulc(0.5))
			}
			f_orig = append(f_orig, m12)

			flag.AddFace(Int4_2(orig_, uint32(iface)), m12, m23)
			flag.AddFace(Int4_2(dual_, v2), m23, m12)

			v1, v2 = v2, v3
		}
		flag.AddFaceVect(f_orig)
	}
	return flag.CreatePoly("a", p)
}

func (p *Polyhedron) Quinto() *Polyhedron {
	flag := NewFlag()
	centers := p.CalcCenters()

	for nface, face := range p.Faces {
		centroid := centers[nface]
		v1 := face[len(face)-2]
		v2 := face[len(face)-1]

		vi4 := make([]Int4, 0, len(face))

		nface := uint32(nface)
		for _, v3 := range face {
			t12 := I4_min(v1, v2)
			ti12 := I4_min3(nface, v1, v2)
			t23 := I4_min(v2, v3)
			ti23 := I4_min3(nface, v2, v3)
			iv2 := Int4_1(v2)

			midpt := p.Vertexes[v1].Add(&p.Vertexes[v2]).Mulc(0.5)
			innerpt := midpt.Add(&centroid).Mulc(0.5)

			flag.AddVertex(t12, *midpt)
			flag.AddVertex(ti12, *innerpt)

			flag.AddVertex(iv2, p.Vertexes[v2])

			flag.AddFaceVect([]Int4{ti12, t12, iv2, t23, ti23})

			vi4 = append(vi4, ti12)

			v1, v2 = v2, v3
		}
		flag.AddFaceVect(vi4)
	}
	return flag.CreatePoly("q", p)
}

func (p *Polyhedron) Hollow(inset_dist float32, thickness float32) *Polyhedron {
	flag := NewFlag()
	flag.SetVertexes(p.Vertexes)

	avgnormals := p.AvgNormals()
	centers := p.CalcCenters()

	fin_ := ToInt("fin")
	fdwn_ := ToInt("fdwn")
	v_ := ToInt("v")

	for i, face := range p.Faces {
		v1 := face[len(face)-1]
		i := uint32(i)

		for _, v2 := range face {

			tw := Tween(&p.Vertexes[v2], &centers[i], inset_dist)

			flag.AddVertex(Int4_4(fin_, i, v_, v2), *tw)
			flag.AddVertex(Int4_4(fdwn_, i, v_, v2), *tw.Sub(avgnormals[i].Mulc(thickness)))

			flag.AddFaceVect([]Int4{Int4_1(v1), Int4_1(v2), Int4_4(fin_, i, v_, v2), Int4_4(fin_, i, v_, v1)})
			flag.AddFaceVect([]Int4{Int4_4(fin_, i, v_, v1), Int4_4(fin_, i, v_, v2), Int4_4(fdwn_, i, v_, v2), Int4_4(fdwn_, i, v_, v1)})

			v1 = v2
		}
	}

	return flag.CreatePoly("h", p)
}

func (p *Polyhedron) Gyro() *Polyhedron {
	cntr_ := ToInt("cntr")

	flag := NewFlag()
	flag.SetVertexes(p.Vertexes)

	centers := p.CalcCenters()

	for i, face := range p.Faces {
		v1 := face[len(face)-2]
		v2 := face[len(face)-1]

		flag.AddVertex(Int4_2(cntr_, uint32(i)), *centers[i].Unit())

		for _, v3 := range face {
			v3 := uint32(v3)

			flag.AddVertex(Int4_2(v1, v2), *OneThird(&p.Vertexes[v1], &p.Vertexes[v2])) // new v in face
			flag.AddFaceVect([]Int4{Int4_2(cntr_, uint32(i)), Int4_2(v1, v2), Int4_2(v2, v1), Int4_1(v2), Int4_2(v2, v3)})

			v1, v2 = v2, v3
		}
	}

	return flag.CreatePoly("g", p)
}

func (p *Polyhedron) Propellor() *Polyhedron {
	flag := NewFlag()
	flag.SetVertexes(p.Vertexes)

	for i, face := range p.Faces {
		v1 := face[len(face)-2]
		v2 := face[len(face)-1]

		for _, v3 := range face {
			v3 := uint32(v3)

			flag.AddVertex(Int4_2(v1, v2), *OneThird(&p.Vertexes[v1], &p.Vertexes[v2])) // new v in face, 1/3rd along edge

			flag.AddFace(Int4_1(uint32(i)), Int4_2(v1, v2), Int4_2(v2, v3))
			flag.AddFaceVect([]Int4{Int4_2(v1, v2), Int4_2(v2, v1), Int4_1(v2), Int4_2(v2, v3)})

			v1, v2 = v2, v3
		}
	}

	return flag.CreatePoly("p", p)
}

func (p *Polyhedron) Dual() *Polyhedron {
	NewFaceMap := func() map[Int4]Int4 {
		face_map := make(map[Int4]Int4)
		for i, face := range p.Faces {
			v1 := face[len(face)-1]
			for _, v2 := range face {
				face_map[Int4_2(v1, v2)] = Int4_1(uint32(i))
				v1 = v2
			}
		}
		return face_map
	}

	flag := NewFlag()
	face_map := NewFaceMap()
	centers := p.CalcCenters()

	for i, face := range p.Faces {
		v1 := face[len(face)-1]
		flag.AddVertex(Int4_1(uint32(i)), centers[i])

		for _, v2 := range face {
			flag.AddFace(Int4_1(v1), face_map[Int4_2(v1, v2)], Int4_1(uint32(i)))
			v1 = v2
		}
	}

	return flag.CreatePoly("d", p)
}

func (p *Polyhedron)Chamfer(dist float32) *Polyhedron {
    orig_ := ToInt("orig")
    hex_ := ToInt("hex")

    flag := NewFlag()
    normals := p.CalcNormals()

    for i, face := range p.Faces {
		v1 := face[len(face)-1]
		v1new := Int4_2(uint32(i), v1)

		for _, v2 := range face {
			flag.AddVertex(Int4_1(v2), *p.Vertexes[v2].Mulc(1 + dist))
			// Add a new vertex, moved parallel to normal.
			v2new := Int4_2(uint32(i), v2)

			flag.AddVertex(v2new, *p.Vertexes[v2].Add(normals[i].Mulc(dist * 1.5)))

			// Four new flags:
			// One whose face corresponds to the original face:
			flag.AddFace(Int4_2(orig_, uint32(i)), v1new, v2new)

			// And three for the edges of the new hexagon:			
			facename := I4_min3(hex_, v1, v2)
			flag.AddFace(facename, Int4_1(v2), v2new)
			flag.AddFace(facename, v2new, v1new)
			flag.AddFace(facename, v1new, Int4_1(v1))

			v1, v1new = v2, v2new
		}
	}
	return flag.CreatePoly("c", p)
}


func (p *Polyhedron)Inset(n int, inset_dist float32, popout_dist float32) *Polyhedron {
	f_ := ToInt("f")
	ex_ := ToInt("ex")

	flag := NewFlag()
	flag.SetVertexes(p.Vertexes)
	normals := p.CalcNormals()
	centers := p.CalcCenters()

	found_any := false
	for i, face := range p.Faces {
		v1 := face[len(face)-1]
		for _, v2 := range face {
			if len(face) == n || n == 0 {
				found_any = true

				flag.AddVertex(Int4_3(f_, uint32(i), v2), *Tween(&p.Vertexes[v2], &centers[i], inset_dist).Add(normals[i].Mulc(popout_dist)))
				flag.AddFaceVect([]Int4{Int4_1(v1), Int4_1(v2), Int4_3(f_, uint32(i), v2), Int4_3(f_, uint32(i), v1)})
				// new inset, extruded face
				flag.AddFace(Int4_2(ex_, uint32(i)), Int4_3(f_, uint32(i), v1), Int4_3(f_, uint32(i), v2))
			} else {
				flag.AddFace(Int4_1(uint32(i)), Int4_1(v1), Int4_1(v2)) // same old flag, if non-n
			}

			v1 = v2
		}
	}
	if !found_any {
		fmt.Println("no $(n) components where found")
	}

	return flag.CreatePoly("n", p)
}
