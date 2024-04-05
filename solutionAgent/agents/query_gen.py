"""Module for question generator"""
import os
import json
from langchain_community.llms import HuggingFaceEndpoint
from dotenv import load_dotenv


class QueryAgent:
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
        
    def gen_ques(self, logs):
        """Generate questions from the logs."""
        prompt = f"""You are a helpful assistant that helps generate 3 highly descriptive relevant queries from a set of error logs for google search.
        Take a note of the module and library causing error or any component that could help in searching the web for more information. You return a json with the queries. Do not forget to add relevant log
        statements and to each of your queries. Your return type is json. You only output in the format specified below.
        Here are the logs: {logs}\n
        Output format is {{"queries": [{{"question":"your question","context":"the context for the question"}}]}}. Json:"""
        result = self.llm(prompt)
        return json.loads(result)
