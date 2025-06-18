import json
from config import LOG_FILE_PATH
from retriever.rag_chain import rag_chain

def analyze_logs():
    results = []
    logs = []
    with open(LOG_FILE_PATH, "r") as f:
        for line in f:
            logs.append(json.loads(line))

    for log_entry in logs:
        try:
            print("Hello world")
            print(log_entry)
            print("Proceeding")

            timestamp = log_entry.get("timestamp")
            error_message = log_entry.get("text_payload")

            prompt = f"""
                The following is an error log from the application at {timestamp}:

                {error_message}

                You have access to the application's codebase. Based on this log, suggest a specific code-level fix or change. Respond only with the corrected code snippet and any necessary comments to explain what was changed.
            """

            result = rag_chain.invoke({"query": prompt})
            suggestion = result["result"]
            source_docs = result["source_documents"]

            results.append({
                "timestamp": timestamp,
                "error_message": error_message,
                "suggested_fix": suggestion
            })
            print("******************")
            print("Suggestion : ")
            print(results)
        except Exception as e:
            results.append({"error": str(e), "raw_log": log_entry})
    return results
