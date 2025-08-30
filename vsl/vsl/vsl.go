package vsl

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
)

// simple helper func's
func GenerateVSL(vslExpr string) (nChannels int, sampleRate int, buffer []float32, err error) {
	vsl := NewVSLCompiler(vslExpr)
	if vsl.compiler.err {
		return 0, 0, []float32{}, fmt.Errorf("%s", vsl.compiler.ErrorMsg())
	}
	return vsl.channels, int(vsl.sampleRate), vsl.GenerateWave(), nil
}
func GenerateVSLFile(path string) (nChannels int, sampleRate int, buffer []float32, err error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return 0, 0, []float32{}, fmt.Errorf("%s", err)
	}
	return GenerateVSL(string(content))
}

// VM Stack
const maxStack = 1024 * 4

type Stack struct {
	data [maxStack]float64
	sp   int
}

func (s *Stack) push(v float64) {
	s.data[s.sp] = v
	s.sp++
}

func (s *Stack) at(index int) float64 {
	return s.data[index]
}

func (s *Stack) operator_sp1(op Token) {
	s.sp--
	data_sp_1 := &s.data[s.sp-1]
	data_sp := &s.data[s.sp]

	switch op {
	case tkPLUS:
		*data_sp_1 += *data_sp
	case tkMINUS:
		*data_sp_1 -= *data_sp
	case tkMULT:
		*data_sp_1 *= *data_sp
	case tkDIV:
		*data_sp_1 /= *data_sp
	case tkEQ:
		*data_sp_1 = bool2float(*data_sp_1 == *data_sp)
	case tkNE:
		*data_sp_1 = bool2float(*data_sp_1 != *data_sp)
	case tkLT:
		*data_sp_1 = bool2float(*data_sp_1 < *data_sp)
	case tkGT:
		*data_sp_1 = bool2float(*data_sp_1 > *data_sp)
	case tkLE:
		*data_sp_1 = bool2float(*data_sp_1 <= *data_sp)
	case tkGE:
		*data_sp_1 = bool2float(*data_sp_1 >= *data_sp)
	case tkPOWER:
		*data_sp_1 = math.Pow(*data_sp_1, *data_sp)
	}
}

func (s *Stack) operator_sp0(op Token, t float64) {

	data_sp_1 := &s.data[s.sp-1]

	switch op {
	case tkFACT:
		*data_sp_1 = factorial(*data_sp_1)
	case tkNEG:
		*data_sp_1 = -*data_sp_1
	case tkFSIN:
		*data_sp_1 = math.Sin(*data_sp_1)
	case tkFCOS:
		*data_sp_1 = math.Cos(*data_sp_1)
	case tkFTAN:
		*data_sp_1 = math.Tan(*data_sp_1)
	case tkFASIN:
		*data_sp_1 = math.Asin(*data_sp_1)
	case tkFACOS:
		*data_sp_1 = math.Acos(*data_sp_1)
	case tkFATAN:
		*data_sp_1 = math.Atan(*data_sp_1)
	case tkFEXP:
		*data_sp_1 = math.Exp(*data_sp_1)
	case tkFINT:
		*data_sp_1 = math.Floor(*data_sp_1)
	case tkFABS:
		*data_sp_1 = math.Abs(*data_sp_1)
	case tkFLOG:
		if *data_sp_1 > 0 {
			*data_sp_1 = math.Log(*data_sp_1)
		} else {
			*data_sp_1 = 0
		}
	case tkFLOG10:
		if *data_sp_1 > 0 {
			*data_sp_1 = math.Log10(*data_sp_1)
		} else {
			*data_sp_1 = 0
		}
	case tkFSQRT:
		if *data_sp_1 >= 0 {
			*data_sp_1 = math.Sqrt(*data_sp_1)
		} else {
			*data_sp_1 = 0
		}
	case tkSEC:
		*data_sp_1 = *data_sp_1 * 2 * math.Pi
	case tkOSC:
		*data_sp_1 = math.Sin(t * *data_sp_1)
	case tkABS:
		*data_sp_1 = math.Abs(*data_sp_1)
	}
}

// VSLCompiler

