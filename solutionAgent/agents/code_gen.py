"""Module for code snippet generator"""
import os
import json
import tiktoken
from langchain_openai import ChatOpenAI
from dotenv import load_dotenv
from utils.utils import parse_json

class CodeGen:
    """Class for the chat agent."""
    def __init__(self, endpoint):
        load_dotenv()
        self.endpoint = endpoint
        self.plain_output_llm = ChatOpenAI(
            api_key=os.getenv("OPENAI_API_KEY"),
            model=endpoint,
            temperature=0.4,
            max_tokens=1024,
            frequency_penalty=1.1,
        )

        self.llm = ChatOpenAI(
            api_key=os.getenv("OPENAI_API_KEY"),
            model=endpoint,
            temperature=0.4,
            max_tokens=1024,
            frequency_penalty=1.1,
            model_kwargs={"response_format": {"type": "json_object"}}
        )

    def gen_code(self, logs, current_code, predicted_solutions, languageId):
        """Generate code snippets from the logs."""
        result = {
            "explanation": ""
        }
        if self.__count_tokens(current_code) > 200:
            result["error"] = "Too long code block! Try to provide a shorter code snippet to fix."
            return result
        
        initial_reflection = self.__base_code_matching_issue_reflection(current_code, logs, languageId)
        if bool(initial_reflection["relevant"]):
            context = self.__context_gen(current_code, languageId, logs)
            result["explanation"] = context
            result = json.dumps(result)
            return json.loads(result)
        else:
            result["error"] = "Ups. Code block doesn't seem to be relevant to this issue try to choose other code block to fix."
            return result
    
    def __base_code_matching_issue_reflection(self, current_code, logs, languageId):
        """Reflect if the code matches the issue context as Software Engineer persona"""
        
        prompt = f"""Your are helpful assistant who helps estimate if the code can be fixed based on provided logs 
        You return json with true or false. True if the code could produce the following log output and false otherwise.
        You will be punished for false positives.
        Code: {current_code}\n
        Language: {languageId}\n
        Logs: {logs}\n
        Your output format is {{"relevant":"answer"}}. Json:"""
        result = self.__execute(prompt, "json")
        result = parse_json(result)
        return json.loads(result)
    
    def __context_gen(self, current_code, languageId, logs):
        """Pinpoint issue in code as Software Engineer persona"""
        prompt = f"""You are an expert in {languageId} and you are asked to pinpoint the fragment of code which most likely caused the issue based on logs.
        Propose the change. You return it along with the explanation why do you think it is the cause.
        Code: {current_code}\n
        Logs: {logs}\n"""

        context = self.__execute(prompt, "plain")

        validation_prompt = f"""You are an expert in {languageId} and you are asked to review debugging process of less experienced engineer and propose other change if previous one was wrong.
        Return just one fragment of code and one explanation. You will be punished for more and wrong answers.
        Code: {current_code}\n
        Context: {context}
        Logs: {logs}\n"""

        context = self.__execute(validation_prompt, "plain")

        return context
    
    def __count_tokens(self, code: str):
        model_encoding = tiktoken.encoding_for_model(self.endpoint)
        return len(model_encoding.encode(code))
    
    def __execute(self, formatted_prompt: str, output_format: str = "json"):
        messages = [
                ("human", formatted_prompt),
        ]
        if output_format == "json":
            return self.llm.invoke(messages).content
        else:
            return self.plain_output_llm.invoke(messages).content
