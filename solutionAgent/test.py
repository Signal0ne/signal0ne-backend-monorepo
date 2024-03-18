import os
import json
import requests
import time
from datasets import load_dataset

### Use only if master dataset is not available
logset = [
    """Error occurred type=\"error\" text=\"Missing job runner for an existing job - #######\" stackTrace=\"   at Kudu.Core.Jobs.ContinuousJobsManager.EnableJob(String jobName)\r\n   at Kudu.Services.Jobs.JobsController.EnableContinuousJob(String jobName)\r\n   at lambda_method(Closure , Object , Object[] )\r\n   at System.Web.Http.Controllers.ReflectedHttpActionDescriptor.ActionExecutor.<>c__DisplayClass10.<GetExecutor>b__9(Object instance, Object[] methodParameters)""",
    """[11/18/2020, 11:56:23] [Nursery] Error: Unsupported state or unable to authenticate data\r\n    at Decipheriv.final (internal/crypto/cipher.js:172:29)\r\n    at Parser.decryptPayload (/usr/local/lib/node_modules/homebridge-mi-hygrothermograph/lib/parser.js:191:14)\r\n    at Parser.parse (/usr/local/lib/node_modules/homebridge-mi-hygrothermograph/lib/parser.js:59:12)\r\n    at Scanner.parseServiceData (/usr/local/lib/node_modules/homebridge-mi-hygrothermograph/lib/scanner.js:171:52)\r\n    at Scanner.onDiscover (/usr/local/lib/node_modules/homebridge-mi-hygrothermograph/lib/scanner.js:92:25)\r\n    at Noble.emit (events.js:322:22)\r\n    at Noble.onDiscover (/usr/local/lib/node_modules/homebridge-mi-hygrothermograph/node_modules/@abandonware/noble/lib/noble.js:196:10)\r\n    at NobleBindings.emit (events.js:310:20)\r\n    at NobleBindings.onDiscover (/usr/local/lib/node_modules/homebridge-mi-hygrothermograph/node_modules/@abandonware/noble/lib/hci-socket/bindings.js:169:10)\r\n    at Gap.emit (events.js:310:20)\r\n    at Gap.onHciLeAdvertisingReport (/usr/local/lib/node_modules/homebridge-mi-hygrothermograph/node_modules/@abandonware/noble/lib/hci-socket/gap.js:244:10)\r\n    at Hci.emit (events.js:310:20)\r\n    at Hci.processLeAdvertisingReport (/usr/local/lib/node_modules/homebridge-mi-hygrothermograph/node_modules/@abandonware/noble/lib/hci-socket/hci.js:656:12)\r\n    at Hci.processLeMetaEvent (/usr/local/lib/node_modules/homebridge-mi-hygrothermograph/node_modules/@abandonware/noble/lib/hci-socket/hci.js:612:10)\r\n    at Hci.onSocketData (/usr/local/lib/node_modules/homebridge-mi-hygrothermograph/node_modules/@abandonware/noble/lib/hci-socket/hci.js:483:12)\r\n    at BluetoothHciSocket.emit (events.js:310:20)""",
    """Traceback:\r\nFile \"/PATH/src/django/django/core/handlers/base.py\" in get_response\r\n  163.                 response = response.render()\r\nFile \"/PATH/src/django/django/template/response.py\" in render\r\n  156.             self.content = self.rendered_content\r\nFile \"/PATH/src/django/django/template/response.py\" in rendered_content\r\n  133.         content = template.render(context, self._request)\r\nFile \"/PATH/src/django/django/template/backends/django.py\" in render\r\n  83.         return self.template.render(context)\r\nFile \"/PATH/src/django/django/template/base.py\" in render\r\n  211.             return self._render(context)\r\nFile \"/PATH/src/django/django/template/base.py\" in _render\r\n  199.         return self.nodelist.render(context)\r\nFile \"/PATH/src/django/django/template/base.py\" in render\r\n  905.                 bit = self.render_node(node, context)\r\nFile \"/PATH/src/django/django/template/debug.py\" in render_node\r\n  80.             return node.render(context)\r\nFile \"/PATH/src/django/django/template/loader_tags.py\" in render\r\n  151.                 return template.render(context.new(values))\r\nFile \"/PATH/src/django/django/template/base.py\" in render\r\n  211.             return self._render(context)\r\nFile \"/PATH/src/django/django/template/base.py\" in _render\r\n  199.         return self.nodelist.render(context)\r\nFile \"/PATH/src/django/django/template/base.py\" in render\r\n  905.                 bit = self.render_node(node, context)\r\nFile \"/PATH/src/django/django/template/debug.py\" in render_node\r\n  80.             return node.render(context)\r\nFile \"/PATH/lib/python3.4/site-packages/formulation/templatetags/formulation.py\" in render\r\n  109.             'formulation': resolve_blocks(tmpl_name, safe_context),\r\nFile \"/PATH/lib/python3.4/site-packages/formulation/templatetags/formulation.py\" in resolve_blocks\r\n  34.         for block in template.nodelist.get_nodes_by_type(BlockNode)\r\n\r\nException Type: AttributeError at /\r\nException Value: 'Template' object has no attribute 'nodelist'""",
    """thread 'main' panicked at 'called `Result::unwrap()` on an `Err` value: Error(PrimaryScreenInfoError(147), State { next_error: None, backtrace: None })', /checkout/src/libcore/result.rs:906:4\r\nnote: Run with `RUST_BACKTRACE=1` for a backtrace.""",
    """09:24:50.210 [main] INFO  com.elastic.support.diagnostics.commands.GenerateManifestCmd - Writing diagnostic manifest.\r\n09:24:50.364 [main] INFO  com.elastic.support.diagnostics.commands.VersionCheckCmd - Getting Elasticsearch Version.\r\n09:24:50.815 [main] INFO  com.elastic.support.diagnostics.commands.DiagVersionCheckCmd - Checking for diagnostic version updates.\r\n09:24:51.163 [main] DIAG  com.elastic.support.diagnostics.chain.DiagnosticChainExec - Error encountered running diagnostic. See logs for additional information.  Exiting application.\r\njava.lang.NumberFormatException: For input string: \"\"\r\n\tat java.lang.NumberFormatException.forInputString(NumberFormatException.java:65) ~[?:?]\r\n\tat java.lang.Integer.parseInt(Integer.java:662) ~[?:?]\r\n\tat java.lang.Integer.parseInt(Integer.java:770) ~[?:?]\r\n\tat com.elastic.support.diagnostics.commands.RunClusterQueriesCmd.buildStatementsByVersion(RunClusterQueriesCmd.java:37) ~[support-diagnostics-7.0.6.jar:7.0.6]\r\n\tat com.elastic.support.diagnostics.commands.RunClusterQueriesCmd.execute(RunClusterQueriesCmd.java:28) ~[support-diagnostics-7.0.6.jar:7.0.6]\r\n\tat com.elastic.support.diagnostics.chain.Chain.execute(Chain.java:33) ~[support-diagnostics-7.0.6.jar:7.0.6]\r\n\tat com.elastic.support.diagnostics.chain.DiagnosticChainExec.runDiagnostic(DiagnosticChainExec.java:18) [support-diagnostics-7.0.6.jar:7.0.6]\r\n\tat com.elastic.support.diagnostics.DiagnosticService.exec(DiagnosticService.java:57) [support-diagnostics-7.0.6.jar:7.0.6]\r\n\tat com.elastic.support.diagnostics.DiagnosticApp.main(DiagnosticApp.java:31) [support-diagnostics-7.0.6.jar:7.0.6]"""
]

dataset_url = "Signal0ne/logs-for-evaluation"
test_output_dir_name = 'output-test-results'
url = "http://localhost:8081/run_analysis"
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