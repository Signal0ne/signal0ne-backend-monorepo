"""Module for question generator"""
import os
import json
from langchain_community.llms import HuggingFaceEndpoint
from dotenv import load_dotenv


class ChatAgent:
    """Class for the chat agent."""
    def __init__(self, endpoint):
        load_dotenv()
        self.llm = HuggingFaceEndpoint(
            endpoint_url=endpoint,
            task="text-generation",
            max_new_tokens=512,
            top_k=50,
            temperature=0.4,
            repetition_penalty=1.1,
        )
        
    def gen_ques(self, logs):
        """Generate questions from the logs."""
        prompt = f"""You are a helpful assistant that helps generate 3 highly descriptive relevant queries from a set of error logs for google search.
        You can also include parts from the logs in your questions. You return a json with the queries. Do not forget to add relevant log
        statements and to each of your queries. Your return type is json.
        Here are the logs: {logs}\n
        Output format is {{"question":"your question","context":"the context for the question"}}. Json:"""
        result = self.llm(prompt)
        print(result)
        return json.loads(result)
    
if __name__ == "__main__":
    load_dotenv()
    endpoint = os.getenv("ENDPOINT_URL")
    agent = ChatAgent(endpoint)
    logs = """Error occurred type=\"error\" text=\"Missing job runner for an existing job - #######\" stackTrace= at Kudu.Core.Jobs.ContinuousJobsManager.EnableJob(String jobName)\n   at Kudu.Services.Jobs.JobsController.EnableContinuousJob(String jobName)\r\n   at lambda_method(Closure , Object , Object[] )\n   at System.Web.Http.Controllers.ReflectedHttpActionDescriptor.ActionExecutor.<>c__DisplayClass10.<GetExecutor>b__9(Object instance, Object[] methodParameters)"""
    print(agent.gen_ques(logs))
