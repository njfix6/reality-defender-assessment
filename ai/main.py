
import whisper
from flask import Flask, jsonify, request
from minio import Minio


app = Flask(__name__)


client = Minio("play.min.io",
    access_key="Q3AM3UQ867SPQQA43P2F",
    secret_key="zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG",
    region="us-east-1",
)

@app.route("/speech-to-text", methods=['GET', 'POST'])
def speechToText():

    # get json body with audio file name
    content = request.json
    filename = content['filename']
    print("processing: " + filename)


    response = client.get_object("reality-defender-assessment-nick", "reality-defender-assessment-nick-"+filename)

    tempFile = "my-object.txt"

    with open(tempFile, "wb") as f:
        for data in response.stream():
            f.write(data)    

    model = whisper.load_model("base")
    result = model.transcribe(tempFile)
    text = result["text"]
    print(text)

    response.close()

    return jsonify({"text":text})


@app.route("/language", methods=['GET', 'POST'])
def language():

    # get json body with audio file name
    content = request.json
    filename = content['filename']
    print("processing: " + filename)


    response = client.get_object("reality-defender-assessment-nick", "reality-defender-assessment-nick-"+filename)

    tempFile = "my-object.txt"

    with open(tempFile, "wb") as f:
        for data in response.stream():
            f.write(data)  


    model = whisper.load_model("base")

    # load audio and pad/trim it to fit 30 seconds
    audio = whisper.load_audio(tempFile)
    audio = whisper.pad_or_trim(audio)

    # make log-Mel spectrogram and move to the same device as the model
    mel = whisper.log_mel_spectrogram(audio).to(model.device)

    # detect the spoken language
    _, probs = model.detect_language(mel)

    detectedLanguage = max(probs, key=probs.get)

    print(f"Detected language: {detectedLanguage}")

    return jsonify({"language":detectedLanguage})
  



if __name__ == '__main__':
    app.run(host="localhost", port=8000, debug=True)