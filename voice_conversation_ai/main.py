import base64

from fastapi import FastAPI, WebSocket

from voice_conversation_ai.conversation_service import ConversationService, VoiceVoxTextToSpeechResponse, \
    SimpleTextDisplayResponse
from voice_conversation_ai.llm.ollama_llm import OllamaLLM
from voice_conversation_ai.llm.openai_llm import OpenAILLM
from voice_conversation_ai.transcriber import Transcriber
import whisper

app = FastAPI()


@app.get("/")
def heart_beat():
    return {"message": "ok"}


@app.get("/ws_test")
def websocket_test():
    return "ok"


@app.websocket("/ws_test")
async def websocket_test(websocket: WebSocket):
    await websocket.accept()
    transcriber = Transcriber(whisper=whisper)
    openai_llm = OpenAILLM(model_name="gpt-3.5-turbo")
    # ollama_llm = OllamaLLM()

    # response = VoiceVoxTextToSpeechResponse()
    response = SimpleTextDisplayResponse()

    conversation_logs = [
        {"role": "assistant",
         "content": "お電話ありがとうございます。首無し商事株式会社、自動応答システムでございます。"},
    ]
    while True:
        data = await websocket.receive_json()
        decoded_wav = base64.b64decode(data["media"]["payload"])
        # with open("decoded_audio.wav", "wb") as f:
        #     f.write(decoded_wav)

        transcribed_text = transcriber.transcribe(wav_bytes=decoded_wav)
        service = ConversationService(llm=openai_llm)
        clean_text = service.revise_dirty_text(input=transcribed_text, conversation_logs=conversation_logs)
        conversation_logs.append(
            {"role": "user", "content": clean_text}
        )

        response = service.chat_with_ai_operator(conversation_logs=conversation_logs)
        conversation_logs.append(
            {"role": "assistant", "content": response}
        )
        await websocket.send_text(response)
