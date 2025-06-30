import os

def save_code(code: str, file_path: str):

# Normalize and resolve the relative path
    resolved_path = os.path.normpath(file_path)
    # Ensure directory exists
    os.makedirs(os.path.dirname(resolved_path), exist_ok=True)

    # Write (overwrite) the file
    with open(resolved_path, "w", encoding="utf-8") as f:
        f.write(code)

    print(f"✅ Code saved to: {resolved_path}")
    return resolved_path
