package main

// import (
// 	"math"
// 	"github.com/Abhishek-jha-96/Go_SoundSynth/NoiseMaker"
// 	"time"
// )

import (
	"fmt"
	"log"
	"math"
	"github.com/Abhishek-jha-96/Go_SoundSynth/NoiseMaker"
	"time"
	"math/rand"

	"github.com/eiannone/keyboard"
)

// frequency to angular velocity
func w(dHertz float64) float64 {
	return dHertz * 2.0 * math.Pi;
}
// General Purpose Oscillator
const (
	OSC_SINE     int = 0
	OSC_SQUARE   int = 1
	OSC_TRIANGLE int = 2
	OSC_SAW_ANA  int = 3
	OSC_SAW_DIG  int = 4
	OSC_NOISE    int = 5
)

//Oscillator
func ocs(dHertz float64, dTime float64, nType int) float64 {
	switch nType {
	case OSC_SINE: return math.Sin(w(dHertz) * dTime) //Sine waveform
	case OSC_SQUARE: return math.Sin(w(dHertz) * dTime) //Square waveform
	case OSC_TRIANGLE: return math.Asin(math.Sin(w(dHertz) * dTime) * (2.0 / math.Pi)) //Triangle waveform
	case OSC_SAW_ANA:
		{
			var dOutput float64 = 0.0
			for i := 1.0; i < 40.0; i++ {
				dOutput += (math.Sin(i * w(dHertz) * dTime)) / i 
			}
			return dOutput * (2.0 / math.Pi)
		}
	case OSC_SAW_DIG:
		return (2.0 /math.Pi) * (dHertz * math.Pi * math.Mod(dTime, 1.0 / dHertz) - (math.Pi / 2.0))	
	case OSC_NOISE:
		source := rand.NewSource(time.Now().UnixNano())
		r := rand.New(source)
		return 2.0*r.Float64() - 1.0
	default:
		return 0.0	
	}
} 

// Amplitude (attack, Decay, Sustain, Release) Envelope
type sEnvelopeADSR struct {
	dAttackTime float64
	dDecayTime float64
	dSustainAmplitude float64
	dReleaseTime float64
	dStartAmplitude float64
	dTriggerOffTime float64
	dTriggerOnTime float64
	bNoteOn bool
}

// Constructor for sEnvelopeADSR
func NewEnvelopeADSR() *sEnvelopeADSR {
	return &sEnvelopeADSR{
		dAttackTime:       0.10,
		dDecayTime:        0.01,
		dStartAmplitude:   1.0,
		dSustainAmplitude: 0.8,
		dReleaseTime:      0.20,
		bNoteOn:           false,
		dTriggerOffTime:   0.0,
		dTriggerOnTime:    0.0,
	}
}

// NoteOn method
func (env *sEnvelopeADSR) NoteOn(dTimeOn float64) {
	env.dTriggerOnTime = dTimeOn
	env.bNoteOn = true
}	

// NoteOff method
func (env *sEnvelopeADSR) NoteOff(dTimeOff float64) {
	env.dTriggerOffTime = dTimeOff
	env.bNoteOn = false
} 

//GetAmplitude method
func (env *sEnvelopeADSR) GetAmplitude(dTime float64) float64 {
	var dAmplitude float64 = 0.0
	var dLifeTime float64 = dTime - env.dTriggerOnTime

	if env.bNoteOn {
		if dLifeTime <= env.dAttackTime {
			dAmplitude = (dLifeTime / env.dAttackTime) * env.dStartAmplitude
		}
		if dLifeTime > env.dAttackTime && dLifeTime <= (env.dAttackTime + env.dDecayTime) {
			dAmplitude = ((dLifeTime - env.dAttackTime) / env.dDecayTime) * (env.dSustainAmplitude - env.dStartAmplitude) + env.dStartAmplitude
		}
		if dLifeTime > (env.dAttackTime + env.dDecayTime) {
			dAmplitude = env.dSustainAmplitude
		}
	} else {
		dAmplitude = ((dTime - env.dTriggerOffTime) / env.dReleaseTime) * (0.0 - env.dSustainAmplitude) + env.dSustainAmplitude
	}

	if (dAmplitude <= 0.0001) {
		dAmplitude = 0.0
	}

	return dAmplitude
}

// Global variables
var dFrequencyOutput float64 = 0.0
var dOctaveBaseFrequency float64 = 110.0
var d12thRootOf2 float64 = math.Pow(2.0, 1.0/12.0)

// function used to generate sound waves
func (env *sEnvelopeADSR) MakeNoise (dTime float64) float64 {
	dOutput := env.GetAmplitude(dTime) * (1.0*ocs(dFrequencyOutput*0.5, dTime, OSC_SINE) + 1.0*ocs(dFrequencyOutput, dTime, OSC_SAW_ANA))
	return dOutput * 0.4
}

func main() {
    audioInstance := NoiseMaker.NewAudio(44100, 1, 8, 512)
    fmt.Println("Audio initialized")

    envelope := NewEnvelopeADSR()
    audioInstance.SetUserFunction(envelope.MakeNoise)

    // Keyboard setup (as before)
    err := keyboard.Open()
    if err != nil {
        log.Fatal(err)
    }
    defer keyboard.Close()

    var nCurrentKey int = -1
    var bKeyPressed bool

    fmt.Println("Press keys to play notes. Press 'q' to quit.")
    for {
        char, _, err := keyboard.GetKey()
        if err != nil {
            log.Fatal(err)
        }

        if char == 'q' {
            break
        }

        bKeyPressed = false
        for i, k := range "ZSXCFVGBNJMK\xbcL\xbe\xbf" {
            if rune(char) == k {
                if nCurrentKey != i {
                    dFrequencyOutput = dOctaveBaseFrequency * math.Pow(d12thRootOf2, float64(i))
                    fmt.Printf("\rNote On: %f Hz", dFrequencyOutput)
                    nCurrentKey = i
                    envelope.NoteOn(audioInstance.GetTime())
                }
                bKeyPressed = true
                break
            }
        }

        if !bKeyPressed {
            if nCurrentKey != -1 {
                fmt.Printf("\rNote Off    ")
                nCurrentKey = -1
                envelope.NoteOff(audioInstance.GetTime())
            }
            dFrequencyOutput = 0.0
        }

        // Print debug info
        fmt.Printf("\nCurrent time: %f, Frequency: %f\n", audioInstance.GetTime(), dFrequencyOutput)
    }

    audioInstance.Stop()
    fmt.Println("\nAudio stopped")

    fmt.Println("Global Time:", audioInstance.GetTime())
}

/* func main() {
	audio := NoiseMaker.NewAudio(44100, 1, 8, 512)
	audio.SetUserFunction(func(t float64) float64 {
		return math.Sin(2 * math.Pi * 440 * t) // 440 Hz sine wave
	})

	// Let it play for a while
	time.Sleep(5 * time.Second)

	audio.Stop()
} */