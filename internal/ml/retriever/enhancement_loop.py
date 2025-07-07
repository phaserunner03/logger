import os
from typing import List, Dict, Any

from langchain.docstore.document import Document
from langchain.embeddings.base import Embeddings
from langchain_google_genai import ChatGoogleGenerativeAI
from langchain_chroma import Chroma

from config import (
    PERSISTENT_VECTORSTORE_PATH,
    VECTOR_SEARCH_THRESHOLD,
    TOP_K,
)


def build_prompt(query: str, top_docs: List[Document]) -> str:
    header = (
        "### You are a highly proficient code assistant with full access to the project's source code.\n"
        "A developer has encountered the following issue and is requesting your help:\n\n"
        f">>> **User Query:**\n> \"\"\"\n{query}\n\"\"\"\n\n"
        "Based on the codebase and the context above, suggest a detailed and accurate code level fix.\n"
        "Respond with the corrected code and complete file with the fix applied.\n"
        "Make sure the code is valid, handle edge cases, and is well-structured.\n"
        "Make sure to give the entire code only."
    )

    ctx = ["### Relevant files from the repository:\n"]
    for idx, doc in enumerate(top_docs, start=1):
        m = doc.metadata
        snippet = m.get("full_code", "")
        if len(snippet) > 5000:
            snippet = snippet[:5000] + "\n// ... (truncated)\n"

        ctx.extend([
            f"**File #{idx}:** `{m['filename']}`",
            f"- Language: `{m['language']}`",
            f"- Last Modified: `{m['last_modified']}`",
            f"```{m['language']}",
            snippet,
            "```",
            ""
        ])

    footer = (
        "### Instructions:\n"
        "1. Analyze the user's query in the context of the above files.\n"
        "2. Return the complete updated file.\n"
        "3. Handle edge cases and preserve coding style.\n"
        "4. In all cases, produce your answer as a single JSON object **only**, with exactly three keys:\n"
        "   - `changes`: (string) the full updated file text,\n"
        "   - `filename`: (string) the relative path of the file to update,\n"
        "   - `explanation`: (string) a very brief summary of what changed and why.\n\n"
        "### Output JSON Schema:\n"
        "```json\n"
        "{\n"
        "  \"changes\": \"<full file content>\",\n"
        "  \"filename\": \"<relative/path/to/file>\",\n"
        "  \"explanation\": \"<short explanation>\"\n"
        "}\n"
        "```"
    )

    return "\n".join([header] + ctx + [footer])


class EnhancementLoop:
    def __init__(
        self,
        embedding_model: Embeddings,
        llm: ChatGoogleGenerativeAI,
        persist_dir: str = PERSISTENT_VECTORSTORE_PATH,
    ):
        os.makedirs(persist_dir, exist_ok=True)
        if os.path.isdir(persist_dir) and os.listdir(persist_dir):
            print("Vectorstore loaded and retriever created.")
            self.vs = Chroma(
                persist_directory=persist_dir,
                embedding_function=embedding_model
            )
        else:
            print("Creating new vectorstore from documents…")
            from codebase.processor import load_codebase_as_docs
            code_docs = load_codebase_as_docs()
            self.vs = Chroma.from_documents(
                documents=code_docs,
                embedding=embedding_model,
                persist_directory=persist_dir
            )
        self.embedding_model = embedding_model
        self.llm = llm

    def run(self, query: str) -> Dict[str, Any]:
        # 1) embed the query
        q_vec = self.embedding_model.embed_query(query)

        # 2) search using retrieval API that returns scores
        docs_and_scores = self.vs.similarity_search_with_score(
            query=query,
            k=TOP_K
        )

        # 3) filter by threshold
        filtered = [(doc, score) for doc, score in docs_and_scores if score >= VECTOR_SEARCH_THRESHOLD]
        if not filtered:
            return {"error": "No documents passed the similarity threshold", "query": query, "results": []}

        # 4) build and send prompt
        prompt = build_prompt(query, [d for d, _ in filtered])
        resp = self.llm.invoke(prompt)
        text = resp.content if hasattr(resp, 'content') else getattr(resp, 'result', str(resp))
        return text

       
       