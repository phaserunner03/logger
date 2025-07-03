import json
import re
from typing import Dict, Union

def parse_response(text: Union[str, dict]) -> Dict[str, str]:
    if isinstance(text, dict):
        # Already parsed JSON
        return {
            "filename": text.get("filename", ""),
            "changes": text.get("changes", ""),
            "explanation": text.get("explanation", "")
        }

    if not isinstance(text, str):
        raise TypeError("Expected string or dict as input to parse_response")

    # 1) Try plain JSON string
    try:
        parsed = json.loads(text)
        if isinstance(parsed, dict):
            return {
                "filename": parsed.get("filename", ""),
                "changes": parsed.get("changes", ""),
                "explanation": parsed.get("explanation", "")
            }
    except json.JSONDecodeError:
        pass

    # 2) Fallback to markdown-style block
    json_block = re.search(r"```json\s*({[\s\S]*?})\s*```", text)
    if json_block:
        try:
            parsed = json.loads(json_block.group(1))
            return {
                "filename": parsed.get("filename", ""),
                "changes": parsed.get("changes", ""),
                "explanation": parsed.get("explanation", "")
            }
        except json.JSONDecodeError:
            pass

    # 3) Final fallback to regex field extract
    def extract_field(field_name):
        pattern = rf'"{field_name}"\s*:\s*"((?:\\.|[^"\\])*)"'
        match = re.search(pattern, text)
        if match:
            return bytes(match.group(1), "utf-8").decode("unicode_escape")
        return ""

    return {
        "filename": extract_field("filename"),
        "changes": extract_field("changes"),
        "explanation": extract_field("explanation")
    }