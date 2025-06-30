import os
from langchain.text_splitter import RecursiveCharacterTextSplitter
from langchain.docstore.document import Document
from config import CODEBASE_PATH
from codebase.repo_loader import prepare_codebase

def load_codebase_as_docs():
    prepare_codebase()
    docs = []
    splitter = RecursiveCharacterTextSplitter(chunk_size=512, chunk_overlap=50)
    for root, _, files in os.walk(CODEBASE_PATH):
        for f in files:
            file_path = os.path.join(root, f)
            try:
                with open(file_path, "r", encoding="utf-8") as file:
                    text = file.read()
                    splits = splitter.split_text(text)
                    docs.extend([Document(page_content=chunk, metadata={"filename": file_path}) for chunk in splits])
            except Exception:
                continue
    return docs
