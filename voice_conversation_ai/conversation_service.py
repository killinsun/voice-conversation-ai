import json
from abc import ABC
from typing import Any

import pyaudio
import requests

from voice_conversation_ai.llm.llm_base import ILLM, CompletionOptions
from voice_conversation_ai.transcriber import Transcriber


class ConversationResponse(ABC):
    def say(self, text: str) -> Any:
        pass


class VoiceVoxTextToSpeechResponse(ConversationResponse):
    def say(self, text: str) -> None:
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


class SimpleTextDisplayResponse(ConversationResponse):
    def say(self, text: str) -> None:
        print(text)



class ConversationService:
    def __init__(self, llm: ILLM):
        self.llm = llm

    def correct_dirty_transcript(self, input: str, resp: ConversationResponse) -> str:
        system_prompt = """
        あなたは、高精度な文章修正システムです。音声から文字起こしした文章の会話文脈を、自然な形になるように修正します。
        修正後の文章だけを出力します。文章を削ることなく、可能な限り構成し、無理なところはそのまま出力してください。"""
        messages = [
            {"role": "system", "content": system_prompt},
            {"role": "user", "content": input},
        ]
        options = CompletionOptions()

        corrected_text = self.llm.create_completion(messages=messages, options=options)
        print(f"original text: {input} \n corrected text: {corrected_text}")

        return corrected_text

    def chat_with_ai_operator(self, input: str, resp: ConversationResponse) -> str:
        system_prompt = """
        あなたは、AI受付システムです。ユーザーからの文章を元に、AIが返答します。
        """
        messages = [
            {"role": "system", "content": system_prompt},
            {"role": "user", "content": input},
        ]
        options = CompletionOptions()
        answer = self.llm.create_completion(messages=messages, options=options)

        return resp.say(answer)
