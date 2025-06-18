import os
import subprocess
from config import CLONE_REPO,CODEBASE_PATH,REPO_URL

def prepare_codebase():
    if CLONE_REPO:
        if not os.path.exists(CODEBASE_PATH):
            print(f"[INFO] Cloning {REPO_URL} into {CODEBASE_PATH}...")
            subprocess.run(["git", "clone", REPO_URL, CODEBASE_PATH], check=True)
        else:
            print(f"[INFO] Repository already cloned at {CODEBASE_PATH}")
    else:
        if not os.path.exists(CODEBASE_PATH):
            raise FileNotFoundError(f"[ERROR] CODEBASE_PATH {CODEBASE_PATH} does not exist and cloning is disabled.")
        print(f"Use code base found at {CODEBASE_PATH}")
            