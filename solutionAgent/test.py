import os
import json
import requests
import time
from datasets import load_dataset

### Use only if master dataset is not available
logset = [
    """2024/03/25 17:18:11 Error while building: \n # github.com/go-openapi/swag \n /go/pkg/mod/github.com/go-openapi/swag@v0.23.0/string_bytes.go:7:29: undefined: unsafe.StringData\n note: module requires Go 1.20""",
    """W0322 22:21:02.781142 1 watcher.go:338] watch chan error: etcdserver: mvcc: required revision has been compacted\n W0322 22:21:08.017022 1 watcher.go:338] watch chan error: etcdserver: mvcc: required revision has been compacted""",
    """1710880233: New connection from 172.17.0.1:35838 on port 1883.\n1710880233: Client <unknown> disconnected due to protocol error.\n1710880233: New connection from 172.17.0.1:35854 on port 1883.\n1710880233: Client <unknown> disconnected due to protocol error""",
    """[2024/03/19 06:58:38] [provider.go:55] Performing OIDC Discovery...\n [2024/03/19 06:58:38] [main.go:60] ERROR: Failed to initialise OAuth2 Proxy: error initialising provider: could not create provider data: error building OIDC ProviderVerifier: could not get verifier builder: error while discovery OIDC configuration: failed to discover OIDC configuration: error performing request: Get \"http://keycloak:7810/realms/bionic-gpt/.well-known/openid-configuration\": dial tcp: lookup keycloak on 127.0.0.11:53: no such host"""
]

dataset_url = "Signal0ne/logs-for-evaluation"
test_output_dir_name = 'output-test-results'
url = "http://localhost:8000/run_analysis"
results = []

# dataset = load_dataset(dataset_url, split=None)
# logset = dataset['train']['logs']

# limit the number of logs
logset = logset

if not os.path.exists(test_output_dir_name):
    os.makedirs(test_output_dir_name)


for log in logset:
    data = {
        "logs": log,
    }
    response = None
    while response is None:
        response = requests.post(url, json=data)

    print(response)
    res = {
        "log": log,
        "result": response.json()
        }
    
    results.append(res)
    time.sleep(1)

with open(f"{test_output_dir_name}/results.json", "w") as f:
    json.dump(results, f, indent=4)