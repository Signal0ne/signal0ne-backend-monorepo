"""Module for code snippet generator"""
import os
import json
from langchain_openai import ChatOpenAI
from dotenv import load_dotenv
from websearch.search import GoogleCustomSearch
from utils.utils import parse_json

class CodeGen:
    """Class for the chat agent."""
    def __init__(self, endpoint):
        load_dotenv()
        self.llm = ChatOpenAI(
            api_key=os.getenv("OPENAI_API_KEY"),
            model=endpoint,
            temperature=0.4,
            max_tokens=1024,
            frequency_penalty=1.1,
            # model_kwargs={"response_format": {"type": "json_object"}}
        )
        self.search = GoogleCustomSearch()

    def gen_code(self, logs, current_code, predicted_solutions, languageId):
        """Generate code snippets from the logs."""

        initial_reflection = self.__base_code_macthing_issue_reflection(current_code, logs, languageId)
        if bool(initial_reflection["relevant"]):

            context = self.__context_gen(current_code, languageId, logs)

            prompt = f"""You are a helpful assistant that helps to fix code written in {languageId}. 
            You return a json with the code snippet. You only return the code no explanation.
            Your return type is json newline must be \\n.
            Here is the broken code: {current_code}\n
            Here is additional context of proposed solutions by other engineer: {context} You need to adjust this suggestion to make it work with code you are fixing.\n
            Your output format is {{"code":"your code snippet"}}. Json:"""
            result = self.__execute(prompt)
            result = parse_json(result)
            result_object = json.loads(result)
            result = self.__base_reflection(result_object["code"], languageId, context)
            return result
        else:
            return {"code": ""}
    
    def __base_reflection(self, initial_result, languageId, context):
        """Reflect on the result as Code Reviewer persona"""
        
        prompt = f"""You are a helpful assistant that helps to review code written in {languageId} and apply changes to it if needed.
        If the code is correct you return the code as is. 
        You return a json with the code snippet. You only return the code no explanation.
        Your return type is json newline must be \\n.
        Here is the code to be reviewed: {initial_result}\n
        Here is the context of the issue: {context}\n
        Your output format is {{"code":"your code snippet with your changes"}}. Json:"""
        result = self.__execute(prompt)
        result = parse_json(result)
        return json.loads(result)
    
    def __base_code_macthing_issue_reflection(self, current_code, logs, languageId):
        """Reflect if the code matches the issue context as Software Engineer persona"""
        
        prompt = f"""Your are helpful assistant who helps estimate if the code can be fixed based on provided logs 
        You return json with true or false. True if the code could produce the following log output and false otherwise.
        You will be punished for false positives.
        Code: {current_code}\n
        Language: {languageId}\n
        Logs: {logs}\n
        Your output format is {{"relevant":"answer"}}. Json:"""
        result = self.__execute(prompt)
        result = parse_json(result)
        return json.loads(result)
    
    def __context_gen(self, current_code, languageId, logs):
        """Pinpoint issue in code as Software Engineer persona"""
        prompt = f"""You are an expert in {languageId} and you are asked to pinpoint the fragment of code which most likely caused the issue based on logs.
        Propose the change. You return it along with the explanation why do you think it is the cause.
        Code: {current_code}\n
        Logs: {logs}\n"""

        context = self.__execute(prompt)

        validation_prompt = f"""You are an expert in {languageId} and you are asked to review debugging process of less experienced engineer and propose other change if previous one was wrong.
        Return just one fragment of code and one explanation. You will be punished for more and wrong answers.
        Code: {current_code}\n
        Context: {context}
        Logs: {logs}\n"""

        context = self.__execute(validation_prompt)

        return context
    
    def __execute(self, formatted_prompt: str):
        messages = [
                ("human", formatted_prompt),
        ]
        return self.llm.invoke(messages).content
