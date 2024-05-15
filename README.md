```
curl -X POST http://localhost:8080/upload \
  -F "file=@/Users/nicholasfix/dev/reality-defender-assessment/server/test/test.txt" \
  -H "Content-Type: multipart/form-data"
```

```

curl http://localhost:8080/create-user \
    --include \
    --header "Content-Type: application/json" \
    --request "POST" \
    --data '{"username": "test"}'

```

```
curl ws://localhost:8080/process/text-to-string

```

curl -X POST http://localhost:8080/upload \
 -H "Content-Type: multipart/mixed" \
 -F "file=@/Users/nicholasfix/dev/reality-defender-assessment/server/test/test.txt" \
 -F 'metadata={"username": "test2"}'

TODO:
