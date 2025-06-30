from services.log_processor import analyze_logs
from retriever.rag_chain import retriever
from utils.parser import extract_code_block
from utils.io import save_code

from flask import Flask,request,jsonify


app = Flask(__name__)


@app.route("/suggest-fix",methods=["POST"])
def suggest_fix():
    log = request.json.get("error_message", None)
    print(log)
    results = analyze_logs(log)
    print("---------------------------")

    for idx,entry in enumerate(results):
        code_response = entry.get("suggested_fix","").strip()
        filename = entry.get("affected_file", "unknown_file.js")
       
        code,file= extract_code_block(code_response)
        if not code:
            print(f"Skipping entry #{idx+1}: No code found.")
            continue
        print(file)
        filename = filename.strip()
        path = save_code(code, filename)
        print(f"[✅] Saved: {path}")
    return jsonify({"message": "Fix suggestions processed successfully"}), 200

    
if __name__ == "__main__":
    print("port starting")
    app.run(port=8080, debug=True)

