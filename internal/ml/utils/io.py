import os

def save_code(code, filename, output_dir="suggested_fixes"):
    os.makedirs(output_dir, exist_ok=True)
    path = os.path.join(output_dir, filename)
    with open(path, "w") as f:
        f.write(code)
    return path
