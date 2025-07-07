import os 
from dotenv import load_dotenv

load_dotenv()

CLONE_REPO=True
REPO_URL= "https://github.com/Manav7603/buggy_app"
PROJECT_ID = "logger-462111"
REGION = "us-central1"
BATCH_SIZE = 5
CODEBASE_PATH = os.getenv("CODE_BASE_PATH","./codebase")
EMBEDDING_MODEL = "models/embedding-001"
GEMINI_LLM="gemini-2.0-flash"
CLAUDE_LLM="claude-3-sonnet"
OPEN_AI_LLM= "gpt-4"
PERSISTENT_VECTORSTORE_PATH = "./internal/data"
VECTOR_SEARCH_THRESHOLD = 0.6 
TOP_K = 10 
