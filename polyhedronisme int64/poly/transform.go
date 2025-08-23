package poly

func (p *Polyhedron)Kiss_n(n int, apex_dist float32) *Polyhedron{
    flag := NewFlag()

    p.CalcNormals()
    p.CalcCenters()

    f_ := ToInt("f")

    for nface, face := range p.Faces {
        v1 := face[len(face)-1]
        fname := Int4_2(f_, nface)

        for _,v2 := range face {
            iv2 := Int4_1(v2)
            flag.AddVertex(iv2, p.Vertexes[v2])

            if len(face) == n || n == 0 {
                flag.AddVertex(fname, *p.Centers[nface].Add(p.Normals[nface].Scale(apex_dist)))
                flag.AddFaceVect([]Int4{Int4_1(v1), iv2, fname})
            } else {
                flag.AddFace(Int4_1(nface), Int4_1(v1), iv2)
            }
            v1 = v2
        }
    }

    return flag.CreatePoly("k", p)
}

func (p *Polyhedron)Ambo() *Polyhedron{
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
                flag.AddVertex(m12, *p.Vertexes[v1].Add(&p.Vertexes[v2]).Scale(0.5))
            }
            f_orig = append(f_orig, m12)

            flag.AddFace(Int4_2(orig_, iface), m12, m23)
            flag.AddFace(Int4_2(dual_, v2), m23, m12)
    
            v1, v2 = v2, v3
        }
        flag.AddFaceVect(f_orig)
    }
    return flag.CreatePoly("a", p)
}

func (p *Polyhedron)Quinto() *Polyhedron{
    flag := NewFlag()
    centers := p.CalcCenters()

    for nface, face := range p.Faces {
        centroid := centers[nface]
        v1 := face[len(face)-2]
        v2 := face[len(face)-1]

        vi4 := make([]Int4, 0, len(face))

        for _, v3 := range face {
            t12 := I4_min(v1, v2)
            ti12 := I4_min3(nface, v1, v2)
            t23 := I4_min(v2, v3)
            ti23 := I4_min3(nface, v2, v3)
            iv2 := Int4_1(v2)

            midpt := p.Vertexes[v1].Add(&p.Vertexes[v2]).Scale(0.5)
            innerpt := midpt.Add(&centroid).Scale(0.5)

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

func (p *Polyhedron)Hollow(inset_dist float32, thickness float32) *Polyhedron{
    flag := NewFlag()
    flag.SetVertexes(p.Vertexes)

    avgnormals := p.AvgNormals()
    centers := p.CalcCenters()

    fin_ := ToInt("fin")
    fdwn_ := ToInt("fdwn")
    v_ := ToInt("v")

    for i, face := range p.Faces {
        v1 := face[len(face)-1]

        for _, v2 := range face {
            tw := Tween(&p.Vertexes[v2], &centers[i], inset_dist)

            flag.AddVertex(Int4_4(fin_, i, v_, v2), *tw)
            flag.AddVertex(Int4_4(fdwn_, i, v_, v2), *tw.Sub(avgnormals[i].Scale(thickness)))

            flag.AddFaceVect([]Int4{Int4_1(v1), Int4_1(v2), Int4_4(fin_, i, v_, v2), Int4_4(fin_, i, v_, v1)})
            flag.AddFaceVect([]Int4{Int4_4(fin_, i, v_, v1), Int4_4(fin_, i, v_, v2), Int4_4(fdwn_, i, v_, v2), Int4_4(fdwn_, i, v_, v1)})

            v1 = v2
        }
    }

    return flag.CreatePoly("h", p)
}