"""Module for output ranker"""
import os
import json
import re
from typing import List
from langchain_community.llms import HuggingFaceEndpoint
from dotenv import load_dotenv


class RankAgent:
    """Class for the chat agent."""
    def __init__(self, endpoint):
        load_dotenv()
        self.llm = HuggingFaceEndpoint(
            endpoint_url=endpoint,
            task="text-generation",
            max_new_tokens=100,
            top_k=30,
            temperature=0.3,
            repetition_penalty=1.1,
        )
        
    def rank(self, outputs, logs):
        """Generate questions from the logs."""
        prompt = f"""System: You are a helpful assistant that helps ranking the top 3 outputs of websearch based on how relevant
        they are to solving the error recieved in logs. You may use snippet or summary do do your ranking.
        You only give ranking of the indexes of the websearch results.
        Here are user logs: {logs}\n
        Here are the websearch results: {outputs}\n
        Only return the index of the most relevant outputs. Use json output format specified below.\n
        Output format is {{"ranks":[1,2,4,.....]}}.\n
        Do not give any alternate answers or any other information except json.
        Output Json:"""
        i=0
        while i<3:
            try:
                i+=1
                result = self.llm(prompt)
                match = re.search(r'{(.*?)}', result)
                if match:
                    extracted_string = match.group(1)
                    result = json.loads("{"+extracted_string+"}")
                    outputs = outputs.replace('\n', '')
                    context = json.loads(outputs)
                    ranks = result['ranks']
                    selected_context = []
                    for rank in ranks:
                        for item in context:
                            if item["index"] == rank:
                                selected_context.append({"url": item["url"],"snippet":item["snippet"], "summary": item["summary"]})
                                break
                else:
                    selected_context = json.loads("")
                return selected_context
            except Exception as e:
                print(f"Error in decoding json: {e}")
                continue
        return []

