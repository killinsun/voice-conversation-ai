from typing import Optional
import ollama

from openai import OpenAI

from voice_conversation_ai.llm.llm_base import ILLM, CompletionOptions


class OllamaLLM(ILLM):
    def __init__(self, model_name: str = "llama3"):
        self.model_name = model_name
        self.client = OpenAI()

    def create_completion(self, messages: list[dict[str, str]], options: Optional[CompletionOptions]) -> str:
        ollama_options = {
            "top_k": 1,
            "temperature": options.temperature,
            "top_p": options.top_p,
            "num_predict": 1,
            "penalize_newline": True,
        }

        result = ollama.chat(
            model=self.model_name,
            messages=messages,
            options=ollama_options
        )
        return result["message"]["content"]
