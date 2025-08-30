// vsl parser

package vsl

import (
	"fmt"
	"strconv"
)

type Token int

// Tokens
const (
	tkSNULL Token = iota
	tkCONST
	tkLET
	tkRPN
	tkFUNC
	tkRET
	tkPARAM
	tkALGEBRAIC
	tkNUMBER
	tkIDENT
	tkIDENT_t
	tkPLUS
	tkMINUS
	tkMULT
	tkDIV
	tkOPAREN
	tkCPAREN
	tkOCURL
	tkCCURL
	tkOSQARE
	tkCSQUARE
	tkBACKSLASH
	tkRANDOM
	tkVERT_LINE
	tkOLQUOTE
	tkCLQUOTE
	tkYINYANG
	tkSEQUENCE
	tkFACT
	tkTILDE
	tkPOWER
	tkPERIOD
	tkSEMICOLON
	tkCOMMA
	tkCOLON
	tkEQ
	tkGT
	tkGE
	tkLT
	tkLE
	tkNE

	tkFSIN // funcs
	tkFCOS
	tkFTAN
	tkFEXP
	tkFLOG
	tkFLOG10
	tkFINT
	tkFSQRT
	tkFASIN
	tkFACOS
	tkFATAN
	tkFABS
	tkSPI
	tkSPHI
	tkSWAVE
	
	tkSEC
	tkOSC
	tkABS
	tkSAW
	tkSAW1
	tkLAP
	
	tkPUSH_CONST
	tkPUSH_T
	tkPUSH_ID
	tkPOP
	tkNEG
	tkSWAVE1
	tkSWAVE2
	tkFLOAT
	
	tkN_DO
	tkN_RE
	tkN_MI
	tkN_FA
	tkN_SOL
	tkN_LA
	tkN_SI
	tkFLAT
	tkSHARP
)

// Symbols
type SymbolMap map[string]Token

func MergeMaps(maps []SymbolMap) SymbolMap {
	merged := make(SymbolMap)
	for _, m := range maps {
		for key, value := range m {
			merged[key] = value
		}
	}
	return merged
}

var smChars = SymbolMap{
	"+": tkPLUS, "*": tkMULT, "·": tkMULT, "/": tkDIV, "(": tkOPAREN, ")": tkCPAREN, "{": tkOCURL,
	"}": tkCCURL, "[": tkOSQARE, "]": tkCSQUARE, "\\": tkBACKSLASH, "?": tkRANDOM, "!": tkFACT, "^": tkPOWER,
	".": tkPERIOD, ",": tkCOMMA, ":": tkCOLON, ";": tkSEMICOLON, "=": tkEQ, "~": tkTILDE, "π": tkSPI, "Ø": tkSPHI,
	"|": tkVERT_LINE, "‹": tkOLQUOTE, "›": tkCLQUOTE, "⬳": tkSAW, "∿": tkFSIN, "τ": tkIDENT_t,
	"☯": tkYINYANG, "§": tkSEQUENCE, "➡": tkRET, "♭": tkFLAT, "♯": tkSHARP,
}

var smWords = SymbolMap{
	"sin": tkFSIN, "cos": tkFCOS, "tan": tkFTAN, "exp": tkFEXP,
	"log": tkFLOG, "log10": tkFLOG10, "int": tkFINT, "sqrt": tkFSQRT,
	"asin": tkFASIN, "acos": tkFACOS, "atan": tkFATAN, "abs": tkFABS,
	"pi": tkSPI, "phi": tkSPHI, "wave": tkSWAVE, "wave1": tkSWAVE1, "wave2": tkSWAVE2,
	"sec": tkSEC, "osc": tkOSC,
	"saw": tkSAW, "saw1": tkSAW1, "lap": tkLAP, "t": tkIDENT_t,
	"const": tkCONST, "rpn": tkRPN, "algebraic": tkALGEBRAIC, "let": tkLET, "float": tkFLOAT,
	"func": tkFUNC,
}

var sm2ch = SymbolMap{
	">=": tkGE,
	"<=": tkLE,
	"<>": tkNE,
	"->": tkRET,
}

var smReserved = MergeMaps([]SymbolMap{smChars, smWords, sm2ch})

var smInitial = SymbolMap{
	"-": tkMINUS, ">": tkGT, "<": tkLT,
}
var smNotes = SymbolMap{
	"do": tkN_DO, "re": tkN_RE, "mi": tkN_MI,
	"fa": tkN_FA, "sol": tkN_SOL, "la": tkN_LA,
	"si": tkN_SI,
}

var allSymbols = MergeMaps([]SymbolMap{smReserved, smInitial, smNotes})

// Parser
type Parser struct {
	expr        string
	runes       []rune
	ch          rune
	ich, lineno int
	id, str     string
	nval        float64
	i0, i1      int
	err         bool
	err_message string
	sym         Token
}

func NewParser(expr string) *Parser {
	p := &Parser{
		expr:        expr,
		runes:       []rune(expr),
		ich:         0,
		lineno:      1,
		id:          "",
		str:         "",
		nval:        0,
		i0:          0,
		i1:          0,
		err:         false,
		err_message: "",
		sym:         tkSNULL,
	}
	p.getch()
	return p
}

func (p *Parser) getch() rune {
	if p.ich < len(p.runes) {
		p.ch = p.runes[p.ich]
		p.ich++
	} else {
		p.ch = 0
	}
	if p.ch == '\n' {
		p.lineno++
	}
	return p.ch
}
func (p *Parser) ungetch() {
	if p.ich > 1 {
		p.ich--
		p.ch = p.runes[p.ich-1]
	}
}

func (p *Parser) skip_blanks() {
	for p.ch != 0 && p.ch <= ' ' {
		p.getch()
	}
}

func (p *Parser) skip_to_eol() {
	for p.ch != 0 && p.ch != '\n' && p.ch != '\r' {
		p.getch()
	}
}

func (p *Parser) skip_multiline_comment() {
	for ; p.ch != 0 && p.ch != '/'; p.getch() {
		p.getch()
		for p.ch != 0 && (p.ch != '*') {
			p.getch()
		}
	}
	p.getch() // skip last '/'
}

func (p *Parser) skip_blank_comments() { // skip blanks & comments // /**/
	for in_comment := true; in_comment; {
		p.skip_blanks()

		if p.ch == '/' { // skip comment
			if p.getch() == '/' {
				p.skip_to_eol()
			} else if p.ch == '*' { // /**/
				p.skip_multiline_comment()
			} else {
				p.ungetch()

				in_comment = false
			}
		} else {
			break
		}
	}

	p.skip_blanks()
}

func (p *Parser) isAlpha() bool {
	return (p.ch >= 'a' && p.ch <= 'z' || p.ch >= 'A' && p.ch <= 'Z' || p.ch == '_')
}

func (p *Parser) isDigit() bool {
	return (p.ch >= '0' && p.ch <= '9')
}

func (p *Parser) getsym() Token { // -> sym
	p.id = ""
	p.sym = tkSNULL
	ok := false

	p.skip_blank_comments()

	if p.ch == 0 {
		return tkSNULL
	}
	if p.ch != 0 && p.isAlpha() { // ident / res. word / func

		for p.ch != 0 && (p.isAlpha() || p.isDigit()) {
			p.id += string(p.ch)
			p.getch()
		}

		if p.sym, ok = smReserved[p.id]; !ok {
			if p.sym, ok = smNotes[p.id]; !ok {
				p.sym = tkIDENT
			}
		}
	} else if p.isDigit() || p.ch == '.' { // number: dddd.ddde-dd
		for p.ch != 0 && (p.isDigit() || p.ch == '.') {
			p.id += string(p.ch)
			p.getch()
		}
		if p.ch == 'e' || p.ch == 'E' { // exp
			for {
				p.id += string(p.ch)
				p.getch()
				if p.ch == 0 || !(p.isDigit() || p.ch == '+' || p.ch == '-') {
					break
				}
			}
		}
		p.sym = tkNUMBER
		err := error(nil)
		p.nval, err = strconv.ParseFloat(p.id, 64)
		if err != nil {
			p.err = true
			p.err_message = fmt.Sprintf("malformed number: %s", p.id)
		}
	} else if p.ch != 0 {
		if p.sym, ok = smReserved[string(p.ch)]; !ok {
			ch_ant := p.ch
			p.getch()
			if p.sym, ok = smReserved[string(ch_ant)+string(p.ch)]; !ok { // double char symbol >=, <=, <>, ->
				if p.sym, ok = smInitial[string(ch_ant)]; !ok { // -,<,>
					p.err = true
				} else {
					if p.ch != 0 {
						p.ungetch()
					}
				}
			}
		}

		p.getch()
	}

	return p.sym
}

func (p *Parser) symToToken() string {
	switch p.sym {
	case tkIDENT:
		return p.id
	case tkNUMBER:
		return fmt.Sprintf("%f", p.nval)
	default:
		for k, v := range allSymbols {
			if v == p.sym {
				return k
			}
		}
	}
	return ""
}
