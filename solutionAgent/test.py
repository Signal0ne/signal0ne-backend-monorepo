import os
import json
import requests
import time
from datasets import load_dataset

### Use only if master dataset is not available
logset = [
    """09:24:50.210 [main] INFO  com.elastic.support.diagnostics.commands.GenerateManifestCmd - Writing diagnostic manifest.\r\n09:24:50.364 [main] INFO  com.elastic.support.diagnostics.commands.VersionCheckCmd - Getting Elasticsearch Version.\r\n09:24:50.815 [main] INFO  com.elastic.support.diagnostics.commands.DiagVersionCheckCmd - Checking for diagnostic version updates.\r\n09:24:51.163 [main] DIAG  com.elastic.support.diagnostics.chain.DiagnosticChainExec - Error encountered running diagnostic. See logs for additional information.  Exiting application.\r\njava.lang.NumberFormatException: For input string: \"\"\r\n\tat java.lang.NumberFormatException.forInputString(NumberFormatException.java:65) ~[?:?]\r\n\tat java.lang.Integer.parseInt(Integer.java:662) ~[?:?]\r\n\tat java.lang.Integer.parseInt(Integer.java:770) ~[?:?]\r\n\tat com.elastic.support.diagnostics.commands.RunClusterQueriesCmd.buildStatementsByVersion(RunClusterQueriesCmd.java:37) ~[support-diagnostics-7.0.6.jar:7.0.6]\r\n\tat com.elastic.support.diagnostics.commands.RunClusterQueriesCmd.execute(RunClusterQueriesCmd.java:28) ~[support-diagnostics-7.0.6.jar:7.0.6]\r\n\tat com.elastic.support.diagnostics.chain.Chain.execute(Chain.java:33) ~[support-diagnostics-7.0.6.jar:7.0.6]\r\n\tat com.elastic.support.diagnostics.chain.DiagnosticChainExec.runDiagnostic(DiagnosticChainExec.java:18) [support-diagnostics-7.0.6.jar:7.0.6]\r\n\tat com.elastic.support.diagnostics.DiagnosticService.exec(DiagnosticService.java:57) [support-diagnostics-7.0.6.jar:7.0.6]\r\n\tat com.elastic.support.diagnostics.DiagnosticApp.main(DiagnosticApp.java:31) [support-diagnostics-7.0.6.jar:7.0.6]"""
]

dataset_url = "Signal0ne/logs-for-evaluation"
test_output_dir_name = 'output-test-results'
url = "http://localhost:8000/run_analysis"
results = []

# dataset = load_dataset(dataset_url, split=None)
# logset = dataset['train']['logs']

# limit the number of logs
logset = logset[:1]

if not os.path.exists(test_output_dir_name):
    os.makedirs(test_output_dir_name)


for log in logset:
    data = {
        "logs": log,
    }
    response = None
    while response is None:
        response = requests.post(url, json=data)

    res = {
        "log": log,
        "result": response.json()
        }
    
    results.append(res)
    time.sleep(1)

with open(f"{test_output_dir_name}/results.json", "w") as f:
    json.dump(results, f, indent=4)