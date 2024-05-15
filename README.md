# Reality Defender Assessment

## Architecture

![Architecture](./architecture.png)

## Running Local Docker Compose

```
docker compose build
docker compose up -d
```

## API

#### Creating a user

```
curl http://localhost:8080/create-user \
    --include \
    --header "Content-Type: application/json" \
    --request "POST" \
    --data '{"username": "test"}'
```

#### Uploading a file

```
curl -X POST http://localhost:8080/upload \
  -F "file=@/Users/nicholasfix/dev/reality-defender-assessment/server/test/test.txt" \
  -H "Content-Type: multipart/form-data"
```

#### Uploading a file

```
curl -X POST http://localhost:8080/upload \
  -F "file=@/Users/nicholasfix/dev/reality-defender-assessment/server/test/test.txt" \
  -H "Content-Type: multipart/form-data"
```

#### Calling Speech to text

Call socket endpoint

```
ws://localhost:8080/process/speech-to-text?filename=<filename>
```

#### Calling Language Detection

Call socket endpoint

```
ws://localhost:8080/process/language?filename=<filename>
```
