package dc

import (
	"fmt"
	"math"
	"math/cmplx"
	"math/rand/v2"
	"strconv"
	"strings"
	"unicode"
)

type Token int

const (
	t_snull Token = iota
	t_number
	t_ident_i
	t_ident_z
	t_plus
	t_minus
	t_mult
	t_div
	t_oparen
	t_cparen
	t_power
	t_period
	t_comma

	// function names
	t_fsin
	t_fcos
	t_ftan
	t_fexp
	t_flog
	t_flog10
	t_fint
	t_fsqrt
	t_fasin
	t_facos
	t_fatan
	t_fabs

	t_fc
	t_spi
	t_sphi
	t_pushc
	t_pushz
	t_pushi
	t_pushcc
	t_neg
)

const phi = 0.6180339887
const max_stack = 32

var reserved_word = []string{"sin", "cos", "tan", "exp", "log", "log10", "int", "sqrt", "asin", "acos", "atan", "abs", "c", "pi", "phi"}

var resWordMap = map[string]Token{
	"sin": t_fsin, "cos": t_fcos, "tan": t_ftan,
	"exp": t_fexp, "log": t_flog, "log10": t_flog10,
	"int": t_fint, "sqrt": t_fsqrt, "asin": t_fasin,
	"acos": t_facos, "atan": t_fatan, "abs": t_fabs,
	"c": t_fc, "pi": t_spi, "phi": t_sphi,
}

type ZCompiler struct {
	expr        string
	ixpr        int
	sym         Token
	ch          byte
	nval        float64
	id          string
	err         bool
	err_message string
	code        []int
	constants   []float64
}

func NewCompiler(expr string) ZCompiler {
	zc := ZCompiler{
		expr:        strings.ToLower(expr), // case insensitive
		ixpr:        0,
		sym:         t_snull,
		ch:          0,
		nval:        0,
		id:          "",
		err:         false,
		err_message: "",
		code:        []int{},
		constants:   []float64{},
	}

	zc.getch()
	zc.compile()
	return zc
}

func (zc *ZCompiler) getch() byte {
	zc.ch = 0
	if zc.ixpr < len(zc.expr) {
		zc.ch = zc.expr[zc.ixpr]
		zc.ixpr++
	}
	return zc.ch
}

func (zc *ZCompiler) getsym() Token {
	zc.sym = t_snull
	zc.id = ""
	zc.nval = 0

	for zc.ch != 0 && zc.ch <= ' ' { // skip blanks
		zc.getch()
	}

	switch {
	case unicode.IsLetter(rune(zc.ch)):
		zc.id = ""
		for unicode.IsLetter(rune(zc.ch)) || unicode.IsDigit(rune(zc.ch)) {
			zc.id += string(zc.ch)
			zc.getch()
		}

		switch zc.id {
		case "z":
			zc.sym = t_ident_z
		case "i":
			zc.sym = t_ident_i
		default:
			if sym, found := resWordMap[zc.id]; found {
				zc.sym = sym
			} else {
				zc.err = true
				zc.err_message = fmt.Sprintf("string not recognized: %s", zc.id)
			}
		}
	case unicode.IsDigit(rune(zc.ch)):
		zc.id = ""
		for unicode.IsDigit(rune(zc.ch)) || zc.ch == '.' || zc.ch == 'e' || zc.ch == 'E' {
			zc.id += string(zc.ch)
			zc.getch()
		}
		if nval, err := strconv.ParseFloat(zc.id, 64); err != nil { // check number
			zc.err = true
			zc.err_message = fmt.Sprintf("malformed number: %s", zc.id)
		} else {
			zc.nval = nval
			zc.sym = t_number
		}
	default:
		switch zc.ch {
		case '+':
			zc.sym = t_plus
		case '-':
			zc.sym = t_minus
		case '*':
			zc.sym = t_mult
		case '/':
			zc.sym = t_div
		case '(': //
			zc.sym = t_oparen
		case ')':
			zc.sym = t_cparen
		case '^':
			zc.sym = t_power
		case ',':
			zc.sym = t_comma
		case '.':
			zc.sym = t_period
		case 0:
			zc.sym = t_snull
		default:
			zc.sym = t_snull
			zc.err = true
			zc.err_message = fmt.Sprintf("character not recognized: %c", zc.ch)
		}
		zc.getch() // advance to next char
	}
	return zc.sym
}

func (zc *ZCompiler) compile() {
	zc.getsym()
	zc.ce0()
}

func (zc *ZCompiler) gen_i(token Token, i int) {
	zc.code = append(zc.code, int(token))
	zc.code = append(zc.code, i)
}

func (zc *ZCompiler) gen_f(token Token, d float64) {
	zc.code = append(zc.code, int(token))
	zc.code = append(zc.code, len(zc.constants))
	zc.constants = append(zc.constants, d)
}

