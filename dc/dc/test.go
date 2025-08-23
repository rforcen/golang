package dc

import "fmt"

// /////// test compiler
func TestCompiler() {
	fmt.Println("Testing compiler...")
	for _, expr := range Presets {
		zc := NewCompiler(expr)

		fmt.Print(expr, ":")
		for zc.getsym() != t_snull {
			fmt.Printf("%v, ", zc.sym)
		}
		fmt.Println()

		fmt.Println("code:", zc.code)
		fmt.Println("consts:", zc.constants)
		fmt.Println("error:", zc.err, zc.err_message)
		fmt.Println("decompile:", zc.decompile())
		for i := 0; i < 100000; i++ {
			zc.execute(complex(1, 1))
		}
		fmt.Println("execute(1,1):", zc.execute(complex(1, 1)))
		fmt.Println("-------------------------------------------------")
	}
}

func TestRnd() {
	zc := GenRandomExpression(30)
	fmt.Println(zc.expr)
	fmt.Println(zc.code)
	fmt.Println(zc.execute(complex(1, 1)))
}

// /////////////////// dc
func Test_dc1() {
	expr := Presets[1]
	w := 256
	dc_ := NewDC(w, w, expr)
	dc_.GenImageMt()
	dc_.WritePng("0.png")

}
func Test_dc() {
	w := 256 * 4
	for i, expr := range Presets {
		dc_ := NewDC(w, w, expr)
		dc_.GenImageSt()
		fmt.Printf(" LapSt: %4.0f ms, ", dc_.Lap)
		dc_.GenImageMt()
		fmt.Printf("LapMt: %4.0f ms, %02d %s\n", dc_.Lap, i, expr)
		dc_.WritePng(fmt.Sprintf("%d.png", i))
	}
}
func Test_dc_rand() {
	w := 256 * 4
	for i := range 1 {
		dc_ := NewDC(w, w, "")
		dc_.Random(30)
		fmt.Printf("%4.0f ms, %s\n", dc_.Lap, dc_.GetExpression())
		dc_.WritePng(fmt.Sprintf("%d.png", i))
	}
}
func Test_GenRandom() {
	for i:= range 10 {
		expr := GenRandom(4)
		zc := NewCompiler(expr)
		for range 1000 {
			zc.execute(complex(1,1))
		}
		if zc.Ok() {
			fmt.Printf("%02d: %s\n", i, expr)
		}
	}
}