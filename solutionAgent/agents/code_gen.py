"""Module for code snippet generator"""
import os
import json
from langchain_community.llms import HuggingFaceEndpoint
from dotenv import load_dotenv
from utils.utils import parse_json

class CodeGen:
    """Class for the chat agent."""
    def __init__(self, endpoint):
        load_dotenv()
        self.llm = HuggingFaceEndpoint(
            endpoint_url=endpoint,
            task="text-generation",
            max_new_tokens=250,
            top_k=50,
            temperature=0.4,
            repetition_penalty=1.1,
        )

    def gen_code(self, logs, current_code, predicted_solutions, lnguageId):
        """Generate code snippets from the logs."""
        prompt = f"""You are a helpful assistant that helps to fix code written in {lnguageId}. 
        You return a json with the code snippet. You only return the code no explanation.
        Your return type is json.
        Here are the logs: {logs} caused by issue in code fix it\n
        Here is the broken code: {current_code}\n
        Your output format is {{"code":"your code snippet"}}. Json:"""
        result = self.llm(prompt)
        result = parse_json(result)
        print(result)
        return json.loads(result)