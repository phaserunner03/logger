from services.log_processor import analyze_logs
from retriever.rag_chain import retriever
from utils.parser import extract_code_block
from utils.common import sanitize_filename
from flask import Flask,request,jsonify


app = Flask(__name__)


@app.route("/suggest-fix",methods=["POST"])
def suggest_fix():
    results = analyze_logs()
    print("---------------------------")

    for idx,entry in enumerate(results):
        code_response = entry.get("suggested_fix","").strip()
        timestamp = sanitize_filename(entry.get("timestamp", f"fix_{idx}"))
        code = extract_code_block(code_response)
        if not code:
            print(f"⛔ Skipping entry #{idx+1}: No code found.")
            continue        
        return jsonify({
        "timestamp": timestamp,
        "suggested_fix": code
    })

    
if __name__ == "__main__":
    print("port starting")
    app.run(port=8080, debug=True)

