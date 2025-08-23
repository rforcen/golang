package qh

type FaceList struct {
	Head *Face
	Tail *Face
}

func (fl *FaceList) clear() {
	fl.Head, fl.Tail = nil, nil
}

func (fl *FaceList) add(face *Face) {
	if fl.Head == nil {
		fl.Head = face
	} else {
		fl.Tail.Next = face
	}
	face.Next = nil
	fl.Tail = face
}

func (fl *FaceList) first() *Face {
	return fl.Head
}

func (fl *FaceList) isEmpty() bool {
	return fl.Head == nil
}
