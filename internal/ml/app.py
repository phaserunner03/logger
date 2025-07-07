from flask import Flask, request, jsonify
from retriever.enhancement_loop import EnhancementLoop
from retriever.embedder import embeddings
from utils.parser import parse_response

from config import GEMINI_LLM, CLAUDE_LLM, OPENAI_LLM, PROJECT_ID, REGION

from langchain_google_genai import ChatGoogleGenerativeAI

from anthropic import AnthropicVertex  

import os 
os.environ["GOOGLE_APPLICATION_CREDENTIALS"] = "/Users/bhavya.shah/Documents/Go/logger/key.json"
 
app = Flask(__name__)
shared_embeddings = embeddings

llm_gemini = ChatGoogleGenerativeAI(model=GEMINI_LLM, temperature=0, max_tokens=3072)
loop_gemini = EnhancementLoop(embedding_model=shared_embeddings, llm=llm_gemini)

claude_client = AnthropicVertex(project_id=PROJECT_ID, region="us-east5")  

def process_claude_request(error_message: str):
    if not error_message:
        return jsonify({"error": "No `error_message` provided"}), 400

    try:
        response = claude_client.messages.create(
            model=CLAUDE_LLM,
            max_tokens=3072,
            messages=[
                {"role": "user", "content": error_message}
            ],
        )
        raw_text = response.content
        print(f"Raw response from Claude LLM: {raw_text}")
        parsed = parse_response(raw_text)
    except Exception as e:
        return jsonify({"error": f"Claude error: {str(e)}"}), 500

    return jsonify({
        "filename": parsed["filename"].strip() or "unknown_file.js",
        "changes": parsed["changes"].strip(),
        "explanation": parsed["explanation"].strip() or "Warning: LLM output might be incomplete.",
        "raw_text": raw_text,
    }), 200

@app.route("/suggest-fix/gemini", methods=["POST"])
def suggest_fix_gemini():
    return process_fix_request(loop_gemini, request.json.get("error_message", ""))

@app.route("/suggest-fix/claude", methods=["POST"])
def suggest_fix_claude():
    return process_claude_request(request.json.get("error_message", ""))  

# @app.route("/suggest-fix/openai", methods=["POST"])
# def suggest_fix_openai():
#     return process_fix_request(loop_openai, request.json.get("error_message", ""))

@app.route("/health", methods=["GET"])
def health_check():
    return jsonify({"status": "ok"}), 200

def process_fix_request(loop, error_message: str):
    if not error_message:
        return jsonify({"error": "No `error_message` provided"}), 400

    raw_text = loop.run(error_message)
    print(f"Raw response from LLM: {raw_text}")

    try:
        parsed = parse_response(raw_text)
    except Exception as e:
        return jsonify({"error": f"Invalid LLM response: {str(e)}"}), 500

    return jsonify({
        "filename": parsed["filename"].strip() or "unknown_file.js",
        "changes": parsed["changes"].strip(),
        "explanation": parsed["explanation"].strip() or "Warning: LLM output might be incomplete.",
        "raw_text": raw_text,
    }), 200

if __name__ == "__main__":
    print("port starting")
    app.run(port=8080, debug=True)
