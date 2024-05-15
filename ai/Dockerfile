FROM python:3.11-alpine



RUN python -m pip install openai-whisper

RUN sudo apt update && sudo apt install ffmpeg 

ENTRYPOINT [ "pipenv run flask --app main run" ]