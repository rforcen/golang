package sh

/*
#cgo CFLAGS: -g
#include <stdio.h>
#include <stdlib.h>

typedef struct {
    double x, y, z;
 } CCoord;

typedef struct {
    CCoord coord;
    CCoord normal;
    CCoord color;
    CCoord uv;
 } CLocation;

void printLocations(CLocation* locations, int count) {
    for (int i = 0; i < count; i++) {
        printf("Location %d:\n", i);
        printf("  Coord: (%.2f, %.2f, %.2f)\n", locations[i].coord.x, locations[i].coord.y, locations[i].coord.z);
        printf("  Normal: (%.2f, %.2f, %.2f)\n", locations[i].normal.x, locations[i].normal.y, locations[i].normal.z);
        printf("  Color: (%.2f, %.2f, %.2f)\n", locations[i].color.x, locations[i].color.y, locations[i].color.z);
        printf("  UV: (%.2f, %.2f, %.2f)\n", locations[i].uv.x, locations[i].uv.y, locations[i].uv.z);
    }
 }
*/
import "C"

import (
	"fmt"
	"math/rand"
	"time"
	"unsafe"
)

func test01() {
	fmt.Println("SH test")
	start := time.Now()
	s := NewSH(512, rand.Intn(25), RandCode())
	s.CalcMesh()

	fmt.Println(time.Since(start))
	s.WriteObj("test.obj")
}

func test02() {
	s := NewSH(512, rand.Intn(25), RandCode())
	s.CalcMesh()

	locations := s.Mesh
	fmt.Println(locations[0:4])

	C.printLocations((*C.CLocation)(unsafe.Pointer(&locations[0])), 4)

}
func DoTest() {
	test01()
	test02()
}
