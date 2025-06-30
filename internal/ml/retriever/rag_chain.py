from langchain_google_genai import ChatGoogleGenerativeAI
from langchain_chroma import Chroma
from langchain.chains import RetrievalQA
import os

from retriever.embedder import embeddings
from codebase.processor import load_codebase_as_docs
from config import LLM_MODEL

PERSISTENT_VECTORSTORE_PATH = "./internal/data"

llm = ChatGoogleGenerativeAI(model=LLM_MODEL, temperature=1, max_tokens=1000)
code_docs = load_codebase_as_docs()

print("Vectorstore loaded and retriever created.")
if os.path.exists(PERSISTENT_VECTORSTORE_PATH):
    vectorstore = Chroma(persist_directory=PERSISTENT_VECTORSTORE_PATH, embedding_function=embeddings)
else:
    vectorstore = Chroma.from_documents(
        documents=code_docs,
        embedding=embeddings,
        persist_directory=PERSISTENT_VECTORSTORE_PATH
    )


retriever = vectorstore.as_retriever(search_type="similarity",search_kwargs={"k": 10})    
rag_chain = RetrievalQA.from_chain_type(
llm=llm,
retriever=retriever,
chain_type="stuff",
return_source_documents=True
)

#what if there are mutliple changes in the multiple files 