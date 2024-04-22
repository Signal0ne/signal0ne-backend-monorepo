"""Module for title and summary generator"""
import os
import json
from langchain_community.llms import HuggingFaceEndpoint
from langchain_openai.llms import OpenAI
from dotenv import load_dotenv
from utils.utils import parse_json


class TitleAgent:
    """Class for the chat agent."""
    def __init__(self, endpoint,tier):
        load_dotenv()
        if tier == 2:
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
        else:
            self.llm = HuggingFaceEndpoint(
                endpoint_url=endpoint,
                task="text-generation",
                max_new_tokens=512,
                top_k=50,
                temperature=0.4,
                repetition_penalty=1.1,
            )
            self.prompt = """Act like a software debugger and generate a title for these
                         error logs and give a single paragraph summary of the error logs in technical detail in a json format.
                         logs: {logs})
                         Output json format is {{\"title\": \"your title\", \"logsummary\": \"your summary\"}}
                         json: """
        
    def gen_title(self, logs):
        """Generate questions from the logs."""
        result = self.llm(self.prompt.format(logs=logs))
        result = parse_json(result)
        print(result)
        return json.loads(result)
