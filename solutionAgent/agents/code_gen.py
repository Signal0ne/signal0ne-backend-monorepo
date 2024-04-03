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

    def gen_code(self, logs, current_code, predicted_solutions):
        """Generate code snippets from the logs."""
        prompt = f"""You are a helpful assistant that helps generate code snippets from a set of error logs for a given code snippet. 
        Take a note of the module and library causing error or any component that could help in generating code snippets. You return a json with the code snippet. 
        Your return type is json. You only output in the format specified below.
        Here are the logs: {logs}\n
        Here is the current code: {current_code}\n
        Here are the predicted solutions that can help with debugging: {predicted_solutions}\n
        Output format is {{"code":"your code snippet"}}. Json:"""
        result = self.llm(prompt)
        print(result)
        return json.loads(result)