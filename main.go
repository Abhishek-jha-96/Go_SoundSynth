package main

import (
	"fmt"
	"log"
	"math"
	"github.com/Abhishek-jha-96/Go_SoundSynth/NoiseMaker"
	"time"

	"github.com/eiannone/keyboard"
)

// Global variables
var dFrequencyOutput float64 = 0.0
var dOctaveBaseFrequency float64 = 110.0
var d12thRootOf2 float64 = math.Pow(2.0, 1.0/12.0)

// function used to generate sound waves
func MakeNoise(dTime float64) float64 {
	dOutput := math.Sin(dFrequencyOutput * 2.0 * math.Pi * dTime)
	return dOutput * 0.5
}

func main() {
	audioInstance := NoiseMaker.NewAudio(44100, 1, 8, 512)
	fmt.Println("Audio initialized")

	audioInstance.SetUserFunction(MakeNoise)

	// Create a new keyboard input instance
	err := keyboard.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer keyboard.Close()

	var nCurrentKey int = -1
	var bKeyPressed bool
	done := make(chan bool)

	go func() {
		defer close(done)
		for {
			bKeyPressed = false
			char, _, err := keyboard.GetKey()
			if err != nil {
				// If error is due to keyboard closing, exit the goroutine
				if err.Error() == "operation canceled" {
					return
				}
				log.Fatal(err)
			}

			for i, k := range "ZSXCFVGBNJMK\xbcL\xbe\xbf" {
				if rune(char) == k {
					if nCurrentKey != i {
						dFrequencyOutput = dOctaveBaseFrequency * math.Pow(d12thRootOf2, float64(i))
						fmt.Printf("\rNote On: %f Hz", dFrequencyOutput)
						nCurrentKey = i
					}
					bKeyPressed = true
					break
				}
			}

			if !bKeyPressed {
				if nCurrentKey != -1 {
					fmt.Printf("\rNote Off                          ")
					nCurrentKey = -1
				}
				dFrequencyOutput = 0.0
			}

			time.Sleep(16 * time.Millisecond) // Sleep to prevent busy looping
		}
	}()

	// Run for 5 seconds to demonstrate functionality
	time.Sleep(5 * time.Second)

	// Stop the audio and close keyboard
	audioInstance.Stop()
	fmt.Println("\nAudio stopped")

	// Close the keyboard to stop the goroutine
	keyboard.Close()

	// Wait for the goroutine to finish
	<-done

	fmt.Println("Global Time:", audioInstance.GetTime())
	fmt.Println("User Process:", audioInstance.UserProcess(0.0))
}
