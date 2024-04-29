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
	pr := pcm.NewPCMRecorder(audioSystem, fmt.Sprintf(baseDir+"/file"), 30)

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

	log.Println(mediaStream)
	log.Println(jsonData)
	log.Println(string(jsonData))

	log.Printf("Send: %d to Server", mediaStream.SequenceNumber)

	return nil
}

// func main() {
// 	fmt.Println("Streaming. Press Ctrl + C to stop.")

// 	conn, err := grpc.Dial(
// 		"localhost:8080",

// 		grpc.WithTransportCredentials(insecure.NewCredentials()),
// 		grpc.WithBlock(),
// 	)
// 	if err != nil {
// 		log.Fatal("Connection failed.")
// 		return
// 	}
// 	defer conn.Close()
// 	client = transcriptorpb.NewTranscriptorServiceClient(conn)

// 	wavStream, err := client.StreamWav(context.Background())
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	res, err := wavStream.CloseAndRecv()
// 	if err != nil {
// 		log.Fatalf("Error closing and receiving StreamWav: %v", err)
// 	}
// 	fmt.Printf("Done: %v\n", res.GetDone())
// 	fmt.Println("Streaming finished.")
// }
