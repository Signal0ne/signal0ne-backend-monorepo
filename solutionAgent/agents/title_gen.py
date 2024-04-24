"""Module for title and summary generator"""
import os
import json
from langchain_openai.llms import OpenAI
from langchain_openai import ChatOpenAI
from dotenv import load_dotenv
from utils.utils import parse_json


class TitleAgent:
    """Class for the chat agent."""
    def __init__(self, endpoint,tier):
        load_dotenv()
        self.tier = tier
        if tier == 2:
            self.llm = ChatOpenAI(
                api_key=os.getenv("OPENAI_API_KEY"),
                name=endpoint,
                temperature=0.4,
                max_tokens=512,
                frequency_penalty=1.1
            )
            self.prompt = """Act like a software debugger and generate a title for these
                         error logs and give a single paragraph summary of the error logs in technical detail in a json format.
                         logs: {logs})
                         Output json format is {{\"title\": \"your title\", \"logsummary\": \"your summary\"}}
                         json: """
        else:
            self.llm = OpenAI(
                api_key=os.getenv("OPENAI_API_KEY"),
                name=endpoint,
                temperature=0.4,
                max_tokens=512,
                frequency_penalty=1.1
            )
            self.prompt = """Act like a software debugger and generate a title for these
                         error logs and give a single paragraph summary of the error logs in technical detail in a json format.
                         logs: {logs})
                         Output json format is {{\"title\": \"your title\", \"logsummary\": \"your summary\"}}
                         json: """
        
    def gen_title(self, logs):
        """Generate questions from the logs."""
        formatted_prompt = self.prompt.format(logs=logs)
        result = self.__execute(formatted_prompt)
        result = parse_json(result)
        print(result)
        return json.loads(result)

    def __execute(self, formatted_prompt: str):
        if self.tier == 2:
            messages = [
                ("human", formatted_prompt),
            ]
            return self.llm.invoke(messages).content
        else:
            return self.llm(formatted_prompt)
