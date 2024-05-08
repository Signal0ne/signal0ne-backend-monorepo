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
                temperature=0.4,
                max_tokens=512,
                frequency_penalty=1.1
            )
            self.prompt = """System: You are a helpful software engineer whose job is to filter the relevant and most specific logs from logtail.
                            Filter the logs given below to only the most relevant and specific logs. Do not include any irrelevant logs.
                            Logs: {logs}"""
        else:
            self.llm = OpenAI(
                    api_key=os.getenv("OPENAI_API_KEY"),
                    name=endpoint,
                    temperature=0.4,
                    max_tokens=512,
                    frequency_penalty=1.1
                )
            self.prompt = """System: You are a helpful software engineer whose job is to filter the relevant and most specific logs from logtail.
                            Filter the logs given below to only the most relevant and specific logs. Do not include any irrelevant logs.
                            Logs: {logs}"""
        
    def filter_relevant_logs(self, logs):
        filter_prompt = self.prompt.format(logs=logs)
        return self.__execute(filter_prompt)

    def __execute(self, formatted_prompt: str):
        if self.tier == 2:
            messages = [
                ("human", formatted_prompt),
            ]
            return self.llm.invoke(messages).content
        else:
            return self.llm(formatted_prompt)