// helpers
func bool2float(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
func factorial(x float64) float64 {
	if x == 0 {
		return 1
	}
	return x * factorial(x-1)
}

// saw(x), evaluate between 0..2PI, return -1..+1
func saw(x float64) float64 {

	ret, fm := x, x
	if x > TWO_PI {
		fm = math.Mod(x, TWO_PI)
	}
	if fm <= M_PI_2 {
		ret = fm / M_PI_2
	} else if fm <= M_PI3_2 {
		ret = -2 * (fm/math.Pi - 1)
	} else {
		ret = fm/M_PI_2 - 4
	}
	return ret
}

type VSLCompiler struct {
	compiler    *Compiler
	sampleCount int

	sampleRate   float64
	foatingPoint int // 1 f32
	bitsSample   float64
	seconds      float64
	channels     int
	volume       float64
	secEval      int

	errNumber int

	blk_let  FromTo
	blk_code [maxChan]FromTo
}

func NewVSLCompiler(expr string) *VSLCompiler {
	vc := &VSLCompiler{}
	vc.initDefaults()
	vc.compile(expr)

	return vc
}
func (vsl *VSLCompiler) initDefaults() {
	vsl.sampleRate = 44100
	vsl.foatingPoint = 1 // f32
	vsl.bitsSample = -32
	vsl.seconds = 5.
	vsl.channels = 1
	vsl.volume = 0.4
	vsl.secEval = 0
}

func (vsl *VSLCompiler) compile(expr string) bool {
	vsl.compiler = &Compiler{}
	vsl.initDefaults()

	vsl.compiler.compile(expr)

	if !vsl.compiler.parser.err {
		vsl.channels = vsl.compiler.ch

		vsl.blk_let = vsl.compiler.blk_addr._let
		vsl.blk_code = vsl.compiler.blk_addr._code

		// get wave def values from const
		vsl.executeConst() // execute const values

		vsl.compiler.getValue("sample_rate", &vsl.sampleRate)
		vsl.compiler.getValue("bits_sample", &vsl.bitsSample)
		if vsl.bitsSample == -32 {
			vsl.bitsSample = 32
			vsl.foatingPoint = 1
		} else {
			vsl.foatingPoint = 0
		}
		vsl.compiler.getValue("seconds", &vsl.seconds)
		vsl.compiler.getValue("volume", &vsl.volume)
		if vsl.volume > 1 || vsl.volume < 0 {
			vsl.volume = 1
		}
	}

	return !vsl.compiler.parser.err
}

func (vsl *VSLCompiler) executeConst() {
	bc := vsl.compiler.blk_addr._const
	vsl.executeRange(0, bc.from, bc.to)
}

func (vsl *VSLCompiler) executeLet(x float64) {
	vsl.executeRange(x, vsl.blk_let.from, vsl.blk_let.to)
}

func (vsl *VSLCompiler) execute(x float64, channel int) float64 {
	vsl.executeLet(x) // execute let
	return vsl.executeRange(x, vsl.blk_code[channel].from, vsl.blk_code[channel].to)
}

func (vsl *VSLCompiler) executeRange(t float64, from_pc int, to_pc int) float64 {
	// bool err = false;
	stack := Stack{}
	n_params := make([]int, 0)
	sp_base := make([]int, 0)
	sp := &stack.sp

	vsl.secEval++

	*sp = 0
	vsl.errNumber = 0
	code := vsl.compiler.code

	getIntCode := func(pc int) int {
		return int(binary.LittleEndian.Uint64(code[pc : pc+8]))
	}

	for pc := from_pc; pc < to_pc && vsl.errNumber == 0; {
		switch Token(code[pc]) {
		case tkPUSH_CONST:
			pc++
			stack.push(math.Float64frombits(binary.LittleEndian.Uint64(code[pc : pc+8])))
			pc += 8

		case tkPUSH_T:
			pc++
			stack.push(t)

		case tkPUSH_ID:
			pc++
			stack.push(vsl.compiler.tabValues[getIntCode(pc)].di)
			pc += 8

		case tkPARAM: // push param
			pc++
			i0 := sp_base[len(sp_base)-1]
			i1 := n_params[len(n_params)-1]
			i2 := getIntCode(pc)
			stack.push(stack.at(int(i0 - 1 - i1 + i2)))
			pc += 8
		case tkFUNC: // pc, nparams

			stack.push(float64(pc + 1 + 2*8))

			n_params = append(n_params, getIntCode(pc+1+8))
			sp_base = append(sp_base, *sp)
			pc = getIntCode(pc + 1)
		case tkRET:
			nr := getIntCode(pc + 1)
			pc = int(stack.data[*sp-2])
			stack.data[*sp-nr-2] = stack.data[*sp-1]
			*sp -= nr + 2 - 1
			sp_base = sp_base[:len(sp_base)-1]
			n_params = n_params[:len(n_params)-1]

		case tkPOP:
			pc++
			ic := getIntCode(pc)
			vsl.compiler.setValue(ic, stack.data[*sp-1])
			pc += 8

		case tkPLUS, tkMINUS, tkMULT, tkDIV, tkEQ, tkNE, tkLT, tkLE, tkGT, tkGE, tkPOWER:
			stack.operator_sp1(Token(code[pc]))
			pc++

		case tkFACT, tkNEG, tkFSIN, tkFCOS, tkFTAN, tkFASIN, tkFACOS, tkFATAN, tkFEXP, tkFINT, tkFABS, tkFLOG, tkFLOG10, tkFSQRT, tkSEC, tkOSC, tkABS:
			stack.operator_sp0(Token(code[pc]), t)
			pc++

		case tkSWAVE1: // wave(hz)
			pc++
			stack.data[*sp-1] = math.Sin(t * stack.data[*sp-1])
		case tkSWAVE2: // wave(amp, hz)
			pc++
			stack.data[*sp-2] *= math.Sin(t * stack.data[*sp-1])
			*sp--
		case tkSWAVE: // wave(amp, freq, phase)
			pc++
			stack.data[*sp-3] = stack.data[*sp-3] * math.Sin(t*stack.data[*sp-2]+stack.data[*sp-1])
			*sp -= 2

		case tkYINYANG:
			f := &stack.data[*sp-1]
			k := 6. * math.Pi
			pc++
			*f = math.Sin(t**f) * math.Sin(*f/(t+k))

		case tkBACKSLASH: // \{}operator
			pc++
			res := 0.0
			if *sp > 1 {
				if Token(code[pc]) == tkTILDE {
					for *sp--; *sp >= 0; *sp-- {
						res += math.Sin(t * stack.data[*sp])
					}
				} else {
					res = stack.data[*sp-1]
					for *sp -= 2; *sp >= 0; *sp-- {
						switch Token(code[pc]) {
						case tkPLUS:
							res += stack.data[*sp]
						case tkMINUS:
							res -= stack.data[*sp]
						case tkMULT:
							res *= stack.data[*sp]
						case tkDIV:
							res /= stack.data[*sp]
						}
					}
				}
				stack.data[0] = res
				*sp = 1
			}
			pc++

		case tkSEQUENCE:
			{
				n := stack.data[*sp-1]
				end := stack.data[*sp-2]
				ini := stack.data[*sp-3]
				if n < maxStack-10 && end != ini && *sp > 2 {
					if end < ini {
						ini, end = end, ini
					}
					inc := (end - ini) / (n - 1)
					*sp -= 3
					for i := ini; i < end; i += inc {
						stack.push(i)
					}
					stack.push(end)
				}
				pc++
			}

		case tkLAP: // lap(time1,time2)

			s2 := &stack.data[*sp-1]
			s1 := &stack.data[*sp-2]
			if *s2 <= *s1 {
				*s1 = 0
			} else {
				*s1 = bool2float((t >= *s1*(2*math.Pi)) && (t <= *s2*(2*math.Pi)))
			}
			*sp--
			pc++

		case tkSAW1:
			stack.data[*sp-1] = saw(t * stack.data[*sp-1])
			pc++

		// case SAW: // saw(freq, alpha1)
		// {
		//   double &s2 = _stack[-2].d, &s1 = _stack[-1].d;
		//   if (s2 == 0.)
		//     s2 = 0.1;
		//   if (s1 < 0 || s1 > 90)
		//     s1 = 0;
		//   s2 = saw(t * s2, s1);
		//   sp--;
		//   pc++;
		// } break;

		// case HZ2OCT: // hz2oct(freq, oct)
		// {
		//   double &s2 = _stack[-2].d, &s1 = _stack[-1].d;
		//   s2 = FreqInOctave(int(s2), int(s1));
		//   sp--;
		//   pc++;
		// } break;

		// case MAGNETICRING: // MagnetRing(Vol, Hz, Phase, on_count,
		//                    // off_count)
		// {                  // vol is top of _stack

		//   double &vol = _stack[-5].d, hz = _stack[-4].d, ph = _stack[-3].d,
		//          onc = _stack[-2].d, offc = _stack[-1].d;

		//   // double delta=(hz * 2 * M_PI) / samp;
		//   double delta = hz / sample_rate;
		//   if (fmod(sec_eval * delta, (onc + offc)) <= onc)
		//     vol *= sin(t * hz + ph);
		//   else
		//     vol = 0;
		//   sp -= 4;
		//   pc++;
		// } break;

		default:
			// err = true;
			vsl.errNumber = 1
		}
	}

	if vsl.errNumber != 0 { // any error??
		if *sp > 0 {
			stack.data[*sp-1] = 0
		}
		// err = true;
	}
	if *sp == 1 {
		return stack.data[0]
	}
	return 0
}

func (vsl *VSLCompiler) generateBuffer(nSamples int) []float32 {
	tinc := 2 * math.Pi / vsl.sampleRate
	t := float64(vsl.sampleCount) * tinc

	buffer := make([]float32, nSamples)
	for ibuff := 0; ibuff < nSamples; ibuff += vsl.channels {
		for nchan := 0; nchan < vsl.channels; nchan++ {
			buffer[ibuff+nchan] = float32(vsl.volume * vsl.execute(t, nchan))
		}
		t += tinc
	}

	vsl.sampleCount += nSamples / vsl.channels
	if float64(vsl.sampleCount) > vsl.seconds*vsl.sampleRate {
		vsl.sampleCount = 0
	}

	return buffer
}

func (vsl *VSLCompiler) GenerateWave() []float32 {
	return vsl.generateBuffer(int(vsl.seconds * vsl.sampleRate))
}