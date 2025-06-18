import os 
from dotenv import load_dotenv

load_dotenv()

CLONE_REPO=False
REPO_URL= "https://github.com/phaserunner03"
# CODEBASE_PATH = "../../buggy_app/backend"
PROJECT_ID = "meet-and-media-sync"
REGION = "us-central1"
BIGQUERY_DATASET = "logging"
BIGQUERY_TABLE_ID = "log_table"
BATCH_SIZE = 5
CODEBASE_PATH = "../../buggy_app/src"
LOG_FILE_PATH = "../../buggy_app/backend/error_logs.json"

EMBEDDING_MODEL = "models/embedding-001"
LLM_MODEL = "gemini-2.0-flash"