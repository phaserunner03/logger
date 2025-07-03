
# @app.route("/suggest-fix",methods=["POST"])
# def suggest_fix():
#     log = request.json.get("error_message", None)
#     print(log)
#     results = analyze_logs(log)
#     print("---------------------------")

#     for idx,entry in enumerate(results):
#         code_response = entry.get("suggested_fix","").strip()
#         filename = entry.get("affected_file", "unknown_file.js")
       
#         code,file= extract_code_block(code_response)
#         if not code:
#             print(f"Skipping entry #{idx+1}: No code found.")
#             continue
#         print(file)
#         filename = filename.strip()
#         path = save_code(code, filename)
#         print(f"[✅] Saved: {path}")
#     return jsonify({"message": "Fix suggestions processed successfully"}), 200

# # instantiate once
# # llm = ChatGoogleGenerativeAI(model=LLM_MODEL, temperature=1, max_tokens=1000)
# # loop = EnhancementLoop(embedding_model=embeddings, llm=llm)

# # @app.route("/suggest-fix", methods=["POST"])
# # def suggest_fix():
# #     payload = request.json or {}
# #     error_message = payload.get("error_message", "")
# #     if not error_message:
# #         return jsonify({"error": "No `error_message` provided"}), 400

# #     out = loop.run(error_message)
# #     return jsonify(out), 200
    


# app.py
from services.log_processor import analyze_logs
from retriever.rag_chain import retriever
from retriever.enhancement_loop import EnhancementLoop
from retriever.embedder import embeddings
from utils.parser import parse_response

from flask import Flask,request,jsonify
from config import LLM_MODEL
from langchain_google_genai import ChatGoogleGenerativeAI

app = Flask(__name__)


# 1) Instantiate your LLM and EnhancementLoop once
llm = ChatGoogleGenerativeAI(model=LLM_MODEL, temperature=1, max_tokens=3072)
loop = EnhancementLoop(embedding_model=embeddings, llm=llm)

@app.route("/suggest-fix", methods=["POST"])
def suggest_fix():
    payload = request.json or {}
    error_message = payload.get("error_message", "")
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
   
@app.route("/health", methods=["GET"])
def health_check():
    """
    Health check endpoint to verify if the service is running.
    """
    return jsonify({"status": "ok"}), 200

if __name__ == "__main__":
    print("port starting")
    app.run(port=8080, debug=True)
