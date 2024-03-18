import os
import time
from fastapi import FastAPI
from pydantic import BaseModel
from agent import ChatAgent

class LogData(BaseModel):
    '''Class for the log data'''
    logs: str

app = FastAPI()

@app.post("/run_analysis")
async def run_chat_agent(data: LogData):
    '''Function to run the chat agent'''
    chat_agent = ChatAgent(os.getenv('ENDPOINT_URL'))
    backup_chat_agent = ChatAgent(os.getenv('BACKUP_ENDPOINT_URL'))
    retries = 0
    while True:
        try:
            print(f"Number of retries {retries}")
            retries =retries + 1
            result = chat_agent.run(data.logs)
            return result
        except Exception as e:
            result = backup_chat_agent.run(data.logs)
            if retries > 4:
                print(f"Unable to process the logs, error: {e}")
                return {"error": f"Unable to process the logs, error: {e}"}
            return result
        