# if __name__ == "__main__":
#     load_dotenv()
#     endpoint = os.getenv("ENDPOINT_URL")
#     chat_agent = RankAgent(endpoint)
#     outputs ="""[
#     {
#         "index": 1,
#         "url": "https://stackoverflow.com/questions/57708114/azure-webjob-there-is-not-enough-space-on-the-disk",
#         "snippet": "Aug 29, 2019 ... ... job files: System.IO.IOException: There is not enough space on the disk. at System.IO.__Error.WinIOError(Int32 errorCode, String\u00a0...",
#         "summary": "I am getting an error on my Azure WebJob about not being enough space on the disk.\nI had it earlier, so I moved it to Standard 1 which says it comes with 50GB.\nHere is the error I get:[08/29/2019 05:19:04 > cd1c7c: SYS ERR ] Failed to copy job files: System.IO.IOException: There is not enough space on the disk.\nNOTE: I have reviewed this post: Azure WebJobs: There is not enough space on the disk - Seems pretty old, also OP was able to resolve with a higher plan.\n3) It is written in C#.net Core 2.2 as a Console Application4) I am NOT using the WebJobs SDKUPDATE 2: Also posted on GitHub"
#     },
#     {
#         "index": 2,
#         "url": "https://github.com/projectkudu/kudu/issues/2559",
#         "snippet": "Sep 14, 2017 ... Checking the Kudu solution here on github for where it is breaking down did not reveal anything obvious to me (BaseJobRunner.cs:201). I am very\u00a0...",
#         "summary": "Have a question about this project?\nSign up for a free GitHub account to open an issue and contact its maintainers and the community.\nPick a username Email Address Password Sign up for GitHubBy clicking \u201cSign up for GitHub\u201d, you agree to our terms of service and privacy statement.\nWe\u2019ll occasionally send you account related emails.\nSign in to your account"
#     },
#     {
#         "index": 3,
#         "url": "https://stackoverflow.com/questions/30354377/azure-webjobs-there-is-not-enough-space-on-the-disk",
#         "snippet": "May 20, 2015 ... ... job files: System.IO.IOException: There is not enough space on the disk. at System.IO.__Error.WinIOError(Int32 errorCode, String\u00a0...",
#         "summary": "I've tried zipping up a console app and uploading it to Azure as a WebJob, as I've done with others, but the upload fails.\nThe upload completes and I can see the WebJob listed in the Azure portal.\nThe log file contains the following:[05/20/2015 15:26:22 > e5f596: SYS INFO] Status changed to Initializing [05/20/2015 15:26:23 > e5f596: SYS INFO] Status changed to Failed [05/20/2015 15:26:23 > e5f596: SYS ERR ] Failed to copy job files: System.IO.IOException: There is not enough space on the disk.\nI've also read this page which suggests that long file names may be the cause of the fault, but I've no file names longer 69 characters.\nUPDATE: I've noticed that when I try an additional WebJob which previously worked, it incurs the same error."
#     },
#     {
#         "index": 4,
#         "url": "https://github.com/projectkudu/kudu/issues/3042",
#         "snippet": "Sep 30, 2019 ... ... kudu/trace ##[error]Error ... I have tried playing around with adding the \"Complete Swap\" task to my deploy job, but it doesn't seem to make a\u00a0...",
#         "summary": "Have a question about this project?\nSign up for a free GitHub account to open an issue and contact its maintainers and the community.\nPick a username Email Address Password Sign up for GitHubBy clicking \u201cSign up for GitHub\u201d, you agree to our terms of service and privacy statement.\nWe\u2019ll occasionally send you account related emails.\nSign in to your account"
#     },
#     {
#         "index": 5,
#         "url": "https://discuss.circleci.com/t/known-hosts-file-not-existing-cant-do-ssh-in-deployment/31338",
#         "snippet": "Jul 15, 2019 ... I'm trying to deploy an angular application dist folder to a server, but it appears that the .ssh directory does not exist in the deploy\u00a0...",
#         "summary": "I\u2019m trying to deploy an angular application dist folder to a server, but it appears that the .ssh directory does not exist in the deploy step of my project.\nDoes anyone know what I\u2019m doing wrong?\nI\u2019ve followed many topics from this forum but it still doesn\u2019t seem to work.\nMy config file is here: https://github.com/wafffly/closetr/blob/master/.circleci/config.ymlDeploy Job logs are here: https://circleci.com/gh/wafffly/closetr/165Please and Thank You"
#     },
#     {
#         "index": 6,
#         "url": "https://github.com/tiangolo/uvicorn-gunicorn-fastapi-docker/issues/19",
#         "snippet": "Oct 23, 2019 ... Description I have another project that utilizes fast api using gunicorn running uvicorn workers and supervisor to keep the api up.",     
#         "summary": "Have a question about this project?\nSign up for a free GitHub account to open an issue and contact its maintainers and the community.\nPick a username Email Address Password Sign up for GitHubBy clicking \u201cSign up for GitHub\u201d, you agree to our terms of service and privacy statement.\nWe\u2019ll occasionally send you account related emails.\nSign in to your account"
#     },
#     {
#         "index": 7,
#         "url": "https://developercommunity.visualstudio.com/t/attach-debugger-to-azure-app-service-not-working/933126",
#         "snippet": "Feb 27, 2020 ... Killed all of the msvmon.exe processes running in the Kudu Process explorer ... error and sometimes it works, I just retry if it fails). ... I don't\u00a0...",
#         "summary": "Sorry this browser is no longer supportedPlease use any other modern browser like 'Microsoft Edge'."
#     },
#     {
#         "index": 8,
#         "url": "https://github.com/pypa/packaging-problems/issues/573",
#         "snippet": "Feb 10, 2022 ... ... it outputted the same error. ... missing libffi-dev system package. ... error error: subprocess-exited-with-error \u00d7 Running setup.py install for\u00a0...",
#         "summary": "Have a question about this project?\nSign up for a free GitHub account to open an issue and contact its maintainers and the community.\nPick a username Email Address Password Sign up for GitHubBy clicking \u201cSign up for GitHub\u201d, you agree to our terms of service and privacy statement.\nWe\u2019ll occasionally send you account related emails.\nSign in to your account"
#     },
#     {
#         "index": 9,
#         "url": "https://answers.microsoft.com/en-us/windows/forum/all/the-requested-operation-requires-elevation/4d45a50e-2e5d-49f7-950c-e6281057491f",
#         "snippet": "The other names in that box reference my current system.) There are options to change permissions under \"Edit\" and \"Advanced\". I don't care\u00a0...", 
#         "summary": ""
#     }
# ]"""
#     logs = """Error occurred type=\"error\" text=\"Missing job runner for an existing job - #######\" stackTrace=\"   at Kudu.Core.Jobs.ContinuousJobsManager.EnableJob(String jobName)\r\n   at Kudu.Services.Jobs.JobsController.EnableContinuousJob(String jobName)\r\n   at lambda_method(Closure , Object , Object[] )\r\n   at System.Web.Http.Controllers.ReflectedHttpActionDescriptor.ActionExecutor.<>c__DisplayClass10.<GetExecutor>b__9(Object instance, Object[] methodParameters)"""
#     ranks = chat_agent.rank(outputs, logs)
#     print(ranks)