import os
import dotenv
from fastapi import FastAPI
from pydantic import BaseModel
from graph import GraphGen
from agents.code_gen import CodeGen
from fastapi.responses import JSONResponse

class LogData(BaseModel):
    '''Class for the log data'''
    logs: str
    isUserPro: bool

class CodeSnippetGen(BaseModel):
    '''Class for the code snippet generation'''
    logs: str
    currentCodeSnippet: str
    predictedSolutions: str
    languageId: str
    isUserPro: bool

app = FastAPI()
dotenv.load_dotenv()

@app.post("/run_analysis")
async def run_chat_agent(data: LogData):
    '''Function to run the chat agent'''
    retries = 0
    if data.isUserPro:
        chat_agent = GraphGen(os.getenv('TIER2_MODEL_ENDPOINT'), tier=2)
    else:
        chat_agent = GraphGen(os.getenv('TIER1_MODEL_ENDPOINT'))
    while True:
        try:
            print(f"Number of retries {retries}")
            retries = retries + 1
            print(f"Processing agent tier {chat_agent.tier}")
            result = chat_agent.run(data.logs)
            return result
        except Exception as e:
            if retries > 8:
                print(f"Unable to process the logs, error: {e}")
                return JSONResponse(status_code=500, content={"message": "Unable to process the logs"})
            print(f"Unable to process the logs, error: {e} ... retrying")
        
@app.post("/generate_code_snippet")
async def generate_code_snippet(data: CodeSnippetGen):
    if not data.isUserPro:
        return ""
    dotenv.load_dotenv()
    chat_agent = CodeGen(os.getenv('CODE_TIER1_MODEL_ENDPOINT'))
    result = chat_agent.gen_code(data.logs, data.currentCodeSnippet, data.predictedSolutions, data.languageId)
    return result
        