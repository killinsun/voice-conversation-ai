package player

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/gordonklaus/portaudio"
)

type VoiceVoxTextToSpeechResponse struct {
	AudioQuery map[string]interface{} `json:"audio_query"`
}

func Say(text string) error {
	params := url.Values{}
	params.Add("text", text)
	params.Add("speaker", "1")

	// audio_queryリクエスト
	audioQueryResp, err := http.PostForm("http://localhost:50021/audio_query", params)
	if err != nil {
		return err
	}
	defer audioQueryResp.Body.Close()

	var vvttsResp VoiceVoxTextToSpeechResponse
	if err := json.NewDecoder(audioQueryResp.Body).Decode(&vvttsResp); err != nil {
		return err
	}

	// synthesisリクエスト
	synthesisBody, err := json.Marshal(vvttsResp.AudioQuery)
	if err != nil {
		return err
	}

	synthesisReq, err := http.NewRequest("POST", "http://localhost:50021/synthesis", bytes.NewBuffer(synthesisBody))
	if err != nil {
		return err
	}

	synthesisReq.Header.Set("Content-Type", "application/json")
	synthesisReq.URL.RawQuery = params.Encode()

	synthesisResp, err := http.DefaultClient.Do(synthesisReq)
	if err != nil {
		return err
	}
	defer synthesisResp.Body.Close()

	voice_audio, _ := ioutil.ReadAll(synthesisResp.Body)

	return playVoice(voice_audio)
}

func playVoice(voiceBytes []byte) error {
	// portaudioの初期化
	portaudio.Initialize()
	defer portaudio.Terminate()

	// オーディオストリームを開く
	stream, err := portaudio.OpenDefaultStream(1, 0, 24000, len(voiceBytes), func(out []int32) {
		for i := range out {
			if len(voiceBytes) > 0 {
				out[i] = int32(voiceBytes[0])
				voiceBytes = voiceBytes[1:]
			}
		}
	})
	if err != nil {
		return err
	}
	defer stream.Close()

	// オーディオストリームの開始
	if err := stream.Start(); err != nil {
		return err
	}
	defer stream.Stop()

	// オーディオデータの供給が終わるまで待機
	for len(voiceBytes) > 0 {
	}

	return nil
}
