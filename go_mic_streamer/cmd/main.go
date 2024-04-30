package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"golang.org/x/net/websocket"

	pcm "github.com/killinsun/voice-conversation-ai/go_mic_streamer/recorder"
)

type MediaStruct struct {
	Track     string `json:"track"`
	Chunk     int    `json:"chunk"`
	Timestamp int    `json:"timestamp"`
	Payload   string `json:"payload"`
}

type MediaStreamStruct struct {
	Event          string      `json:"event"`
	SequenceNumber int         `json:"sequenceNumber"`
	Media          MediaStruct `json:"media"`
	StreamSid      string      `json:"streamSid"`
}

func main() {
	log.SetFlags(log.Lmicroseconds)

	ws, dialErr := websocket.Dial("ws://localhost:8000/ws_test", "", "http://localhost:8000")
	if dialErr != nil {
		log.Fatal(dialErr)
	}
	defer ws.Close()

	baseDir := time.Now().Format("audio_20060102_T150405")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		log.Fatal("Could not create a new directory")
	}

	audioSystem := &pcm.PortAudioSystem{}
	pr := pcm.NewPCMRecorder(audioSystem, fmt.Sprintf(baseDir+"/file"), 30, 100)

	pr.GetDeviceInfo()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	filePathCh := make(chan string)

	var wait sync.WaitGroup
	wait.Add(1)
	go func() {
		if err := pr.Start(sig, filePathCh, &wait); err != nil {
			log.Fatalf("Error starting PCMRecorder: %v", err)
		}
	}()

	go func() {
		for {
			filePath, ok := <-filePathCh
			if !ok {
				break
			}
			b, err := ioutil.ReadFile(filePath)
			if err != nil {
				log.Fatal(err)
			}
			b64EncodedWav := base64.StdEncoding.EncodeToString(b)

			if err := sendMediaStream(ws, b64EncodedWav); err != nil {
				log.Fatal(err)
			}
		}
	}()

	// WebSocketからテキストを受信する処理を追加
	go func() {
		var msg = make([]byte, 512)
		for {
			n, err := ws.Read(msg)
			if err != nil {
				log.Println("Error reading from WebSocket:", err)
				continue
			}
			receivedText := string(msg[:n])
			log.Println("AI:", receivedText)

			// 受信したテキストを処理する
			// log.Println("starting Say")
			// err = player.Say(receivedText)
			// if err != nil {
			// 	log.Fatal("Error!", err)
			// }
		}
	}()

	<-sig
	wait.Wait()
}

func sendMediaStream(ws *websocket.Conn, payload string) error {
	media := MediaStruct{
		"inbound",
		1,
		300,
		payload,
	}
	mediaStream := MediaStreamStruct{
		"media",
		301,
		media,
		"dummy",
	}

	jsonData, err := json.Marshal(mediaStream)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	sendErr := websocket.Message.Send(ws, string(jsonData))
	if sendErr != nil {
		log.Fatal(sendErr)
	}

	log.Printf("Send: %d to Server", mediaStream.SequenceNumber)

	return nil
}
