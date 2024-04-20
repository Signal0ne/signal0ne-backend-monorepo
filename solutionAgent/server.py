import os
import time
from fastapi import FastAPI
from pydantic import BaseModel
from agent import ChatAgent

class LogData(BaseModel):
    '''Class for the log data'''
    logs: str
    isUserPro: bool

app = FastAPI()
master_chat_agent = ChatAgent(os.getenv('ENDPOINT_URL'),tier=1)
master_tier2_chat_agent = ChatAgent(os.getenv('TIER2_MODEL_ENDPOINT'), tier=2)

backup_chat_agent = ChatAgent(os.getenv('BACKUP_ENDPOINT_URL'), tier=1)

@app.post("/run_analysis")
async def run_chat_agent(data: LogData):
    '''Function to run the chat agent'''
    chat_agent = master_chat_agent
    retries = 0
    while True:
        try:
            print(f"Number of retries {retries}")
            retries = retries + 1
            print(f"Processing agent tier {chat_agent.tier}")
            result = chat_agent.run(data.logs)
            return result
        except Exception as e:
            print(f"Unable to process the logs, error: {e} ... retrying")
            result = backup_chat_agent.run(data.logs)
            if retries > 4:
                print(f"Unable to process the logs, error: {e}")
                return {"error": f"Unable to process the logs, error: {e}"}
            return None
        