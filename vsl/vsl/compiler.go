package vsl

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
)

type DatType int

const (
	NUM_ID DatType = iota
	PARAM
	FUNC
	STRING_ID
)

const (
	M_PI3_2 = 3. * math.Pi / 2.
	TWO_PI  = 6.28318530717958623
	M_PI_2  = math.Pi / 2
	phi     = 1.618033988749895
)

type TableValues struct {
	id       string
	types    DatType
	di       float64
	address  int
	param_ix int
	n_params int
}

type NotationType int

const (
	Notation_Algebraic = iota
	Notation_RPN
	Notation_Freq_Mesh
)

// BlockAddres
const maxChan = 64

type FromTo struct {
	from int
	to   int
}
type BlockAddress struct {
	_const, _let, _func FromTo
	_code               [maxChan]FromTo
	lastTo              int
}

func (ba *BlockAddress) setConst(from, to int) {
	ba._const = FromTo{from, to}
	ba.lastTo = to
}
func (ba *BlockAddress) setLet(t int) {
	ba._let = FromTo{ba.lastTo, t}
	ba.lastTo = t
}
func (ba *BlockAddress) setFunc(t int) {
	ba._func = FromTo{ba.lastTo, t}
	ba.lastTo = t
}
func (ba *BlockAddress) setCode(_chan, t int) {
	if _chan < maxChan {
		ba._code[_chan] = FromTo{ba.lastTo, t}
		ba.lastTo = t
	}
}

// Compiler
const maxCode = 1024 * 8

type Compiler struct {
	parser    *Parser
	blk_addr  BlockAddress
	err       bool
	sym       Token
	tabValues []TableValues
	code      [maxCode]byte
	pc        int

	ch, nid  int
	notation NotationType // alg / rpn
}

func (c *Compiler) getsym() Token {
	c.sym = c.parser.getsym()
	return c.sym
}
func (c *Compiler) checkGetsym(checkSym Token) Token {
	if c.sym != checkSym {
		c.err = true
	}
	return c.getsym()
}
func (c *Compiler) getsymCheck(checkSym Token) Token {
	c.getsym()
	c.err = c.sym != checkSym
	return c.sym
}
func (c *Compiler) parseIdEqExpr() {
	for {
		if c.getsym() == tkIDENT {
			id := c.parser.id
			if c.getsym() == tkEQ {
				c.getsym()

				switch c.notation {
				case Notation_Algebraic:
					c.expr_0()
				case Notation_RPN:
					c.rpnExpr()
				}

				c.tabValues = append(c.tabValues, TableValues{id, NUM_ID, 0, c.nid, 0, 0})

				c.generateInt(tkPOP, c.nid)
				c.nid++
			} else {
				c.err = true
			}
		} else {
			c.err = true
		}
		if c.sym != tkCOMMA || c.err {
			break
		}
	}

	if c.sym == tkSEMICOLON {
		c.getsym()
	} else {
		c.err = true
	}
}

func (c *Compiler) parseConst() {
	if c.sym == tkCONST {
		c.parseIdEqExpr()
	}
	c.blk_addr.setConst(0, c.pc)
}

func (c *Compiler) parseLet() {
	if c.sym == tkLET {
		c.parseIdEqExpr()
	}
	c.blk_addr.setLet(c.pc)
}

func (c *Compiler) parseFuncs() {
	for c.sym == tkFUNC {
		c.getsymCheck(tkIDENT)

		c.tabValues = append(c.tabValues, TableValues{id: c.parser.id, types: FUNC, address: c.pc})

		ixtv := len(c.tabValues)
		param_ix := 0

		if c.getsym() == tkOPAREN {
			for {
				c.getsymCheck(tkIDENT)
				c.tabValues = append(c.tabValues, TableValues{id: c.parser.id, types: PARAM, param_ix: param_ix})
				param_ix++
				if c.getsym() != tkCOMMA {
					break
				}
			}
			c.checkGetsym(tkCPAREN)
		}
		c.checkGetsym(tkRET)

		switch c.notation {
		case Notation_RPN:
			c.rpnExpr()
		case Notation_Algebraic:
			c.expr_0()
		}
		c.checkGetsym(tkSEMICOLON)

		c.tabValues = c.tabValues[:ixtv]        // remove refs. to parameters
		c.tabValues[ixtv-1].n_params = param_ix // save # of args in

		c.generateInt(tkRET, param_ix)
	}
	c.blk_addr.setFunc(c.pc) // jump over fun def
}

