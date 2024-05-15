
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

    # try:
    #     response = client.get_object("reality-defender-assessment-nick", "reality-defender-assessment-nick-" + filename)
    #     print("reponse", response)
    # # Read data from response.
    # finally:
    #     response.close()
    #     response.release_conn()


 

  
    
    model = whisper.load_model("base")
    result = model.transcribe(filename)
    text = result["text"]
    print(text)

    return jsonify({"text":text})


@app.route("/language", methods=['GET', 'POST'])
def language():

    # get json body with audio file name
    content = request.json
    filename = content['filename']
    print("processing: " + filename)

    model = whisper.load_model("base")

    # load audio and pad/trim it to fit 30 seconds
    audio = whisper.load_audio(filename)
    audio = whisper.pad_or_trim(audio)

    # make log-Mel spectrogram and move to the same device as the model
    mel = whisper.log_mel_spectrogram(audio).to(model.device)

    # detect the spoken language
    _, probs = model.detect_language(mel)

    detectedLanguage = max(probs, key=probs.get)

    print(f"Detected language: {detectedLanguage}")

    return jsonify({"language":detectedLanguage})
  


def main():
    print("Hello World!")

if __name__ == "__main__":
    main()