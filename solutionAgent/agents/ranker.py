"""Module for output ranker"""
import os
import json
import re
from typing import List
from langchain_community.llms import HuggingFaceEndpoint
from dotenv import load_dotenv


class RankAgent:
    """Class for the chat agent."""
    def __init__(self, endpoint):
        load_dotenv()
        self.llm = HuggingFaceEndpoint(
            endpoint_url=endpoint,
            task="text-generation",
            max_new_tokens=100,
            top_k=30,
            temperature=0.3,
            repetition_penalty=1.1,
        )
        
    def rank(self, outputs, logs):
        """Generate questions from the logs."""
        prompt = f"""System: You are a helpful assistant that helps ranking the top outputs of websearch based on how relevant
        they are to solving the error recieved in logs. You may use snippet or summary do do your ranking.
        You only give ranking of the indexes of the websearch results.
        Here are user logs: {logs}\n
        Here are the websearch results: {outputs}\n
        Only return the index of the most relevant outputs. Use json output format specified below.\n
        Output format is {{"ranks":[1,2,4,.....]}}.\n
        Do not give any alternate answers or any other information except json.
        Output Json:"""
        i=0
        while i<3:
            try:
                i+=1
                result = self.llm(prompt)
                match = re.search(r'{(.*?)}', result)
                if match:
                    extracted_string = match.group(1)
                    result = json.loads("{"+extracted_string+"}")
                    outputs = outputs.replace('\n', '')
                    context = json.loads(outputs)
                    ranks = result['ranks']
                    selected_context = []
                    for rank in ranks:
                        for item in context:
                            if item["index"] == rank:
                                selected_context.append({"url": item["url"],"snippet":item["snippet"], "summary": item["summary"]})
                                break
                else:
                    selected_context = json.loads("")
                return selected_context
            except Exception as e:
                print(f"Error in decoding json: {e}")
                continue
        return []