func (c *Compiler) compile(expr string) bool {
	c.parser = NewParser(expr)

	c.getsym()

	switch c.sym {
	default:
		c.notation = Notation_Algebraic
		return c.compile_algebraic()
	case tkALGEBRAIC:
		c.getsymCheck(tkSEMICOLON)
		c.getsym()
		c.notation = Notation_Algebraic
		return c.compile_algebraic()
	case tkRPN:
		c.notation = Notation_RPN
		c.getsymCheck(tkSEMICOLON)
		c.getsym()
		return c.compile_rpn()
	}
}

func (c *Compiler) compile_rpn() bool {
	c.parseConst() // const let var0=expr, var1=expr;

	if c.err {
		return false
	}
	c.parseLet()
	c.parseFuncs()

	for {
		c.rpnExpr()
		c.checkGetsym(tkSEMICOLON)
		c.blk_addr.setCode(c.ch, c.pc)
		c.ch++

		if c.getsym() == tkSNULL {
			break
		}
	}
	return !c.err
}

func (c *Compiler) compile_algebraic() bool {
	c.parseConst() // const let var0=expr, var1=expr;

	if c.err {
		return false
	}
	c.parseLet()
	c.parseFuncs()

	for {
		c.expr_0() //  expr per channel;

		lastSym := c.sym

		if !c.err && c.sym == tkSEMICOLON {
			c.blk_addr.setCode(c.ch, c.pc)
			c.ch++
			c.getsym()
		}

		if c.err || lastSym != tkSEMICOLON {
			break
		}
	}
	return !c.err
}

func (c *Compiler) getValue(ident string, value *float64) {
	for i := 0; i < len(c.tabValues); i++ {
		if c.tabValues[i].id == ident {
			*value = c.tabValues[i].di
			return
		}
	}
}

func (c *Compiler) setValue(nid int, value float64) {
	c.tabValues[nid].di = value
}

type SymbolSet = map[Token]struct{}

func (c *Compiler) startsImplicitMult() bool {
	implicit_mult_start := SymbolSet{tkIDENT: {}, tkIDENT_t: {}, tkOCURL: {}, tkOSQARE: {}, tkOLQUOTE: {}, tkOPAREN: {},
		tkSPI: {}, tkSPHI: {}, tkNUMBER: {}, tkRANDOM: {}, tkTILDE: {}, tkSEQUENCE: {}}
	_, ok := implicit_mult_start[c.sym]
	return ok || (c.sym >= tkFSIN && c.sym <= tkLAP)
}

func (c *Compiler) getIdentIndex() int {
	for i := 0; i < len(c.tabValues); i++ {
		if c.tabValues[i].id == c.parser.id {
			return i
		}
	}
	return -1
}

// generate code
func setCodeFloat(c *Compiler, f float64) {
	binary.LittleEndian.PutUint64(c.code[c.pc:c.pc+8], math.Float64bits(f)) // code[pc..8]=f
	c.pc += 8
}
func setCodeInt(c *Compiler, i int) {
	binary.LittleEndian.PutUint64(c.code[c.pc:c.pc+8], uint64(i))
	c.pc += 8
}

func (c *Compiler) generateFloat(token Token, f float64) {
	c.code[c.pc] = byte(token)
	c.pc++

	setCodeFloat(c, f)
}

func (c *Compiler) generateInt(token Token, i int) {
	c.code[c.pc] = byte(token)
	c.pc++
	setCodeInt(c, i)
}

func (c *Compiler) generate2Int(token Token, p0, p1 int) {
	c.code[c.pc] = byte(token)
	c.pc++
	setCodeInt(c, p0)
	setCodeInt(c, p1)
}

func (c *Compiler) generate(token Token) {
	c.code[c.pc] = byte(token)
	c.pc++
}

func (c *Compiler) expr_0() {
	if !c.err {
		is_neg := (c.sym == tkMINUS)
		if is_neg {
			c.getsym()
		}

		c.expr_1()

		if is_neg {
			c.generate(tkNEG)
		}

		op_set := SymbolSet{tkEQ: {}, tkNE: {}, tkLT: {}, tkLE: {}, tkGT: {}, tkGE: {}, tkPLUS: {}, tkMINUS: {}}
		for {
			sym_op := c.sym
			if _, ok := op_set[c.sym]; ok {
				c.getsym()
				c.expr_1()
				c.generate(sym_op)
			}
			if _, ok := op_set[c.sym]; !ok {
				break
			}
		}
	}
}

