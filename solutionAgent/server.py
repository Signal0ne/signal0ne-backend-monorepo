import os
import dotenv
from fastapi import FastAPI
from pydantic import BaseModel
from graph import GraphGen

class LogData(BaseModel):
    '''Class for the log data'''
    logs: str

app = FastAPI()

@app.post("/run_analysis")
async def run_chat_agent(data: LogData):
    '''Function to run the chat agent'''
    dotenv.load_dotenv()
    chat_agent = GraphGen(os.getenv('ENDPOINT_URL'))
    backup_chat_agent = GraphGen(os.getenv('BACKUP_ENDPOINT_URL'))
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
        