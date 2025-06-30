import json
from config import LOG_FILE_PATH
from retriever.rag_chain import rag_chain

def analyze_logs():
    results = []
    logs = []
    message_history= []
    with open(LOG_FILE_PATH, "r") as f:
        for line in f:
            logs.append(json.loads(line))
    
    for idx,log_entry in enumerate(logs):
        try:
            print(f"Processing log entry {idx + 1}/{len(logs)}")
            timestamp = log_entry.get("timestamp")
            error_message = log_entry.get("text_payload")
            
            message_history.append({
                "timestamp": timestamp,
                "error_message": error_message
            })
            prior_context = ""
            for i, h in enumerate(message_history[-3:]): 
                prior_context += f"### Previous Error {i+1} ({h['timestamp']}):\n```\n{h['error_message']}\n```\n\n"


            
            prompt = f"""
            You are a code assistant with access to the full source code of the application.
            Below is a recent error log from the application:
            Timestamp: {timestamp}
            Error Message: {error_message}
            
            {prior_context if prior_context else ""}
            Based on the codebase and the context above, suggest a detailed and accurate code level fix.
            Respond with the corrected code and complete file with the fix applied. 
            Make sure the code is valid, handle edge cases, and is well-structured.
            
            Give me the filepath where the fix should be applied.
            Output format:
            ```<language>
            <corrected_code>
            ```
            filename: <path_to_file>
            """
            
            result = rag_chain.invoke({"query": prompt})
            suggestion = result["result"]
            source_docs = result["source_documents"]
            print(f"source_docs: {len(source_docs)}")
            for doc in source_docs:
                print(f"Source Document: {doc.metadata.get('filename', 'unknown')}")
                
            file_name = source_docs[0].metadata.get("filename")

            results.append({
                "timestamp": timestamp,
                "error_message": error_message,
                "suggested_fix": suggestion,
                "affected_file": file_name,
            })
            
        except Exception as e:
            print(f"Error processing log entry {idx + 1}: {str(e)}")
            results.append({"error": str(e), "raw_log": log_entry})
    return results
