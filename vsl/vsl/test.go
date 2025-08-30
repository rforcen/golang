package vsl

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jfreymuth/pulse"
	"github.com/jonchammer/audio-io/wave"
)

func TestParser() {
	expr := "sin(f t/cos(f t)); cos (f t) const ☯ § ✬ πØ➡ 1+20.0 >=+<+>+->--> <> > >= < <= <>☯"
	fmt.Println(expr)

	prs := NewParser(expr)
	for prs.getsym() != tkSNULL {
		fmt.Printf("%v %v|", prs.sym, prs.symToToken())
	}
	fmt.Println()
}

func TestVSLFiles() {
	root := "samples" // or any other path

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {

		if !info.IsDir() && strings.Contains(path, ".vsl") {
			fmt.Printf("----------%s\n", path)

			content, _ := os.ReadFile(path)
			fmt.Printf("%s\n", string(content))
			parser := NewParser(string(content))
			for parser.getsym() != tkSNULL {
				if parser.err {
					fmt.Printf("**********  %v %v| ", parser.sym, parser.err_message)
					break
				}
				fmt.Printf("%v %v| ", parser.sym, parser.symToToken())
			}
			fmt.Println()
		}
		return nil
	})

	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", root, err)
	}
}

func writeWav(path string, nChannels int, sampleRate int, buff []float32) {
	file, err := os.Create(strings.Replace(path, ".vsl", ".wav", 1))
	if err != nil {
		fmt.Printf("failed to create file: %v", err)
	}
	defer file.Close()

	// Create a new wave writer.
	writer, err := wave.NewWriter(file, wave.SampleTypeFloat32, uint32(sampleRate), wave.WithChannelCount(uint16(nChannels)))
	if err != nil {
		log.Fatalf("failed to create wave writer: %v", err)
	}
	defer writer.Flush()

	if err := writer.WriteFloat32(buff); err != nil {
		log.Fatalf("failed to write data: %v", err)
	}
}

func TestCompileVSLFiles() {
	root := "samples" // vsl file path

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {

		if !info.IsDir() && strings.Contains(path, ".vsl") {
			fmt.Printf("%s\n", path)

			nChannels, sampleRate, buff, err := GenerateVSLFile(path)
			if err != nil {
				fmt.Printf("error compiling %s: %v\n", path, err)
				return nil
			}
			if false {
				writeWav(path, nChannels, sampleRate, buff)
			}

		}
		return nil
	})

	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", root, err)
	}
}

func TestCompiler() {
	expr := `
		// vsl code: default-02.vsl

		const volume=0.2, seconds=15,
		f0=440, f1=Ø·f0, 
		delta=2, f01=f0+delta, f11=Øf01;

		let w0=f0/(t+[3]), w1=f01/(t+[3]);

		~f0   ∿(w0 + ☯f1);
		~f01  ∿(w1 + ☯f11);
	`

	vsl := NewVSLCompiler(expr)
	fmt.Printf("code size: %d\n", vsl.compiler.pc)
	fmt.Println(vsl.compiler.code[:vsl.compiler.pc])
	for _, v := range vsl.compiler.tabValues {
		fmt.Printf("%15v %v %v %v %v %v\n", v.id, v.types, v.di, v.address, v.param_ix, v.n_params)
	}

	buff := vsl.GenerateWave()
	fmt.Println(buff[0:100])
}

func TestPlayVSL() {

	vslFile := "samples/_flwr01.vsl"
	expr, err := os.ReadFile(vslFile)

	if err != nil {
		fmt.Printf("error reading file %s\n", vslFile)
		return
	}

	vsl := NewVSLCompiler(string(expr))

	if vsl.compiler.err {
		fmt.Println(vsl.compiler.ErrorMsg())
		return
	}

	fmt.Printf("file \"%s\" ok, sample Rate:%.0f, seconds:%f, channels:%d\n", vslFile, vsl.sampleRate, vsl.seconds, vsl.channels)

	c, err := pulse.NewClient()
	if err != nil {
		panic(err)
	}
	defer c.Close()

	time.Sleep(time.Millisecond * 100) // warmup

	channel_play := pulse.PlaybackStereo
	if vsl.channels == 1 {
		channel_play = pulse.PlaybackMono
	}

	stream, err := c.NewPlayback(
		pulse.Float32Reader(func(outBuff []float32) (int, error) {
			sampBuff := vsl.generateBuffer(len(outBuff))
			if vsl.sampleCount == 0 { // end of playback?
				fmt.Printf("** reached end of stream **\n")
				return 0, pulse.EndOfData
			}
			for i := range outBuff {
				outBuff[i] = sampBuff[i]
			}
			return len(outBuff), nil
		}),
		pulse.PlaybackLatency(0.1),
		channel_play,
		pulse.PlaybackSampleRate(int(vsl.sampleRate)),
		pulse.PlaybackBufferSize(1024))

	if err != nil {
		fmt.Println(err)
		return
	}
	defer stream.Close()

	// t0 := time.Now()

	stream.Start()
	stream.Drain()

	if stream.Error() != nil {
		fmt.Println("Error:", stream.Error())
	}

	for stream.Running() {
		time.Sleep(200 * time.Millisecond)
	}

	if stream.Underflow() {
		fmt.Println("Underflow:", stream.Underflow())
	}
	// fmt.Printf("end of playback %.2f secs, lap: %v, played %.0f samples, %.2f seconds from samples played\n", vsl.seconds, time.Since(t0), vsl.sampleCount/vsl.channels, vsl.sampleCount/vsl.sampleRate/vsl.channels)
}
