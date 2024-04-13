# docker-signalone

## Overview
Signal0ne docker extension is a tool for debugging and monitoring containerized apps, which enables automated insights about failed containers and containers affected by resource usage anomalies.


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
make --directory=./backend start-reload-backend

```

For production: 
```
make --directory=./backend prod-build-backend

```
```
make --directory=./backend start-backend
```
OR
```
make --directory=./backend start-backend-with-init # to start backend with init sample development data
```