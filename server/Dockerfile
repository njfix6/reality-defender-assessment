FROM golang:1.20

WORKDIR /app

COPY go.mod go.sum ./
COPY *.go ./


RUN CGO_ENABLED=0 GOOS=linux go build -o ./server


ENV DB_HOST=host.docker.internal
ENV DB_SERVER=server
ENV DB_PASSWORD=server
ENV DB_NAME=server
ENV DB_PORT=5432

EXPOSE 8080

CMD ["./server"]
