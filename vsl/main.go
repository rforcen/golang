package main

import (
	"fmt"
	"math"
	"time"
	
	"github.com/jfreymuth/pulse"
	"vsl/vsl"
)

type Wave struct {
	t, tStamp, inc, secs, chans, sampleRate, amp, sampCnt float32
}

func newWave(secs float32, sampleRate int, amp float32, chans int) *Wave {
	return &Wave{
		secs:       secs,
		chans:      float32(chans),
		sampleRate: float32(sampleRate),
		amp:        amp,
		t:          0,
		tStamp:     0,
		inc:        2. * math.Pi / float32(sampleRate),
		sampCnt:    0,
	}
}

var wave *Wave

func testWaveOpenAudio() {
	c, err := pulse.NewClient()
	if err != nil {
		panic(err)
	}
	defer c.Close()

	wave = newWave(8., 44100, .1, 2) // create wave
	fmt.Printf("pulse audio test: %f secs\n", wave.secs)

	time.Sleep(time.Millisecond * 100) // warmup

	stream, err := c.NewPlayback(
		pulse.Float32Reader(func(outBuff []float32) (int, error) {
			wave.tStamp += float32(len(outBuff)) / wave.sampleRate / wave.chans
			wave.sampCnt += float32(len(outBuff))

			for i := 0; i < len(outBuff); i += int(wave.chans) {
				if wave.tStamp >= wave.secs { // end of playback?
					fmt.Printf("tStamp : %.2f ** reached end of stream **\n", wave.tStamp)
					return i, pulse.EndOfData
				}
				// generate audio
				wave_sin := func(f float32) float32 {
					return wave.amp * float32(math.Sin(float64(f*wave.t)))
				}
				outBuff[i+0] = wave_sin(400)
				outBuff[i+1] = wave_sin(404)

				wave.t += wave.inc
			}

			return len(outBuff), nil
		}),
		pulse.PlaybackLatency(0.1),
		pulse.PlaybackStereo,
		pulse.PlaybackSampleRate(44100),
		pulse.PlaybackBufferSize(1024))

	if err != nil {
		fmt.Println(err)
		return
	}
	defer stream.Close()

	t0 := time.Now()

	stream.Start()
	stream.Drain()

	if stream.Error() != nil {
		fmt.Println("Error:", stream.Error())
	}

	for stream.Running() {
		time.Sleep(200 * time.Millisecond)
		fmt.Printf("lap: %.1f, samples: %.0f\r", wave.tStamp, wave.sampCnt/wave.chans)
	}

	if stream.Underflow() {
		fmt.Println("Underflow:", stream.Underflow())
	}
	fmt.Printf("end of playback %.2f secs, lap: %v, played %.0f samples, %.2f seconds from samples played\n", wave.secs, time.Since(t0), wave.sampCnt/wave.chans, wave.sampCnt/wave.sampleRate/wave.chans)
}

func TerstCompiler() {
	// vsl.TestParser()
	vsl.TestPlayVSL()
	// vsl.TestVSLFiles()
}

func main() {
	vsl.TestPlayVSL()

}
