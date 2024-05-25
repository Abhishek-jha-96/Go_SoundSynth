package NoiseMaker

import (
	"fmt"
	"log"
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
	m_nBlockFree    int
	m_nBlockCurrent int
	m_pBlockMemory  []float32
	m_userFunction  func(float64) float64
	muxBlockNotZero sync.Mutex
	cvBlockNotZero  *sync.Cond
	thread          *sync.WaitGroup
	m_dGlobalTime   float64
	ctx             *oto.Context
	player          *oto.Player
}

func (a *Audio) Create(nSampleRate int, nChannels int, nBlocks int, nBlockSamples int) bool {
	a.m_bReady = false
	a.m_nSampleRate = nSampleRate
	a.m_nChannels = nChannels
	a.m_nBlockCount = nBlocks
	a.m_nBlockSamples = nBlockSamples
	a.m_nBlockFree = nBlocks
	a.m_nBlockCurrent = 0
	a.m_pBlockMemory = make([]float32, nBlocks*nBlockSamples)
	a.m_userFunction = nil

	// Initialize Oto context for audio playback
	var err error
	a.ctx, err = oto.NewContext(nSampleRate, nChannels, 2, nBlockSamples*nChannels*2)
	if err != nil {
		fmt.Println("Failed to create Oto context:", err)
		return false
	}
	a.player = a.ctx.NewPlayer()
	a.cvBlockNotZero = sync.NewCond(&a.muxBlockNotZero)
	a.thread = &sync.WaitGroup{}
	a.thread.Add(1)
	go a.MainThread()

	a.m_bReady = true
	a.cvBlockNotZero.Signal()

	return true
}

func (a *Audio) Destroy() bool {
	a.Stop()
	return false
}

func (a *Audio) Stop() {
	a.m_bReady = false
	a.cvBlockNotZero.Signal()
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
	dMaxSample := float32(math.MaxInt16)
	nCurrentBlock := 0

	for a.m_bReady {
		a.muxBlockNotZero.Lock()
		for a.m_nBlockFree == 0 && a.m_bReady {
			a.cvBlockNotZero.Wait()
		}
		a.m_nBlockFree--
		a.muxBlockNotZero.Unlock()

		if !a.m_bReady {
			break
		}

		for i := 0; i < a.m_nBlockSamples; i++ {
			nSample := float32(0.0)
			if a.m_userFunction != nil {
				nSample = float32(a.clip(a.m_userFunction(a.m_dGlobalTime), 1.0) * float64(dMaxSample))
			} else {
				nSample = float32(a.clip(a.UserProcess(a.m_dGlobalTime), 1.0) * float64(dMaxSample))
			}
			a.m_pBlockMemory[nCurrentBlock*a.m_nBlockSamples+i] = nSample
			a.m_dGlobalTime += dTimeStep
		}

		a.player.Write(convertToByteArray(a.m_pBlockMemory[nCurrentBlock*a.m_nBlockSamples : (nCurrentBlock+1)*a.m_nBlockSamples]))
		nCurrentBlock++
		nCurrentBlock %= a.m_nBlockCount
		a.muxBlockNotZero.Lock()
		a.m_nBlockFree++
		a.muxBlockNotZero.Unlock()
	}

	log.Println("MainThread finished")
}

func convertToByteArray(samples []float32) []byte {
	buf := make([]byte, len(samples)*2)
	for i, s := range samples {
		v := int16(s)
		buf[i*2] = byte(v)
		buf[i*2+1] = byte(v >> 8)
	}
	return buf
}

func (a *Audio) clip(dSample float64, dMax float64) float64 {
	if dSample >= 0.0 {
		return math.Min(dSample, dMax)
	}
	return math.Max(dSample, -dMax)
}

func NewAudio(nSampleRate int, nChannels int, nBlocks int, nBlockSamples int) *Audio {
	audio := &Audio{}
	audio.Create(nSampleRate, nChannels, nBlocks, nBlockSamples)
	return audio
}
