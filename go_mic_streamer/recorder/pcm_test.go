package pcm

import (
	"fmt"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestStart(t *testing.T) {
	t.Run("Start function should exit when receive a signal", func(t *testing.T) {
		sig := make(chan os.Signal, 1)
		filePathCh := make(chan string, 1)
		var wait sync.WaitGroup

		mockPortAudio := &MockPortAudio{}
		baseDir := time.Now().Format("test_audio_20060102_T150405")
		pr := NewPCMRecorder(mockPortAudio, baseDir, 3)

		wait.Add(1)

		go func() {
			time.Sleep(100 * time.Millisecond)
			sig <- os.Interrupt
		}()

		err := pr.Start(sig, filePathCh, &wait)
		if err != nil {
			t.Errorf("got %v, want nil", err)
		}

		wait.Wait()
	})
}

func TestDetectSilence(t *testing.T) {
	t.Run("All array items are 0", func(t *testing.T) {
		input := []int16{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

		interval := 3
		mockPortAudio := &MockPortAudio{}
		baseDir := time.Now().Format("test_audio_20060102_T150405")
		pr := NewPCMRecorder(mockPortAudio, baseDir, interval)

		got := pr.detectSilence(input)
		want := true

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("Some voices are streamed", func(t *testing.T) {
		input := []int16{0, 0, 0, 120, 120, 44, 66, 10, -12, 0, 0, 0, 0, 0, 0, 0}

		interval := 3
		mockPortAudio := &MockPortAudio{}
		baseDir := time.Now().Format("test_audio_20060102_T150405")
		pr := NewPCMRecorder(mockPortAudio, baseDir, interval)

		got := pr.detectSilence(input)
		want := false

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}

func TestDetectSpeechStopped(t *testing.T) {
	t.Run("Should return true when speech is stopped", func(t *testing.T) {
		interval := 3
		mockPortAudio := &MockPortAudio{}
		baseDir := time.Now().Format("test_audio_20060102_T150405")
		pr := NewPCMRecorder(mockPortAudio, baseDir, interval)
		want := true

		contents := make([]int16, 64)
		// silece after some speech should be recognized as 'stop'
		for i := 0; i < 10; i++ {
			contents[i] = 1
		}
		pr.BufferedContents = contents
		pr.silentCount = 51

		got := pr.detectSpeechStopped()

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("Should return false when speech continue", func(t *testing.T) {
		interval := 3
		mockPortAudio := &MockPortAudio{}
		baseDir := time.Now().Format("test_audio_20060102_T150405")
		pr := NewPCMRecorder(mockPortAudio, baseDir, interval)
		want := false

		contents := make([]int16, 64)
		for i := 0; i < len(contents); i++ {
			contents[i] = 1
		}
		pr.BufferedContents = contents
		pr.silentCount = 0

		got := pr.detectSpeechStopped()

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("Should return false when silence was very short", func(t *testing.T) {
		interval := 3
		mockPortAudio := &MockPortAudio{}
		baseDir := time.Now().Format("test_audio_20060102_T150405")
		pr := NewPCMRecorder(mockPortAudio, baseDir, interval)
		want := false

		contents := make([]int16, 64)
		for i := 0; i < len(contents)-10; i++ {
			contents[i] = 1
		}
		pr.BufferedContents = contents
		pr.silentCount = 10

		got := pr.detectSpeechStopped()

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}

func TestDetectSpeechExceededLimitation(t *testing.T) {
	t.Run("Should return true when speech duration is over an interval", func(t *testing.T) {
		interval := 3
		mockPortAudio := &MockPortAudio{}
		baseDir := time.Now().Format("test_audio_20060102_T150405")
		pr := NewPCMRecorder(mockPortAudio, baseDir, interval)
		want := true

		pr.BufferedContents = make([]int16, 44100*pr.Interval)
		got := pr.detectSpeechExceededLimitation()

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("Should return false when speech duration is not over an interval", func(t *testing.T) {
		interval := 3
		mockPortAudio := &MockPortAudio{}
		baseDir := time.Now().Format("test_audio_20060102_T150405")
		pr := NewPCMRecorder(mockPortAudio, baseDir, interval)
		want := false

		pr.BufferedContents = make([]int16, 44100*pr.Interval-1)
		got := pr.detectSpeechExceededLimitation()

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}

func TestRecord(t *testing.T) {
	t.Run("Should append a new input", func(t *testing.T) {
		interval := 3
		mockPortAudio := &MockPortAudio{}
		baseDir := time.Now().Format("test_audio_20060102_T150405")
		pr := NewPCMRecorder(mockPortAudio, baseDir, interval)
		want := []int16{0, 0, 0, 120, 120, 44, 66, 10, -12, 0, 0, 0, 0, 0, 0, 0}

		pr.record(want, time.Now().Sub(time.Now()))

		got := pr.BufferedContents
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

}

type MockPortAudioStream struct{}

func (*MockPortAudioStream) Close() error {
	fmt.Println("Close")

	return nil
}

func (*MockPortAudioStream) Read() error {
	fmt.Println("Read")

	return nil
}

func (*MockPortAudioStream) Start() error {
	fmt.Println("Start")

	return nil
}

func (*MockPortAudioStream) Stop() error {
	fmt.Println("Stop")

	return nil
}

func (*MockPortAudioStream) Time() time.Duration {
	return time.Now().Sub(time.Now())
}

type MockPortAudio struct{}

func (*MockPortAudio) Initialize() error {
	return nil
}

func (*MockPortAudio) Terminate() error {
	return nil
}

func (*MockPortAudio) GetDeviceInfo() {
	return
}

func (*MockPortAudio) OpenDefaultStream(numInputChannels int, numOutputChannels int, sampleRate float64, framesPerBuffer int, args ...interface{}) (AudioSystemStream, error) {
	return &MockPortAudioStream{}, nil
}
