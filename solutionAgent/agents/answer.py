"""Module for AnswerGenerator class"""
import os
from langchain_community.llms import HuggingFaceEndpoint
from dotenv import load_dotenv
from langchain_openai import ChatOpenAI
from langchain_openai.llms import OpenAI

class AnswerGenerator:
    """Class for the chat agent."""
    def __init__(self, endpoint,tier):
        load_dotenv()
        self.endpoint = endpoint
        self.tier = tier
        if tier == 2:
            self.llm = ChatOpenAI(
                api_key=os.getenv("OPENAI_API_KEY"),
                model=endpoint,
                temperature=0.3,
                max_tokens=512,
                frequency_penalty=1.1
            )
            self.prompt = """System: You are a helpful software engineer whose job is to help solve
        the error in the logs using logs and relevant context given. Use the context provided to resolve issue. Remeber context is just for similar cases not this particular one.
        to solve the error in the logs. Take a note of the what the context says.
        Give max 3 possible solutions to the error with sample code or commands. You will be punished for skipping variable or function names. Do not give any alternate answers or any other information except solution.
        logs: {logs}\n
        context: {context}\n"""
        else:
            self.llm = OpenAI(
                api_key=os.getenv("OPENAI_API_KEY"),
                name=endpoint,
                temperature=0.3,
                max_tokens=512,
                frequency_penalty=1.1
            )
            self.prompt = """System: You are a helpful software engineer whose job is to help solve
        the error in the logs using logs and relevant context given. Use the context provided to resolve issue. Remeber context is just for similar cases not this particular one.
        to solve the error in the logs. Take a note of the what the context says.
        Give max 3 possible solutions to the error with sample code or commands. You will be punished for skipping variable or function names. Do not give any alternate answers or any other information except solution.
        logs: {logs}\n
        context: {context}\n"""

    def generate_answer(self, *args, **kwargs):
        """Generate answer from the logs and context."""
        logs = kwargs.get("logs", "")
        print("Filtered logs: ", logs)
        context = args[0] if args else {}
        urls = list(set([item['url'] for item in context if 'url' in item]))
        formatted_prompt = self.prompt.format(logs=logs, context=str(context))
        solution = self.__execute(formatted_prompt)
        # solution = self.__evaluate(solution, logs)
        return solution, urls
    
    def __evaluate(self, solution: str, logs: str):
        eval_prompt = f"""System: You are a helpful software engineer whose job is to evaluate the solution given below.
        Evaluate the solution given below in context of given logs. If there is any mistake in the solution, edit the solution to be correct.
        Do not add any extra information. Do not change the solution if it is correct.
        Solutions: {solution}
        Logs: {logs}"""
        return self.__execute(eval_prompt)
    
    def __execute(self, formatted_prompt: str):
        if self.tier == 2:
            messages = [
                ("human", formatted_prompt),
            ]
            return self.llm.invoke(messages).content
        else:
            return self.llm(formatted_prompt)