func (zc *ZCompiler) gen(token Token) {
	zc.code = append(zc.code, int(token))
}

func (zc *ZCompiler) ce0() {
	zc.ce1()
	for zc.sym == t_plus || zc.sym == t_minus {
		tk := zc.sym
		zc.getsym()
		zc.ce1()
		zc.gen(tk)
	}
}

func (zc *ZCompiler) ce1() {
	zc.ce2()
	for zc.sym == t_mult || zc.sym == t_div {
		tk := zc.sym
		zc.getsym()
		zc.ce2()
		zc.gen(tk)
	}
}

func (zc *ZCompiler) ce2() {
	zc.ce3()
	for zc.sym == t_power {
		tk := zc.sym
		zc.getsym()
		zc.ce3()
		zc.gen(tk)
	}
}

func (zc *ZCompiler) ce3() {
	switch zc.sym {
	case t_ident_z:
		zc.gen(t_pushz)
		zc.getsym()
	case t_ident_i:
		zc.gen(t_pushi)
		zc.getsym()
	case t_number:
		zc.gen_f(t_pushc, zc.nval)
		zc.getsym()
	case t_oparen:
		zc.getsym()
		zc.ce0()
		zc.getsym()
	case t_minus:
		zc.getsym()
		zc.ce3()
		zc.gen(t_neg)
	case t_plus:
		zc.getsym()
		zc.ce3()
	case t_fsin, t_fcos, t_ftan, t_fexp, t_flog, t_flog10, t_fint, t_fsqrt, t_fasin, t_facos, t_fatan, t_fabs:
		tk := zc.sym
		zc.getsym()
		zc.ce3()
		zc.gen(tk)
	case t_fc:
		zc.getsym() // c(1,2)
		zc.getsym() // (
		zc.ce3()
		zc.getsym() // ,
		zc.ce3()
		zc.getsym() // )
		zc.gen(t_fc)
	case t_spi:
		zc.gen_f(t_pushc, math.Pi)
		zc.getsym()
	case t_sphi:
		zc.gen_f(t_pushc, phi)
		zc.getsym()

	default:
		zc.err = true
		zc.err_message = fmt.Sprintf("unexpected symbol: %v", zc.sym)

	}
}

func (zc *ZCompiler) execute(z complex128) complex128 {
	stack := make([]complex128, max_stack) // local stack -> mt support, fixed length
	sp := 0

	for pc := 0; pc < len(zc.code); pc++ {
		switch Token(zc.code[pc]) {
		case t_pushc:
			pc++
			stack[sp] = complex(zc.constants[zc.code[pc]], 0)
			sp++
		case t_pushz:
			stack[sp] = z
			sp++
		case t_pushi:
			stack[sp] = complex(0, 1)
			sp++
		case t_neg:
			stack[sp] = complex(-real(stack[sp]), -imag(stack[sp]))
		case t_plus:
			sp--
			stack[sp-1] += stack[sp]
		case t_minus:
			sp--
			stack[sp-1] -= stack[sp]
		case t_mult:
			sp--
			stack[sp-1] *= stack[sp]
		case t_div:
			sp--
			if stack[sp] != complex(0, 0) {
				stack[sp-1] /= stack[sp]
			}
		case t_power:
			sp--
			if stack[sp-1] != complex(0, 0) {
				stack[sp-1] = cmplx.Pow(stack[sp-1], stack[sp])
			}
		case t_fc:
			sp--
			stack[sp-1] = complex(real(stack[sp-1]), real(stack[sp]))
		case t_fsin:
			stack[sp-1] = cmplx.Sin(stack[sp-1])
		case t_fcos:
			stack[sp-1] = cmplx.Cos(stack[sp-1])
		case t_ftan:
			stack[sp-1] = cmplx.Tan(stack[sp-1])
		case t_fexp:
			stack[sp-1] = cmplx.Exp(stack[sp-1])
		case t_flog:
			if stack[sp-1] != complex(0, 0) {
				stack[sp-1] = cmplx.Log(stack[sp-1])
			}
		case t_flog10:
			if stack[sp-1] != complex(0, 0) {
				stack[sp-1] = cmplx.Log10(stack[sp-1])
			}
		case t_fint:
			stack[sp-1] = complex(math.Floor(real(stack[sp-1])), 0)
		case t_fsqrt:
			stack[sp-1] = cmplx.Sqrt(stack[sp-1])
		case t_fasin:
			stack[sp-1] = cmplx.Asin(stack[sp-1])
		case t_facos:
			stack[sp-1] = cmplx.Acos(stack[sp-1])
		case t_fatan:
			stack[sp-1] = cmplx.Atan(stack[sp-1])
		case t_fabs:
			stack[sp-1] = complex(cmplx.Abs(stack[sp-1]), 0)
		}
	}
	if sp == 1 {
		return stack[sp-1]
	} else {
		return complex(0, 0)
	}
}

