import os
import json
import time
import numpy as np
from flask import Flask, request, jsonify
from google.cloud import bigquery, aiplatform
from langchain.vectorstores import Chroma
from langchain.text_splitter import RecursiveCharacterTextSplitter

from langchain.chains import RetrievalQA
from langchain.docstore.document import Document
from langchain_google_genai import GoogleGenerativeAIEmbeddings
from langchain_chroma import Chroma
from langchain_google_genai import ChatGoogleGenerativeAI
from langchain.chains import create_retrieval_chain
from langchain.chains.combine_documents import create_stuff_documents_chain
from langchain_core.prompts import ChatPromptTemplate


app = Flask(__name__)

PROJECT_ID = "meet-and-media-sync"
REGION = "us-central1"
BIGQUERY_DATASET="logging"
BIGQUERY_TABLE_ID="log_table"
BATCH_SIZE = 5
CODEBASE_PATH = "../../codebase"  
LOG_FILE_PATH = "../../codebase/buggy_app/backend/error_logs.txt"
from dotenv import load_dotenv
load_dotenv()
# === EMBEDDINGS + LLM ===
embeddings = GoogleGenerativeAIEmbeddings(model="models/embedding-001")
llm = ChatGoogleGenerativeAI(model="gemini-2.0-flash", temperature=1, max_tokens=1000)

# === Load codebase and build vectorstore ===
def load_codebase_as_docs():
    docs = []
    splitter = RecursiveCharacterTextSplitter(chunk_size=512, chunk_overlap=50)
    for root, _, files in os.walk(CODEBASE_PATH):
        for f in files:
            try:
                with open(LOG_FILE_PATH, "r", encoding="utf-8") as file:
                    text = file.read()
                    splits = splitter.split_text(text)
                    docs.extend([Document(page_content=chunk, metadata={"filename": LOG_FILE_PATH}) for chunk in splits])
            except UnicodeDecodeError:
                print(f"Skipping file {f} due to encoding issues.")
    return docs

print("[+] Indexing codebase...")
code_docs = load_codebase_as_docs()
vectorstore = Chroma.from_documents(documents=code_docs, embedding=embeddings)
retriever = vectorstore.as_retriever(search_type="similarity", search_kwargs={"k": 10})
rag_chain = RetrievalQA.from_chain_type(llm=llm, retriever=retriever, chain_type="stuff", return_source_documents=True)

# === Read logs from file and process ===
@app.route("/process_logs", methods=["GET"])
def process_logs():
    try:
        with open(LOG_FILE_PATH, "r") as f:
            logs = f.readlines()

        results = []
        for log_line in logs:
            try:
                log_entry = json.loads(log_line)
                error_message = log_entry.get("text_payload") or str(log_entry.get("json_payload"))
                timestamp = log_entry.get("timestamp")

                prompt = f"""An error occurred at {timestamp}:
{error_message}

Based on the codebase, what is the fix? Provide only code."""
                suggestion = rag_chain.run(prompt)
                results.append({
                    "timestamp": timestamp,
                    "error_message": error_message,
                    "suggested_fix": suggestion
                })
            except Exception as e:
                results.append({"error": str(e), "raw_log": log_line})

        return jsonify(results)
    except Exception as e:
        return jsonify({"status": "error", "error": str(e)}), 500

if __name__ == "__main__":
    app.run(port=8080, debug=True)
