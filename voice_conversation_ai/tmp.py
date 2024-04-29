import json
import os
from io import BytesIO

import numpy as np
import pyaudio
import requests
import soundfile as sf
import speech_recognition as sr
import whisper
import ollama
from openai import OpenAI
from dotenv import load_dotenv

load_dotenv()

CHAT_BASE = [
    {"role": "system", "content": "あなたは語尾に「なのだ」をつけて会話します。"},
]

print(os.environ["OPENAI_API_KEY"])


def openai_chat(input: str) -> str:
    client = OpenAI()

    response = client.chat.completions.create(
        model="gpt-3.5-turbo",
        messages=[
            *CHAT_BASE,
            {
                "role": "user",
                "content": input
            }
        ],
        temperature=1,
        max_tokens=256,
        top_p=1,
        frequency_penalty=0,
        presence_penalty=0
    )

    return response.choices[0].message.content


def ollama_chat(input: str) -> str:
    USE_MODEL = "llama3"
    result = ollama.chat(
        model=USE_MODEL,
        messages=[
            *CHAT_BASE,
            {"role": "user", "content": input},
        ],
        options={
            # reference: https://github.com/ollama/ollama/blob/main/docs/api.md
            "top_k": 1,
            "temperature": 0.01,
            "top_p": 0.9,
            "num_predict": 5,
            "penalize_newline": True,
        },
    )
    return result["message"]["content"]


def vvox_t2s(text: str):
    params = (
        ('text', text),
        ('speaker', 1)
    )

    query = requests.post(
        f"http://localhost:50021/audio_query",
        params=params
    )

    synthesis = requests.post(
        f"http://localhost:50021/synthesis",
        headers={"Content-Type": "application/json"},
        params=params,
        data=json.dumps(query.json())
    )

    voice = synthesis.content
    pya = pyaudio.PyAudio()

    stream = pya.open(format=pyaudio.paInt16,
                      channels=1,
                      rate=24000,
                      output=True)

    stream.write(voice)
    stream.stop_stream()
    stream.close()
    pya.terminate()


if __name__ == "__main__":
    recognizer = sr.Recognizer()
    while True:
        # 「マイクから音声を取得」参照
        with sr.Microphone(sample_rate=16_000) as source:
            print(">")
            recognizer.pause_threshold = 0.5
            audio = recognizer.listen(source)

        print("...")
        # 「音声データをWhisperの入力形式に変換」参照
        wav_bytes = audio.get_wav_data()
        wav_stream = BytesIO(wav_bytes)
        audio_array, sampling_rate = sf.read(wav_stream)
        audio_fp32 = audio_array.astype(np.float32)

        result = whisper.transcribe(
            audio_fp32,
            path_or_hf_repo="mlx-community/whisper-tiny-mlx",
            initial_prompt="こんにちは、元気ですか？ありがとうございました。"  # 句読点形式で文字起こしさせるため

        )

        print(result["text"])
        vvox_t2s(text="なるほど、なのだ")
        # model_answer = ollama_chat(input=result["text"])
        model_answer = openai_chat(input=result["text"])
        print(model_answer)

        vvox_t2s(text=model_answer)
