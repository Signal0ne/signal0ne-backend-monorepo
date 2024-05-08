"""Module for question generator"""
import os
import json
from langchain_openai.llms import OpenAI
from langchain_openai import ChatOpenAI
from dotenv import load_dotenv
from utils.utils import parse_json


class QueryAgent:
    """Class for the chat agent."""
    def __init__(self, endpoint,tier):
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
            self.prompt = """You are a helpful assistant that helps generate 3 highly descriptive relevant queries from a set of error logs for google search.
        Take a note of the module and library causing error or any component that could help in searching the web for more information. You return a json with the queries. Do not forget to add relevant log
        statements to each of your queries. Your return type is json. You only output in the format specified below.
        Here are the logs: {logs}\n
        Output format is {{"queries": [{{"question":"your question","context":"relevant log statements"}}]}}. Json:"""
        else:
            self.llm = OpenAI(
                api_key=os.getenv("OPENAI_API_KEY"),
                name=endpoint,
                temperature=0.4,
                max_tokens=512,
                frequency_penalty=1.1
            )
            self.prompt = """You are a helpful assistant that helps generate 3 highly descriptive relevant queries from a set of error logs for google search.
        Take a note of the module and library causing error or any component that could help in searching the web for more information. You return a json with the queries. Do not forget to add relevant log
        statements to each of your queries. Your return type is json. You only output in the format specified below.
        Here are the logs: {logs}\n
        Output format is {{"queries": [{{"question":"your question","context":"relevant log statements"}}]}}. Json:"""
        
    def gen_ques(self, logs):
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
