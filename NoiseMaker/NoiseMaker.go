package NoiseMaker

import (
	"fmt"
	"math"
	"sync"

	"github.com/hajimehoshi/oto"
)

type Audio struct {
	m_bReady        bool
	m_nSampleRate   int
	m_nChannels     int
	m_nBlockCount   int
	m_nBlockSamples int
	m_nBlockCurrent int
	m_pBlockMemory  [][]float32
	m_userFunction  func(float64) float64
	thread          *sync.WaitGroup
	m_dGlobalTime   float64
	ctx             *oto.Context
	player          *oto.Player
	blockChan		chan []float32
}

func (a *Audio) Create(nSampleRate int, nChannels int, nBlocks int, nBlockSamples int) bool {
	a.m_bReady = false
	a.m_nSampleRate = nSampleRate
	a.m_nChannels = nChannels
	a.m_nBlockCount = nBlocks
	a.m_nBlockSamples = nBlockSamples
	a.m_nBlockCurrent = 0
	a.m_pBlockMemory = make([][]float32, nBlocks)
	for i := range a.m_pBlockMemory {
		a.m_pBlockMemory[i] = make([]float32, nBlockSamples)
	}
	a.m_userFunction = nil
	a.blockChan = make(chan []float32, nBlocks)

	// Initialize Oto context for audio playback
	var err error
	a.ctx, err = oto.NewContext(nSampleRate, nChannels, 2, nBlockSamples*nChannels*2)
	if err != nil {
		fmt.Println("Failed to create Oto context:", err)
		return false
	}
	a.player = a.ctx.NewPlayer()
	a.thread = &sync.WaitGroup{}
	a.thread.Add(2)
	go a.MainThread()
	go a.PlayThread()

	a.m_bReady = true
	return true
}

func (a *Audio) Destroy() bool {
	a.Stop()
	return true
}

func (a *Audio) Stop() {
	a.m_bReady = false
	close(a.blockChan)
	if a.thread != nil {
		a.thread.Wait()
	}
	if a.player != nil {
		a.player.Close()
	}
	if a.ctx != nil {
		a.ctx.Close()
	}
}

func (a *Audio) UserProcess(dTime float64) float64 {
	return 0.0
}

func (a *Audio) GetTime() float64 {
	return a.m_dGlobalTime
}

func (a *Audio) SetUserFunction(f func(float64) float64) {
	a.m_userFunction = f
}

func (a *Audio) MainThread() {
	defer a.thread.Done()
	a.m_dGlobalTime = 0.0
	dTimeStep := 1.0 / float64(a.m_nSampleRate)

	for a.m_bReady {
		block := a.m_pBlockMemory[a.m_nBlockCurrent]
		for i := range block {
			var sample float32
			if a.m_userFunction != nil {
				sample = float32(a.clip(a.m_userFunction(a.m_dGlobalTime), 1.0))
			} else {
				sample = float32(a.clip(a.UserProcess(a.m_dGlobalTime), 1.0))
			}
			block[i] = sample
			a.m_dGlobalTime += dTimeStep
		}
		a.blockChan <- block
		a.m_nBlockCurrent = (a.m_nBlockCurrent + 1) % a.m_nBlockCount
	}
}

func (a *Audio) PlayThread() {
	defer a.thread.Done()
	for block := range a.blockChan {
		a.player.Write(convertToByteArray(block))
	}
}

func convertToByteArray(samples []float32) []byte {
	buf := make([]byte, len(samples)*2)
	for i, s := range samples {
		v := int16(s * 32767)
		buf[i*2] = byte(v)
		buf[i*2+1] = byte(v >> 8)
	}
	return buf
}

func (a *Audio) clip(dSample float64, dMax float64) float64 {
	return math.Max(-dMax, math.Min(dSample, dMax))
}

func NewAudio(nSampleRate int, nChannels int, nBlocks int, nBlockSamples int) *Audio {
	audio := &Audio{}
	audio.Create(nSampleRate, nChannels, nBlocks, nBlockSamples)
	return audio
}
