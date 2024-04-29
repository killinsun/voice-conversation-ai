from abc import ABC, abstractmethod
from dataclasses import dataclass
from typing import Optional


@dataclass
class CompletionOptions:
    max_tokens: Optional[int] = 256
    temperature: Optional[float] = 1
    top_p: Optional[float] = 1
    frequency_penalty: Optional[float] = 0
    presence_penalty: Optional[float] = 0
    stop: Optional[list[str]] = None

    def dict(self):
        return {k: v for k, v in self.__dict__.items() if v is not None}


class ILLM(ABC):
    @abstractmethod
    def create_completion(self, messages: list[dict[str, str]], options: Optional[CompletionOptions]) -> str:
        raise NotImplementedError
