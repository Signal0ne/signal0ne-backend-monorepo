"""Module for code snippet generator"""
import os
import json
from langchain_community.llms import HuggingFaceEndpoint
from langchain_openai import OpenAI
from dotenv import load_dotenv
from utils.utils import parse_json

class CodeGen:
    """Class for the chat agent."""
    def __init__(self, endpoint):
        load_dotenv()
        openai_api_key = os.getenv("OPENAI_API_KEY")
        self.llm = OpenAI(
                openai_api_key=openai_api_key,
                model_name=endpoint,
                max_tokens=1512,
                temperature=0.4,
                frequency_penalty=1.1,
        )

    def gen_code(self, logs, current_code, predicted_solutions, lnguageId):
        """Generate code snippets from the logs."""
        prompt = f"""You are a helpful assistant that helps to fix code written in {lnguageId}. 
        You return a json with the code snippet. You only return the code no explanation.
        Your return type is json newline must be \\n.
        Here are the logs: {logs} caused by issue in code fix it\n
        Here is the broken code: {current_code}\n
        Here is additional context on the issue and proposed solutions bu other engineer: {predicted_solutions}\n
        Your output format is {{"code":"your code snippet"}}. Json:"""
        result = self.llm(prompt)
        result = parse_json(result)
        json.dumps(result)
        print(result)
        return json.loads(result)