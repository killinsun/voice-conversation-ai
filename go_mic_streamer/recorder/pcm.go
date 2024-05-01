package recorder

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gordonklaus/portaudio"
)

type AudioSystem interface {
	GetDeviceInfo()
	Initialize() error
	OpenDefaultStream(numInputChannels int, numOutputChannels int, sampleRate float64, framesPerBuffer int, args ...interface{}) (AudioSystemStream, error)
	Terminate() error
}

type AudioSystemStream interface {
	Close() error
	Read() error
	Start() error
	Stop() error
	Time() time.Duration
}

type PortAudioSystem struct{}

func (p *PortAudioSystem) Initialize() error {
	return portaudio.Initialize()
}

func (p *PortAudioSystem) Terminate() error {
	return portaudio.Terminate()
}

func (p *PortAudioSystem) OpenDefaultStream(numInputChannels int, numOutputChannels int, sampleRate float64, framesPerBuffer int, args ...interface{}) (AudioSystemStream, error) {
	stream, err := portaudio.OpenDefaultStream(numInputChannels, numOutputChannels, sampleRate, framesPerBuffer, args...)
	return stream, err
}

func (p *PortAudioSystem) GetDeviceInfo() {
	devices, err := portaudio.Devices()
	if err != nil {
		panic("Could not get devices")
	}

	log.Println("Default device:")
	log.Println("Input Device: " + devices[0].HostApi.DefaultInputDevice.Name)
	log.Println("Output Device: " + devices[0].HostApi.DefaultOutputDevice.Name)
	log.Println("Devices:")
	for i := 0; i < len(devices); i++ {
		log.Printf("Index: %d", i)
		log.Println("Name:" + devices[i].Name)
		log.Println("---------------------")
	}
}

type PCMRecorder struct {
	BaseDir              string
	Interval             int
	SilentRatio          int
	BaseLangCode         string
	AltLangCodes         []string
	BufferedContents     []int16
	Input                []int16
	recognitionStartTime time.Duration
	silentCount          int
	unSilentCount        int
	audioSystem          AudioSystem
}

func NewPCMRecorder(audioSystem AudioSystem, baseDir string, interval int, silentRatio int) *PCMRecorder {
	var pr = &PCMRecorder{
		BaseDir:              baseDir,
		Interval:             interval,
		SilentRatio:          silentRatio,
		recognitionStartTime: -1,
		audioSystem:          audioSystem,
	}
	return pr
}

func (pr *PCMRecorder) GetDeviceInfo() {
	pr.audioSystem.Initialize()
	defer pr.audioSystem.Terminate()
	pr.audioSystem.GetDeviceInfo()
}

func (pr *PCMRecorder) Start(sig chan os.Signal, filepathCh chan string, wait *sync.WaitGroup) error {
	pr.audioSystem.Initialize()
	defer pr.audioSystem.Terminate()

	var err error
	stream, err := pr.initializeAudioStream()

	if err != nil {
		log.Fatalf("Could not open default stream \n %v", err)
	}
	(*stream).Start()
	defer (*stream).Close()

	log.Println("Device initialized.")

loop:
	for {
		select {
		case <-sig:
			wait.Done()
			close(filepathCh)
			break loop
		default:
		}

		pr.processAudioInput(filepathCh, stream)
	}

	return nil
}

func (pr *PCMRecorder) Stop() {
	pr.audioSystem.Terminate()
}

func (pr *PCMRecorder) initializeAudioStream() (*AudioSystemStream, error) {

	pr.Input = make([]int16, 64)
	stream, err := pr.audioSystem.OpenDefaultStream(1, 0, 16000, len(pr.Input), pr.Input)
	return &stream, err
}

func (pr *PCMRecorder) processAudioInput(filePathCh chan string, stream *AudioSystemStream) {
	if err := (*stream).Read(); err != nil {
		log.Fatalf("Could not read stream\n%v", err)
	}

	if pr.detectSilence(pr.Input) {
		pr.silentCount++
	} else {
		pr.silentCount = 0
		pr.unSilentCount++
		pr.record(pr.Input, (*stream).Time())
	}

	if pr.isSpeechLengthEnough() && (pr.detectSpeechStopped() || pr.detectSpeechExceededLimitation()) {
		log.Println("speech stopped or exceeded limitation. Starting finalizing.")
		pr.finalizeRecording(filePathCh)
	}
}

func (pr *PCMRecorder) finalizeRecording(filepathCh chan string) {
	outputFileName := fmt.Sprintf(pr.BaseDir+"_%d.wav", int(pr.recognitionStartTime))
	fmt.Println(outputFileName)
	pr.writePCMData(outputFileName, pr.BufferedContents)
	filepathCh <- outputFileName

	pr.BufferedContents = nil
	pr.silentCount = 0
	pr.unSilentCount = 0
	pr.recognitionStartTime = -1
}

func (pr *PCMRecorder) record(input []int16, startTime time.Duration) {
	if pr.recognitionStartTime == -1 {
		pr.recognitionStartTime = startTime
	}
	pr.BufferedContents = append(pr.BufferedContents, input...)
}

func (pr *PCMRecorder) detectSilence(input []int16) bool {
	// サンプリングした音声データの欠片に threshold 以上のものがあれば無音でないと判断
	silent := true
	threshold := int16(pr.SilentRatio)

	for _, bit := range input {
		if abs(int16(bit)) > threshold {
			silent = false
			break
		}
	}

	return silent
}

func (pr *PCMRecorder) isSpeechLengthEnough() bool {
	return pr.unSilentCount > 100
}

func (pr *PCMRecorder) detectSpeechStopped() bool {
	return len(pr.BufferedContents) > 0 && pr.silentCount > 50
}

func (pr *PCMRecorder) detectSpeechExceededLimitation() bool {
	return len(pr.BufferedContents) >= (16000 * pr.Interval)
}

func (pr *PCMRecorder) writePCMData(outputFileName string, pcmData []int16) {
	if exists(outputFileName) {
		log.Fatalf("The audio file is already exists.")
	}
	file, err := os.Create(outputFileName)
	if err != nil {
		log.Fatalf("Could not create a new file to write \n %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Fatalf("Could not close output file \n %v", err)
		}
	}()

	wav := NewWAVEncoder(file, pcmData)
	wav.Encode()
}

func exists(fileName string) bool {
	_, err := os.Stat(fileName)
	return err == nil
}

func abs(x int16) int16 {
	if x < 0 {
		return -x
	}
	return x
}
