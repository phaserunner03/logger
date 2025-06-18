from langchain_google_genai import ChatGoogleGenerativeAI
from langchain_chroma import Chroma
from langchain.chains import RetrievalQA

from retriever.embedder import embeddings
from codebase.processor import load_codebase_as_docs
from config import LLM_MODEL

llm = ChatGoogleGenerativeAI(model=LLM_MODEL, temperature=1, max_tokens=1000)

code_docs = load_codebase_as_docs()
vectorstore = Chroma.from_documents(documents=code_docs, embedding=embeddings)
retriever = vectorstore.as_retriever(search_type="similarity", search_kwargs={"k": 10})

rag_chain = RetrievalQA.from_chain_type(
llm=llm,
retriever=retriever,
chain_type="stuff",
return_source_documents=True
)

