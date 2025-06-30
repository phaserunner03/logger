import re

def extract_code_block(text):
    # Extract code block
    code_pattern = re.compile(r"```(?:[a-zA-Z]+)?\s*(.*?)\s*```", re.DOTALL)
    code_match = code_pattern.search(text)

    # Extract filename
    filename_pattern = re.compile(r"filename:\s*(.*)")
    filename_match = filename_pattern.search(text)

    if not code_match:
        print("⚠️ No code block found in the suggestion.")
        return None, None

    code = code_match.group(1).strip()
    filename = filename_match.group(1).strip() if filename_match else None

    return code, filename
