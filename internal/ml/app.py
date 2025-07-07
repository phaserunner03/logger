from retriever.enhancement_loop import EnhancementLoop
from retriever.embedder import embeddings
from utils.parser import parse_response

from flask import Flask,request,jsonify
from config import GEMINI_LLM,CLAUDE_LLM,OPEN_AI_LLM 
from langchain_google_genai import ChatGoogleGenerativeAI
from langchain_anthropic  import ChatAnthropic
from langchain_openai import ChatOpenAI 



app = Flask(__name__)

shared_embeddings = embeddings
llm_google = ChatGoogleGenerativeAI(model=GEMINI_LLM, temperature=1, max_tokens=3072)
llm_claude = ChatAnthropic(model=CLAUDE_LLM, temperature=1, max_tokens=3072)
llm_openai = ChatOpenAI(model=OPEN_AI_LLM, temperature=1, max_tokens=3072)

loop_google = EnhancementLoop(embedding_model=shared_embeddings, llm=llm_google)
loop_claude = EnhancementLoop(embedding_model=shared_embeddings, llm=llm_claude)
loop_openai = EnhancementLoop(embedding_model=shared_embeddings, llm=llm_openai)

def process_fix_request(loop, error_message: str):
    if not error_message:
        return jsonify({"error": "No `error_message` provided"}), 400

    raw_text = loop.run(error_message)
    print(f"Raw response from LLM: {raw_text}")
    print(f"Type of raw_text: {type(raw_text)}")

    try:
        parsed = parse_response(raw_text)
    except Exception as e:
        return jsonify({"error": f"Invalid LLM response: {str(e)}"}), 500

    filename = parsed["filename"].strip() or "unknown_file.js"
    changes = parsed["changes"].strip()
    explanation = parsed["explanation"].strip() or "Warning: LLM output might be incomplete."

    return jsonify({
        "filename": filename,
        "changes": changes,
        "explanation": explanation,
        "raw_text": raw_text,
    }), 200


@app.route("/suggest-fix/gemini", methods=["POST"])
def suggest_fix_google():
    error_message = (request.json or {}).get("error_message", "")
    return process_fix_request(loop_google, error_message)


@app.route("/suggest-fix/claude", methods=["POST"])
def suggest_fix_claude():
    error_message = (request.json or {}).get("error_message", "")
    return process_fix_request(loop_claude, error_message)


@app.route("/suggest-fix/openai", methods=["POST"])
def suggest_fix_openai():
    error_message = (request.json or {}).get("error_message", "")
    return process_fix_request(loop_openai, error_message)


@app.route("/health", methods=["GET"])
def health_check():
    return jsonify({"status": "ok"}), 200


if __name__ == "__main__":
    print("Starting server on port 8080...")
    app.run(port=8080, debug=True)