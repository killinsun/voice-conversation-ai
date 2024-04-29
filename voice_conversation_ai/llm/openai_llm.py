from typing import Optional

from openai import OpenAI

from voice_conversation_ai.llm.llm_base import ILLM, CompletionOptions
from dotenv import load_dotenv

load_dotenv()


class OpenAILLM(ILLM):

    def __init__(self, model_name: str = "gpt-3.5-turbo"):
        self.model_name = model_name
        self.client = OpenAI()

    def create_completion(self, messages: list[dict[str, str]], options: Optional[CompletionOptions]) -> str:
        response = self.client.chat.completions.create(
            model=self.model_name,
            messages=messages,
            temperature=0.1,
        )

        return response.choices[0].message.content
