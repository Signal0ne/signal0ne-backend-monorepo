import os
import json
from langchain_openai.llms import OpenAI
from langchain_openai import ChatOpenAI
from dotenv import load_dotenv
from utils.utils import parse_json

class CulpritDiscovery:
    def __init__(self, endpoint,tier=1):
        load_dotenv()
        self.tier = tier
        if tier == 2:
            self.llm = ChatOpenAI(
                api_key=os.getenv("OPENAI_API_KEY"),
                model=endpoint,
                temperature=0.4,
                max_tokens=40,
                frequency_penalty=1.1
            )
            self.prompt = self.prompt = """You are a helpful assistant that helps find a file or function with a culrpit from software error stacktarce.
        You return a json with most probable file name or function name you can find in error message containing culprit.
        Here are the logs: {logs}\n
        Output format is {{"filename": "file or function name containing code culprit"}}. Json:"""
        else:
            self.llm = OpenAI(
                api_key=os.getenv("OPENAI_API_KEY"),
                name=endpoint,
                temperature=0.4,
                max_tokens=40,
                frequency_penalty=1.1
            )
            self.prompt = """You are a helpful assistant that helps find a file or function with a culrpit from software error stacktarce.
        You return a json with most probable file name or function name you can find in error message containing culprit.
        Here are the logs: {logs}\n
        Output format is {{"filename": "file or function name containing code culprit"}}. Json:"""
            
    def discover_culprit(self, logs):
        """Generate questions from the logs."""
        formatted_prompt = self.prompt.format(logs=logs)
        result = self.__execute(formatted_prompt)
        result = parse_json(result)
        return json.loads(result)

    def __execute(self, formatted_prompt: str):
        if self.tier == 2:
            messages = [
                ("human", formatted_prompt),
            ]
            return self.llm.invoke(messages).content
        else:
            return self.llm(formatted_prompt)