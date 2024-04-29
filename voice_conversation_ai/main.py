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

    # vvox_response = VoiceVoxTextToSpeechResponse()
    simple_response = SimpleTextDisplayResponse()

    while True:
        simple_response.say("そっかそっか")
        data = await websocket.receive_json()
        decoded_wav = base64.b64decode(data["media"]["payload"])
        # with open("decoded_audio.wav", "wb") as f:
        #     f.write(decoded_wav)

        transcribed_text = transcriber.transcribe(wav_bytes=decoded_wav)
        service = ConversationService(llm=openai_llm)
        service.chat_with_ai_operator(input=transcribed_text, resp=simple_response)

        await websocket.send_text(f"OK! data:{data}")
