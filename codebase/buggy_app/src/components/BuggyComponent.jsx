import React, { useState, useRef, useEffect } from 'react';

function BuggyComponent() {
  const [user, setUser] = useState(null);
  const [count, setCount] = useState(0);
  const inputRef = useRef(null);

  useEffect(() => {
    // Intentional: accessing property of null
    console.log('User name:', user.name); // 💥 TypeError
  }, []);

  const triggerNullPointerError = () => {
    const person = null;
    console.log(person.age); // 💥
  };

  const triggerUndefinedFunctionError = () => {
    let obj = {};
    obj.callMe(); // 💥
  };

  const triggerRefError = () => {
    inputRef.current.value = 'Test'; // 💥
  };

  const triggerStateBug = () => {
    setCount(count++);
    console.log('Count (should not be stale):', count); // 💥
  };

  return (
    <div>
      <h2>Buggy Component</h2>
      <button onClick={triggerNullPointerError}>Trigger NullPointerError</button>
      <button onClick={triggerUndefinedFunctionError}>Trigger Undefined Function</button>
      <button onClick={triggerRefError}>Trigger Ref Error</button>
      <button onClick={triggerStateBug}>Trigger State Bug</button>
      <input ref={null} /> {/* 💥 Wrong ref */}
    </div>
  );
}

export default BuggyComponent;