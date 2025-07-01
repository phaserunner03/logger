import json
import re
from typing import Dict

def parse_response(text: str) -> Dict[str, str]:
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

if __name__ == "__main__":
    response_text = """
```json\n{\n  \"changes\": \"import React, { useState, useRef, useEffect } from 'react';\\n\\n\\nfunction customLogger(message, data = null, severity = 'error') {\\n  const logData = { message, data, severity, timestamp: new Date().toISOString() };\\n  switch (severity) {\\n    case 'error':\\n      console.error(logData);\\n      break;\\n    case 'warn':\\n      console.warn(logData);\\n      break;\\n    case 'info':\\n      console.info(logData);\\n      break;\\n    default:\\n      console.log(logData);\\n  }\\n}\\n\\nfunction BuggyComponent() {\\n  const [user, setUser] = useState(null);\\n  const [count, setCount] = useState(0);\\n  const inputRef = useRef(null);\\n\\n  useEffect(() => {\\n    // This simulates a ReferenceError (person is not defined)\\n    try {\\n      const person = { age: 30 }; // Define person object\\n      customLogger(\\\"Age:\\\", { age: person.age }, 'error');\\n    } catch (err) {\\n      customLogger(\\\"ReferenceError caught in useEffect\\\", { error: err.message, stack: err.stack }, 'error');\\n    }\\n    if (inputRef.current) if (inputRef.current) if (inputRef.current) if (inputRef.current) inputRef.current.value = 'Test';\\n    setCount(prev => prev + 1);\\n    // Intentional: accessing property of null\\n    //customLogger('User name:', { name: user.name }, 'error'); // 💥 TypeError\\n    if (user) {\\n      customLogger('User name:', { name: user.name }, 'error');\\n    } else {\\n      customLogger('User is null or undefined.', null, 'error');\\n    }\\n  }, [user]);\\n\\n  const triggerNullPointerError = () => {\\n    try {\\n      const person = null;\\n      // This will throw a TypeError\\n      customLogger('Trigger NullPointerError', { age: person.age }, 'error');\\n    } catch (err) {\\n      customLogger(\\\"NullPointerError caught\\\", { error: err.message, stack: err.stack }, 'error');\\n    }\\n  };\\n\\n  const triggerUndefinedFunctionError = () => {\\n    let obj = {};\\n    try {\\n      obj.callMe(); // This will throw a TypeError\\n      customLogger(\\\"callMe function called on obj\\\", null, 'error');\\n    } catch (err) {\\n      customLogger(\\\"Undefined function error caught\\\", { error: err.message, stack: err.stack }, 'error');\\n    }\\n  };\\n\\n  const triggerRefError = () => {\\n    if (inputRef.current) {\\n      if (inputRef.current) if (inputRef.current) if (inputRef.current) if (inputRef.current) inputRef.current.value = 'Test';\\n      customLogger(\\\"inputRef.current is set\\\", null, 'error');\\n    } else {\\n      customLogger(\\\"inputRef.current is null\\\", null, 'error');\\n    }\\n  };\\n\\n  const triggerStateBug = () => {\\n    setCount(prevCount => prevCount + 1);\\n    setCount(prevCount => {\\n      customLogger('Count (should not be stale):', { nextCount: prevCount + 1 }, 'error');\\n      return prevCount + 1;\\n    });\\n  };\\n\\n  return (\\n    <div>\\n      <h2>Buggy Component</h2>\\n      <button onClick={triggerNullPointerError}>Trigger NullPointerError</button>\\n      <button onClick={triggerUndefinedFunctionError}>Trigger Undefined Function</button>\\n      <button onClick={triggerRefError}>Trigger Ref Error</button>\\n      <button onClick={triggerStateBug}>Trigger State Bug</button>\\n      <input ref={inputRef} />\\n    </div>\\n  );\\n}\\n\\nexport default BuggyComponent;\\n// Note: This component is intentionally buggy for demonstration purposes.\\n\",\n  \"filename\": \"src/components/BuggyComponent.jsx\",\n  \"explanation\": \"Defined the 'person' object within the useEffect hook to resolve the ReferenceError."
"""
    parsed = parse_response(response_text)
    print(parsed['explanation'])