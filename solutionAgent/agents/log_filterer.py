"""Module for title and summary generator"""
import os
import json
from langchain_openai.llms import OpenAI
from langchain_openai import ChatOpenAI
from dotenv import load_dotenv


class LogFilterer:
    """Class for the log filterer."""
    def __init__(self, endpoint,tier=1):
        load_dotenv()
        self.tier = tier
        if tier == 2:
            self.llm = ChatOpenAI(
                api_key=os.getenv("OPENAI_API_KEY"),
                model=endpoint,
                temperature=0.3,
                max_tokens=512,
                frequency_penalty=1.3
            )
            self.prompt = """System: You are a helpful software engineer whose job is to filter the relevant logs from logtail to resolve the issue in the logs.
                            Filter the logs given below to only the most relevant logs that directly indicate issue. Do not include any irrelevant logs.
                            Logs: {logs}"""
        else:
            self.llm = ChatOpenAI(
                api_key=os.getenv("OPENAI_API_KEY"),
                model=endpoint,
                temperature=0.3,
                max_tokens=512,
                frequency_penalty=1.3
            )
            self.prompt = """System: You are a helpful software engineer whose job is to filter the relevant logs from logtail to resolve the issue in the logs.
                            Filter the logs given below to only the most relevant logs that directly indicate issue. Do not include any irrelevant logs.
                            Logs: {logs}"""
        
    def filter_relevant_logs(self, logs, severity="INFO"):
        filter_prompt = self.prompt.format(logs=logs, severity=severity)
        return self.__execute(filter_prompt)

    def __execute(self, formatted_prompt: str):
        if self.tier == 2:
            messages = [
                ("human", formatted_prompt),
            ]
            return self.llm.invoke(messages).content
        else:
            messages = [
                ("human", formatted_prompt),
            ]
            return self.llm.invoke(messages).content