func (c *Compiler) expr_1() {
	if !c.err {
		c.expr_2()
		for {
			sym_op := c.sym
			if c.startsImplicitMult() { // not operator-> implicit *, i.e. 2{440}
				c.expr_2()
				c.generate(tkMULT)
			} else {
				switch c.sym {
				case tkMULT, tkDIV:
					c.getsym()
					c.expr_2()
					c.generate(sym_op)
				}
			}
			if c.sym != tkMULT && c.sym != tkDIV && !c.startsImplicitMult() {
				break
			}
		}
	}
}

func (c *Compiler) expr_2() {
	if !c.err {
		c.expr_3()
		for {
			if c.sym == tkPOWER {
				c.getsym()
				c.expr_3()
				c.generate(tkPOWER)
			}
			if c.sym != tkPOWER {
				break
			}
		}
	}
}

func (c *Compiler) expr_3() {
	if !c.err {
		switch c.sym {
		case tkOPAREN:
			c.getsym()
			c.expr_0()
			c.checkGetsym(tkCPAREN)
		case tkNUMBER:
			c.generateFloat(tkPUSH_CONST, c.parser.nval)
			c.getsym()
		
		case tkFLOAT:
			c.generateFloat(tkPUSH_CONST, -32.) // this is the floating_point=true value
			c.getsym()
		case tkIDENT_t: //  't' special var is the parameter in eval call
			c.generate(tkPUSH_T)
			c.getsym()
		case tkIDENT:
			idix := c.getIdentIndex()
			if idix != -1 {
				tv := c.tabValues[idix]
				switch tv.types {
				case STRING_ID:
					c.generateInt(tkPUSH_ID, idix)
				case NUM_ID:
					c.generateInt(tkPUSH_ID, idix)
				case PARAM:
					c.generateInt(tkPARAM, tv.param_ix)
				case FUNC:
					if tv.n_params != 0 {
						c.getsymCheck(tkOPAREN)
						c.getsym()
						for np := 0; np < tv.n_params-1; np++ {
							c.expr_0()
							c.checkGetsym(tkCOMMA)
						}
						c.expr_0()
						c.err = c.sym != tkCPAREN
					}
					c.generate2Int(tkFUNC, tv.address, tv.n_params)
				}
			} else {
				c.err = true
			}
			c.getsym()
		case tkMINUS:
			c.getsym()
			c.expr_3()
			c.generate(tkNEG)
		case tkPLUS:
			c.getsym()
			c.expr_3()
			// +expr nothing to generate
		case tkFACT:
			c.getsym()
			c.expr_3()
			c.generate(tkFACT)
		case tkTILDE:
			c.getsym()
			c.expr_3()
			c.generate(tkSWAVE1)
		case tkYINYANG:
			c.getsym()
			c.expr_3()
			c.generate(tkYINYANG)

		case tkSEQUENCE: // (from, to, inc)
			c.checkGetsym(tkOPAREN)
			c.getsym()
			c.expr_0()
			c.checkGetsym(tkCOMMA)
			c.expr_0()
			c.checkGetsym(tkCOMMA)
			c.expr_0()
			c.checkGetsym(tkCOMMA)
			c.generate(tkSEQUENCE)

		case tkRANDOM:
			c.generateFloat(tkPUSH_CONST, rand.Float64())
			c.getsym()

		case tkOCURL: // {hz}, {amp,hz}, {amp, hz, phase}
			c.getsym()
			c.expr_0()
			if c.sym == tkCOMMA {
				c.getsym()
				c.expr_0()
				if c.sym == tkCOMMA {
					c.getsym()
					c.expr_0()
					c.generate(tkSWAVE)
				} else {
					c.generate(tkSWAVE2)
				}
			} else {
				c.generate(tkSWAVE1)
			}
			c.checkGetsym(tkCCURL)

		case tkOSQARE: // []==sec
			c.getsym()
			c.expr_0()

			c.generate(tkSEC)
			c.checkGetsym(tkCSQUARE)

		case tkVERT_LINE: // |abs|
			c.getsym()
			c.expr_0()
			c.generate(tkABS)
			c.checkGetsym(tkVERT_LINE)

		case tkOLQUOTE: // «f»  -> exp(f*t)
			c.getsym()
			c.expr_0()
			c.checkGetsym(tkCLQUOTE)

			c.generate(tkPUSH_T)
			c.generate(tkMULT)
			c.generate(tkFEXP)

		case tkBACKSLASH: // \s:e\ -> lap(start, end)
			c.getsym()
			if c.sym == tkCOLON { // \:e\ -> lap(0, end)
				c.generateFloat(tkPUSH_CONST, 0.0)
			} else {
				c.expr_0()
			}
			c.getsym() // :
			c.expr_0()
			c.checkGetsym(tkBACKSLASH) // '\'
			c.generate(tkLAP)

		case tkFSIN, tkFCOS, tkFTAN, tkFASIN, tkFACOS, tkFATAN, tkFEXP, tkFINT, tkFABS, tkFLOG, tkFLOG10, tkFSQRT, tkSEC, tkOSC, tkABS:
			tsym := c.sym
			c.getsym()
			c.expr_3()
			c.generate(tsym)

		case tkSPI:
			c.getsym()
			c.generateFloat(tkPUSH_CONST, math.Pi)

		case tkSPHI:
			c.getsym()
			c.generateFloat(tkPUSH_CONST, phi)

		case tkSWAVE: // wave(amp, hz, phase)
			c.getsymCheck(tkOPAREN)
			c.getsym()
			c.expr_0()
			c.checkGetsym(tkCOMMA)
			c.expr_0()
			c.checkGetsym(tkCOMMA)
			c.expr_0()
			c.checkGetsym(tkCPAREN)
			c.generate(tkSWAVE)

		
		// 2 parameter funcs.
		case tkLAP: 
			tsym := c.sym
			c.checkGetsym(tkOPAREN)
			c.getsym()
			c.expr_0()
			c.checkGetsym(tkCOMMA)
			c.expr_0()
			c.checkGetsym(tkCPAREN)
			c.generate(tsym)

		case tkSAW: // saw(freq, alpha)
			c.getsym()
			c.getsym()
			c.expr_0()
			if c.sym == tkCOMMA {
				c.getsym()
				c.expr_0()
				c.getsym()
				c.generate(tkSAW)
			} else {
				c.getsym()
				c.generate(tkSAW1)
			}

		case tkSNULL:
		default:
			c.err = true // syntax error
		}
	}
}

