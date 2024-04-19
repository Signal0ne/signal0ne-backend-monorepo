import os
import dotenv
from fastapi import FastAPI
from pydantic import BaseModel
from graph import GraphGen
from agents.code_gen import CodeGen

class LogData(BaseModel):
    '''Class for the log data'''
    logs: str

class CodeSnippetGen(BaseModel):
    '''Class for the code snippet generation'''
    logs: str
    currentCodeSnippet: str
    predictedSolutions: str
    languageId: str

app = FastAPI()
dotenv.load_dotenv()
chat_agent = GraphGen(os.getenv('ENDPOINT_URL'))
backup_chat_agent = GraphGen(os.getenv('BACKUP_ENDPOINT_URL'))

@app.post("/run_analysis")
async def run_chat_agent(data: LogData):
    '''Function to run the chat agent'''
    retries = 0
    while True:
        try:
            print(f"Number of retries {retries}")
            retries =retries + 1
            result = chat_agent.run(data.logs)
            return result
        except Exception as e:
            print(f"Unable to process the logs, error: {e}")
            result = backup_chat_agent.run(data.logs)
            return result
        
@app.post("/generate_code_snippet")
async def generate_code_snippet(data: CodeSnippetGen):
    dotenv.load_dotenv()
    chat_agent = CodeGen(os.getenv('ENDPOINT_URL'))
    result = chat_agent.gen_code(data.logs, data.currentCodeSnippet, data.predictedSolutions, data.languageId)
    return result
        