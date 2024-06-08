FROM python:3.9-slim
WORKDIR /app
COPY requirements.txt .
COPY ./dist/ . 
RUN pip install --no-cache-dir -r requirements.txt && \
pip install opentelemetry_api-1.26.0.dev0-py3-none-any.whl && \
pip install opentelemetry_semantic_conventions-0.47b0.dev0-py3-none-any.whl && \
pip install opentelemetry_sdk-1.26.0.dev0-py3-none-any.whl && \
pip install opentelemetry_instrumentation-0.47b0.dev0-py3-none-any.whl && \
pip install opentelemetry_distro-0.47b0.dev0-py3-none-any.whl  && \
pip install opentelemetry_proto-1.26.0.dev0-py3-none-any.whl && \
pip install opentelemetry_exporter_otlp_proto_common-1.26.0.dev0-py3-none-any.whl && \
pip install opentelemetry_exporter_otlp_proto_http-1.26.0.dev0-py3-none-any.whl && \
pip install opentelemetry_exporter_otlp_proto_grpc-1.26.0.dev0-py3-none-any.whl && \
pip install opentelemetry_util_http-0.47b0.dev0-py3-none-any.whl && \
pip install opentelemetry_instrumentation_asgi-0.47b0.dev0-py3-none-any.whl && \
pip install opentelemetry_instrumentation_fastapi-0.47b0.dev0-py3-none-any.whl && \
pip install opentelemetry_instrumentation_requests-0.47b0.dev0-py3-none-any.whl && \
pip install opentelemetry_instrumentation_langchain-0.47b0.dev0-py3-none-any.whl && \
opentelemetry-bootstrap -a install
COPY . .
EXPOSE 8081
CMD [ \
"opentelemetry-instrument", \
"--traces_exporter", "otlp", \
"uvicorn", "server:app", \
# "--reload", "--reload-dir", "./" , \
"--host", "0.0.0.0", "--port", "8081"]
