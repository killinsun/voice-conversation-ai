from io import BytesIO
import soundfile as sf
import numpy as np


class Transcriber:

    def __init__(self, whisper):
        self.whisper = whisper

    def transcribe(self, wav_bytes: bytes, model_name: str = "mlx-community/whisper-medium-mlx") -> str:
        wav_stream = BytesIO(wav_bytes)
        audio_array, sampling_rate = sf.read(wav_stream)
        audio_fp32 = audio_array.astype(np.float32)

        result = self.whisper.transcribe(
            audio_fp32,
            path_or_hf_repo=model_name,
            initial_prompt="こんにちは、元気ですか？ありがとうございました。"  # 句読点形式で文字起こしさせるため
        )

        return result["text"]
