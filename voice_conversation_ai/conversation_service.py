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
    def say(self, text: str) -> str:
        print(text)
        return text


class ConversationService:
    def __init__(self, llm: ILLM):
        self.llm = llm

    def revise_dirty_text(self, input: str, conversation_logs: list[dict[str, str]]) -> str:
        system_prompt = """
        あなたは、高精度な文章修正システムです。
        あなたは、ユーザーが電話で発話した音声データを、AIによって文字起こししたテキストを受け取ります。
        電話応対をしているとイメージして、そのテキストの文脈に合うように修正してください。
        あなたが修正するのは、ユーザーの発話テキストのみです。
        
        # example
        input: 田中さんって言いますか？
        output: 田中さんっていますか？
        
        input: お世話しております
        output: お世話になっております
        
        input: ゼロハチゼロ
        output: 080
        """
        messages = [
            {"role": "system", "content": system_prompt},
            {"role": "user", "content": input},
        ]
        options = CompletionOptions()

        revised_text = self.llm.create_completion(messages=messages, options=options)
        print(f"original: {input} -> revised: {revised_text}")

        return revised_text

    def chat_with_ai_operator(self, conversation_logs: list[dict[str, str]]) -> str:
        system_prompt = """
        あなたは、首無商事株式会社の電話応対システムです。
        お客様からのお問い合わせに対して、電話で回答をしているBOTです。
        要件を聞いて、適切な回答をしてください。
        電話対応は以下のように行います。
        1. 宛先と要件を確認する（「ご用件をお伺いします」）
        2. 担当から折り返すと伝える。折り返しのために、再度お客様の名前と、電話番号を伺う。
         例: お名前と電話番号をお願いします。
            user: 田中太郎です。
            you: 田中太郎様ですね。お電話番号をお伺いします。
            user: 080
            assistant: 080
            user: 1234
            assistant: 1234
            user: 5678
            assistant: 5678
            user: それです
            assistant: 080-1234-5678ですね。折り返しの担当者が折り返しの電話をさせていただきます。
        3. 伺った情報を復唱して確認する。
        
        注意：電話番号は、11桁または10桁です。
        """
        messages = [
            {"role": "system", "content": system_prompt},
            *conversation_logs,
        ]
        options = CompletionOptions()
        answer = self.llm.create_completion(messages=messages, options=options)

        return answer
