import os
from langchain.text_splitter import RecursiveCharacterTextSplitter
from langchain.docstore.document import Document
from config import CODEBASE_PATH
from codebase.repo_loader import prepare_codebase
from datetime import datetime
# —– Main loader with expanded metadata —–
def load_codebase_as_docs():
    prepare_codebase()
    docs = []
    splitter = RecursiveCharacterTextSplitter(chunk_size=512, chunk_overlap=50)

    for root, _, files in os.walk(CODEBASE_PATH):
        for fname in files:
            

            path = os.path.join(root, fname)
            try:
                # read the full file once
                with open(path, "r", encoding="utf-8") as f:
                    full_text = f.read()

                # metadata fields
                metadata = {
                    "filename": path,
                    "full_code": full_text,
                    "last_modified": datetime.fromtimestamp(
                        os.path.getmtime(path)
                    ).isoformat(),
                    "language": os.path.splitext(fname)[1].lstrip("."),
                }

                # split into chunks, but carry over the same metadata
                for chunk in splitter.split_text(full_text):
                    docs.append(Document(page_content=chunk, metadata=metadata))

            except Exception as err:
                print(f"Skipping {path}: {err}")
                continue

    print(f"Loaded {len(docs)} code chunks with metadata")
    return docs
# def load_codebase_as_docs():
    # prepare_codebase()
    # docs = []
    # splitter = RecursiveCharacterTextSplitter(chunk_size=512, chunk_overlap=50)
    # for root, _, files in os.walk(CODEBASE_PATH):
    #     for f in files:
    #         file_path = os.path.join(root, f)
    #         try:
    #             with open(file_path, "r", encoding="utf-8") as file:
    #                 text = file.read()
    #                 splits = splitter.split_text(text)
    #                 docs.extend([Document(page_content=chunk, metadata={"filename": file_path}) for chunk in splits])
    #         except Exception:
    #             continue
    # return docs