func (c *Compiler) ErrorMsg() string {
	if c.err {
		return fmt.Sprintf("syntax error near line %d, char %c, position: %d", c.parser.lineno, c.parser.ch, c.parser.ich)
	}
	return "ok"
}

func (c *Compiler) CompileRPN() bool {
	c.parseConst() // const let var0=expr, var1=expr;

	if !c.err {
		c.parseLet()
		c.parseFuncs()

		for {
			c.rpnExpr()
			c.checkGetsym(tkSEMICOLON)
			c.blk_addr.setCode(c.ch, c.pc)
			c.ch++
			if c.sym != tkSNULL {
				break
			}
		}
	}
	return !c.err
}

func (c *Compiler) rpnExpr() {
	for {
		switch c.sym {
		case tkNUMBER:
			c.generateFloat(tkPUSH_CONST, c.parser.nval)
		//  't' special var is the parameter in eval call
		case tkIDENT_t:
			c.generate(tkPUSH_T)
		case tkIDENT:
			idix := c.getIdentIndex()
			if idix != -1 {
				tv := c.tabValues[idix]
				switch tv.types {
				case NUM_ID:
					c.generateInt(tkPUSH_ID, idix)
				case STRING_ID:
					c.generateInt(tkPUSH_ID, idix)
				case PARAM:
					c.generateInt(tkPARAM, tv.param_ix)
				case FUNC:
					c.generate2Int(tkFUNC, tv.address, tv.n_params)
				}
			} else {
				c.err = true
			}
		case tkSPI:
			c.generateFloat(tkPUSH_CONST, math.Pi)
		case tkSPHI:
			c.generateFloat(tkPUSH_CONST, phi)
		case tkTILDE:
			c.generate(tkSWAVE1)
		case tkSEQUENCE:
			c.generate(tkSEQUENCE)
		case tkBACKSLASH:
			{ // \{operator:+-*/} compress stack w/operator
				operators := map[Token]struct{}{
					tkPLUS: {}, tkMINUS: {}, tkMULT: {}, tkDIV: {}, tkTILDE: {},
				}
				c.generate(tkBACKSLASH)
				if _, ok := operators[c.getsym()]; ok {
					c.generate(c.sym)
				} else {
					c.err = true
				}

			}
		case tkYINYANG, tkMINUS, tkPLUS, tkDIV, tkMULT, tkFSIN, tkFCOS, tkFTAN, tkFASIN, tkFACOS, tkFATAN, tkFEXP, tkFINT, tkFABS, tkFLOG, tkFLOG10, tkFSQRT, tkSEC, tkOSC, tkABS:
			c.generate(c.sym)
		
		case tkSNULL:
		default:
			c.err = true
		}
		c.getsym()
		if !(c.sym != tkSEMICOLON && c.sym != tkCOMMA && c.sym != tkSNULL) {
			break
		}
	}
}
