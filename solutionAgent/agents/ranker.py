"""Module for output ranker"""
import os
import json
import re
from typing import List
from langchain_openai import ChatOpenAI
from langchain_openai.llms import OpenAI
from dotenv import load_dotenv


class RankAgent:
    """Class for the chat agent."""
    def __init__(self, endpoint,tier):
        load_dotenv()
        self.tier = tier
        if tier == 2:
            self.llm = ChatOpenAI(
                api_key=os.getenv("OPENAI_API_KEY"),
                model=endpoint,
                temperature=0.4,
                max_tokens=100,
                frequency_penalty=1.1
            )
            self.prompt = """System: You are a helpful assistant that helps ranking the top outputs of websearch based on how relevant
        they are to solving the error recieved in logs. You may use snippet or summary do do your ranking.
        You only give ranking of the indexes of the websearch results.
        Here are user logs: {logs}\n
        Here are the websearch results: {outputs}\n
        Only return the index of the most relevant outputs. Use json output format specified below.\n
        Output format is {{"ranks":[1,2,4,.....]}}.\n
        Do not give any alternate answers or any other information except json.
        Output Json:"""
        else:
            self.llm = ChatOpenAI(
                api_key=os.getenv("OPENAI_API_KEY"),
                model=endpoint,
                temperature=0.4,
                max_tokens=100,
                frequency_penalty=1.1
            )
            self.prompt = """System: You are a helpful assistant that helps ranking the top outputs of websearch based on how relevant
        they are to solving the error recieved in logs. You may use snippet or summary do do your ranking.
        You only give ranking of the indexes of the websearch results.
        Here are user logs: {logs}\n
        Here are the websearch results: {outputs}\n
        Only return the index of the most relevant outputs. Use json output format specified below.\n
        Output format is {{"ranks":[1,2,4,.....]}}.\n
        Do not give any alternate answers or any other information except json.
        Output Json:"""
        
    def rank(self, outputs, logs):
        """Generate questions from the logs."""
        formatted_prompt = self.prompt.format(logs=logs, outputs=outputs)
        i=0
        while i<3:
            try:
                i+=1
                result = self.__execute(formatted_prompt)
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
    
    def __execute(self, formatted_prompt: str):
        if self.tier == 2:
            messages = [
                ("human", formatted_prompt),
            ]
            return self.llm.invoke(messages).content
        else:
            return self.llm(formatted_prompt)
