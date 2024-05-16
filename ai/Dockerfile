FROM python:3.12.3

RUN apt update && apt install ffmpeg -y

RUN pip install pipenv
RUN pip install Flask
RUN pip install whisper
RUN pip install openai-whisper
RUN pip install minio

WORKDIR /app

COPY Pipfile ./
COPY Pipfile.lock ./

COPY main.py ./

CMD ["pipenv", "run", "python", "main.py"]
