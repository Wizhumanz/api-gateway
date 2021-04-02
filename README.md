# API Gateway, Anastasia Trading Bot

API gateway to receive web app client requests and forward to other services.

## Local Dev

```
chmod +x run.sh
./run.sh
```

Inside `run.sh`:
```
export PORT=8000

# build/ directory ignored by git
go build -o build/api .

build/api
```


`PORT` env var must be passed for local dev. This env var is present by default in Cloud Run production environment.

### Docker

```
cd api
docker build -t <img-name> .
docker run -e AUTH=password -e PORT=8000 --name <container-name> -p 8000:8000 <img-name>
```

### [GCP Datastore testing](https://cloud.google.com/datastore/docs/reference/libraries#client-libraries-install-go):

1. Must authenticate: `export GOOGLE_APPLICATION_CREDENTIALS="/path/to/auth/my-key.json"` in current shell session.
