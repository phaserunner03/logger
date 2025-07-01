# retriever/enhancement_loop.py

import os
from typing import List, Dict, Any

from langchain.docstore.document import Document
# from langchain.vectorstores import Chroma
from langchain.embeddings.base import Embeddings  # abstract
from langchain_google_genai import ChatGoogleGenerativeAI
# new
from langchain_chroma import Chroma


# config values you already have
from config import (
    PERSISTENT_VECTORSTORE_PATH,
    EMBEDDING_MODEL,
    LLM_MODEL,
    VECTOR_SEARCH_THRESHOLD,  # e.g. 0.7
    TOP_K                  # e.g. 10
)


def build_prompt(query: str, top_docs: List[Document]) -> str:
    """
    Builds a detailed prompt for Gemini that:
      1. Presents the user’s query.
      2. Provides full metadata (filename, language, last modified) and source text for each top hit.
      3. Clearly instructs Gemini to produce a precise code‑level fix or enhancement.
      4. Specifies an unambiguous output format.

    Returns:
        A single prompt string.
    """
    # 1) Query introduction
    header = (
        "### You are a highly proficient code assistant with full access to the project's source code.\n"
        "A developer has encountered the following issue and is requesting your help:\n\n"
        f">>> **User Query:**\n> \"\"\"\n{query}\n\"\"\"\n\n"
    )

    # 2) Context files
    ctx = ["### Relevant files from the repository:\n"]
    for idx, doc in enumerate(top_docs, start=1):
        m = doc.metadata
        ctx.append(f"**File #{idx}:** `{m['filename']}`")
        ctx.append(f"- Language: `{m['language']}`")
        ctx.append(f"- Last Modified: `{m['last_modified']}`")
        ctx.append("```" + m['language'])
        # truncate extremely long files for performance:
        snippet = m['full_code']
        if len(snippet) > 5000:
            snippet = snippet[:5000] + "\n// ... (truncated)\n"
        ctx.append(snippet)
        ctx.append("```")
        ctx.append("")  # blank line

    # 3) Clear instructions
    footer = (
        "### Instructions:\n"
        "1. Analyze the user’s query in the context of the above files.\n"
        "2. If only a small change is needed, provide a **unified diff** showing the exact lines to add, modify, or remove.\n"
        "3. If a full-file replacement is required, return the complete updated file.\n"
        "4. Handle edge cases and maintain existing code style and patterns.\n"
        "5. Include a brief explanation (1–2 sentences) of what was changed and why.\n\n"
        "### Output Format (must follow exactly):\n"
        "```diff\n"
        "<unified-diff-or-full-file-content>\n"
        "```\n"
        "**Filename:** `<relative/path/to/file>`\n"
        "**Explanation:** <short explanation here>\n"
    )

    # 4) Combine all parts
    prompt = "\n".join([header] + ctx + [footer])
    return prompt





class EnhancementLoop:
    def __init__(
        self,
        embedding_model: Embeddings,
        llm: ChatGoogleGenerativeAI,
        persist_dir: str = PERSISTENT_VECTORSTORE_PATH,
    ):
        # 1) Ensure the vector‐store directory exists
        os.makedirs(persist_dir, exist_ok=True)

        # 2) Load or initialize Chroma
        if os.path.isdir(persist_dir) and os.listdir(persist_dir):
            print("Vectorstore loaded and retriever created.")
            self.vs = Chroma(
                persist_directory=persist_dir,
                embedding_function=embedding_model
            )
        else:
            print("Creating new vectorstore from documents…")
            # you need to actually load your docs here:
            from codebase.processor import load_codebase_as_docs
            code_docs = load_codebase_as_docs()

            self.vs = Chroma.from_documents(
                documents=code_docs,
                embedding=embedding_model,
                persist_directory=persist_dir
            )
            # if your version of Chroma requires an explicit persist call:
            # self.vs.persist()

        # 3) keep references for later
        self.embedding_model = embedding_model
        self.llm = llm

    def run(self, query: str) -> Dict[str, Any]:
        # 1) embed the query via the public API, not the private _embedding_function
        q_vec = self.embedding_model.embed_query(query)

        # 2) perform vector search with scores
        docs_and_scores = self.vs.similarity_search_by_vector(
            q_vec, k=TOP_K, include_scores=True
        )

        # 3) filter by threshold
        filtered = [(doc, score) for doc, score in docs_and_scores
                    if score >= VECTOR_SEARCH_THRESHOLD]
        if not filtered:
            return {
                "error": "No documents passed the similarity threshold",
                "query": query,
                "results": []
            }

        # 4) build prompt
        prompt = build_prompt(query, [doc for doc, _ in filtered])

        # 5) invoke Gemini
        resp = self.llm.invoke({"query": prompt})

        return {
            "query":        query,
            "matched_docs": [
                {"filename": d.metadata["filename"], "score": s}
                for d, s in filtered
            ],
            "suggestion":   resp["result"]
        }