# docker-signalone

## Overview
Signal0ne docker extension is a tool for debugging and monitoring containerized apps, which enables automated insights about failed containers and containers affected by resource usage anomalies.

## Demo video
[![YouTube](http://i.ytimg.com/vi/DYkBbBCSm1Y/hqdefault.jpg)](https://www.youtube.com/watch?v=DYkBbBCSm1Y)

[VsCode Extension](https://github.com/Signal0ne/signal0ne-vscode-extension)

[Docker Extension](https://github.com/Signal0ne/signal0ne-docker-extension)

## How to use locally

### Prerequisites
- Docker with compose
- Docker Desktop client
- Make

### Env variables
```
cp backend/.env.template backend/.default.env
# In backend/.default.env replace _APIKEY_ with your Huggingface API key 
# Adjust other variables if needed(optional)

```

### Backend

For development (using hot reload) : 
```
make --directory=./backend dev-build-backend
```

```
make --directory=./backend start-backend

```

For production: 
```
make --directory=./backend prod-build-backend

```
```
make --directory=./backend deploy-backend
```
