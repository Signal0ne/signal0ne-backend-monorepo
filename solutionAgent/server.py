import os
import dotenv
from fastapi import FastAPI
from pydantic import BaseModel
from graph import GraphGen
from agents.code_gen import CodeGen
from transformers import pipeline, AutoTokenizer, AutoModelForSeq2SeqLM

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

tokenizer = AutoTokenizer.from_pretrained('VidhuMathur/bart-log-summarization')
model = AutoModelForSeq2SeqLM.from_pretrained('facebook/bart-base')  
model_pipeline = pipeline("summarization", model=model, tokenizer=tokenizer, framework="pt")

app = FastAPI()
dotenv.load_dotenv()
@app.post("/run_analysis")
async def run_chat_agent(data: LogData):
    '''Function to run the chat agent'''
    retries = 0
    if data.isUserPro:
        chat_agent = GraphGen(model_pipeline,os.getenv('TIER2_MODEL_ENDPOINT'), tier=2)
    else:
        print("Running tier 1 model")
        chat_agent = GraphGen(model_pipeline,os.getenv('TIER1_MODEL_ENDPOINT'))
    while True:
        try:
            retries = retries + 1
            result = chat_agent.run(data.logs)
            return result
        except Exception as e:
            if retries > 8:
                print(f"Unable to process the logs, error: {e}")
                return
            print(f"Unable to process the logs, error: {e} ... retrying")
        
@app.post("/generate_code_snippet")
async def generate_code_snippet(data: CodeSnippetGen):
    if not data.isUserPro:
        return ""
    dotenv.load_dotenv()
    chat_agent = CodeGen(os.getenv('CODE_TIER1_MODEL_ENDPOINT'))
    result = chat_agent.gen_code(data.logs, data.currentCodeSnippet, data.predictedSolutions, data.languageId)
    return result
        