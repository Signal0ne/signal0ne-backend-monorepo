"""Module for code snippet generator"""
import os
import json
from langchain_community.llms import HuggingFaceEndpoint
from dotenv import load_dotenv

class CodeGen:
    """Class for the chat agent."""
    def __init__(self, endpoint):
        load_dotenv()
        self.llm = HuggingFaceEndpoint(
            endpoint_url=endpoint,
            task="text-generation",
            max_new_tokens=512,
            top_k=50,
            temperature=0.4,
            repetition_penalty=1.1,
        )

    def gen_code(self, logs, current_code, predicted_solutions, lnguageId):
        print(lnguageId)
        """Generate code snippets from the logs."""
        prompt = f"""You are a helpful assistant that helps to fix code written in {lnguageId} based on the error and logs for a given code snippet. 
        You return a json with the code snippet. You only return the code no explanation.
        Your return type is json.
        Here are the logs: {logs}\n
        Here is the current code: {current_code}\n
        Output format is {{"code":"your code snippet"}}. Json:"""
        result = self.llm(prompt)
        print(result)
        return json.loads(result)