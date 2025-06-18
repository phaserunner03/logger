import re

def extract_code_block(text):
    matches = re.findall(r"```(?:\w+)?\s*(.*?)```", text, re.DOTALL)
    return matches[0].strip() if matches else None
