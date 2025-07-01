
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
from utils.parser import extract_code_block
from utils.io import save_code
from retriever.enhancement_loop import EnhancementLoop
from retriever.embedder import embeddings

from flask import Flask,request,jsonify
from config import LLM_MODEL
from langchain_google_genai import ChatGoogleGenerativeAI

app = Flask(__name__)


# 1) Instantiate your LLM and EnhancementLoop once
llm = ChatGoogleGenerativeAI(model=LLM_MODEL, temperature=1, max_tokens=1000)
loop = EnhancementLoop(embedding_model=embeddings, llm=llm)

@app.route("/suggest-fix", methods=["POST"])
def suggest_fix():
    # 2) Extract the error_message payload
    payload = request.json or {}
    error_message = payload.get("error_message", "")
    if not error_message:
        return jsonify({"error": "No `error_message` provided"}), 400

    # 3) Run the enhancement loop
    out = loop.run(error_message)

    # 4) If there was an error or nothing passed the threshold, just return that
    if out.get("error"):
        return jsonify(out), 200

    # 5) Otherwise, parse the suggestion, extract code, and save files exactly as before
    suggestion = out["suggestion"]
    # `extract_code_block` returns (code, filename_in_block) if the AI followed the format
    code, file_from_block = extract_code_block(suggestion)

    # Fallback: if the block didn’t parse, pick the first matched doc’s filename
    if not code:
        # take the first match from out["matched_docs"]
        file_to_save = out["matched_docs"][0]["filename"]
        code = suggestion
    else:
        file_to_save = file_from_block

    # normalize and save
    file_to_save = file_to_save.strip()
    path = save_code(code, file_to_save)

    # 6) Return a summary JSON
    return jsonify({
        "query":        out["query"],
        "matched_docs": out["matched_docs"],
        "saved_path":   path,
        "explanation":  suggestion.split("\n")[-1]  # if you included a “**Explanation:**” line
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
