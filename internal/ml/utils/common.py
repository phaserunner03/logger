import re

def sanitize_filename(s):
    return re.sub(r'[^\w\-_.]', '_', s)
