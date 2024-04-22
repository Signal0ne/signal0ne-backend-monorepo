"""Module for AnswerGenerator class"""
import os
from langchain_community.llms import HuggingFaceEndpoint
from dotenv import load_dotenv
from langchain_openai.llms import OpenAI

class AnswerGenerator:
    """Class for the chat agent."""
    def __init__(self, endpoint,tier):
        load_dotenv()
        self.endpoint = endpoint
        if tier == 2:
            self.llm = OpenAI(
                api_key=os.getenv("OPENAI_API_KEY"),
                name=endpoint,
                temperature=0.3,
                max_tokens=512,
                frequency_penalty=1.1
            )
        else:
            self.llm = HuggingFaceEndpoint(
                endpoint_url=self.endpoint,
                task="text-generation",
                max_new_tokens=512,
                top_k=50,
                temperature=0.3,
                repetition_penalty=1.1,
            )

    def generate_answer(self, *args, **kwargs):
        """Generate answer from the logs and context."""
        logs = kwargs.get("logs", "")
        context = args[0] if args else {}
        urls = list(set([item['url'] for item in context if 'url' in item]))
        answer_prompt = f"""System: You are a helpful software engineer whose job is to help solve
        the error in the logs using logs and relevant context given. Use the context provided to resolve issue. Remeber context is just for similar cases not this particular one.
        to solve the error in the logs. Take a note of the what the context says.
        Give max 3 possible solutions to the error with sample code or commands. You will be punished for skipping variable or function names. Do not give any alternate answers or any other information except solution.
        logs: {logs}\n
        context: {str(context)}\n"""
        solution = self.llm(answer_prompt)
        return solution, urls

