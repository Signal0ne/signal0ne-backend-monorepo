FROM python:3.9-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
EXPOSE 8081
CMD ["uvicorn", "server:app","--workers", "4", "--host", "0.0.0.0", "--port", "8081"]