const prec_inf = 999

func (zc *ZCompiler) decompile() string {

	type TStack struct {
		val  string
		prec int
	}

	sp := 0
	prec_ := prec_inf
	stack := make([]TStack, 1000)

	op_prec := map[Token]int{
		t_plus:  0,
		t_minus: 0,
		t_mult:  1,
		t_div:   1,
		t_power: 2,
	}
	op_2char := map[Token]rune{
		t_plus:  '+',
		t_minus: '-',
		t_mult:  '*',
		t_div:   '/',
		t_power: '^',
	}

	for pc := 0; pc < len(zc.code); pc++ {
		tk := Token(zc.code[pc])
		switch tk {
		case t_pushc:
			pc++
			stack[sp] = TStack{
				val:  fmt.Sprintf("%.2f", zc.constants[zc.code[pc]]),
				prec: prec_inf,
			}
			sp++
		case t_pushz:
			stack[sp] = TStack{
				val:  "z",
				prec: prec_inf,
			}
			sp++
		case t_pushi:
			stack[sp] = TStack{
				val:  "i",
				prec: prec_inf,
			}
			sp++
		case t_plus, t_minus, t_mult, t_div, t_power:
			sp--
			prec_ = op_prec[tk]

			if stack[sp-1].prec < prec_ { // left <op> | ( left ) <op>
				stack[sp-1].val = "(" + stack[sp-1].val + ")"
			}

			stack[sp-1].val += string(op_2char[tk])

			if stack[sp].prec < prec_ { //  right | (right)
				stack[sp-1].val += "(" + stack[sp].val + ")"
			} else {
				stack[sp-1].val += stack[sp].val
			}

			stack[sp-1].prec = prec_
		case t_fsin, t_fcos, t_ftan, t_fasin, t_facos, t_fatan, t_fexp, t_flog, t_flog10,
			t_fint, t_fsqrt, t_fabs:
			stack[sp-1].val = reserved_word[int(tk)-int(t_fsin)] + "(" +
				stack[sp-1].val + ")"
		case t_fc:
			sp--
			stack[sp-1].val = "c(" + stack[sp-1].val + "," + stack[sp].val + ")"
		default:
		}
	}

	if sp != 0 {
		return stack[sp-1].val
	} else {
		return ""
	}
}

func GenRandomExpression(complexity int) ZCompiler {
	zc := ZCompiler{}

	// helpers on zc
	gen_rand_push := func() { // generate pushz/pushc
		if rand.Float64() < 0.6 {
			zc.gen(t_pushz)
		} else {
			zc.gen_i(t_pushc, rand.IntN(len(zc.constants)))
		}
	}

	gen_rand_operator := func() { // generate a random operator
		opers := []Token{t_plus, t_minus, t_mult, t_div, t_power}
		zc.gen(opers[rand.IntN(len(opers))])
	}

	gen_rand_function := func() { // generate a random function
		zc.gen(Token(int(t_fsin) + rand.IntN(int(t_fabs)-int(t_fsin))))
	}

	gen_rand_array := func(n int) []float64 {
		constants := make([]float64, n)
		for i := range constants {
			constants[i] = rand.Float64()
		}
		return constants
	}

	// generate constants
	zc.constants = gen_rand_array(1000)

	// generate code
	gen_rand_push()
	for range complexity {
		gen_rand_push()
		gen_rand_operator()

		if rand.Float64() >= 0.6 {
			gen_rand_function()
		}
	}

	zc.expr = zc.decompile()

	return zc
}

func GenRandom(complexity int) string {
	// helpers
	randFloat := func() string {
		return strconv.FormatFloat(rand.Float64(), 'f', 2, 64)
	}
	randOper := func() string {
		opers := []string{"+", "-", "*", "/", "^"}
		return opers[rand.IntN(len(opers))]
	}
	randFunc := func() string {
		funcs := []string{"sin", "cos", "tan", "asin", "acos", "atan", "exp", "log", "log10", "int", "sqrt", "abs"}
		return funcs[rand.IntN(len(funcs))]
	}

	// deal with leafs(z, num, c(#,#)) and nodes ( (expr), expr <oper> expr, func(expr))

	if complexity==0 { // random end leaf
		switch rand.IntN(3) {
		case 0:
			return "z"
		case 1:
			return randFloat()
		case 2:
			return "c(" + randFloat() + "," + randFloat() + ")"
		}
	} else {
		switch rand.IntN(10) {
		case 0:
			return "(" + GenRandom(complexity-1) + ")"
		case 1,2,3,4,5,6:
			return GenRandom(complexity-1) + randOper() + GenRandom(complexity-1) 
		case 7,8,9:
			return randFunc() + "(" + GenRandom(complexity-1) + ")"
		}
	}	
	return ""
}

func (zc *ZCompiler) Ok() bool {
	return !zc.err
}