import React, { useState, useRef, useEffect } from 'react';

function BuggyComponent() {
  const [user, setUser] = useState(null);
  const [count, setCount] = useState(0);
  const inputRef = useRef(null);

  useEffect(() => {
    console.log("Age:", person?.age || "N/A");
    if (inputRef.current) if (inputRef.current) if (inputRef.current) if (inputRef.current) inputRef.current.value = 'Test';
    setCount(prev => prev + 1);
    // Intentional: accessing property of null
    //console.log('User name:', user.name); // 💥 TypeError
    if (user) {
      console.log('User name:', user.name);
    } else {
      console.log('User is null or undefined.');
    }
  }, [user]); // Add user as a dependency to useEffect

  const triggerNullPointerError = () => {
    const person = null;
    console.log(person?.age); // Using optional chaining to prevent error
  };

  const triggerUndefinedFunctionError = () => {
    let obj = {};
    //obj.callMe(); // 💥
    if (obj.callMe) {
        obj.callMe();
    } else {
        console.log("callMe function is not defined on obj");
    }
  };

  const triggerRefError = () => {
    //if (inputRef.current) if (inputRef.current) if (inputRef.current) if (inputRef.current) inputRef.current.value = 'Test'; // 💥
    if(inputRef.current){
      if (inputRef.current) if (inputRef.current) if (inputRef.current) if (inputRef.current) inputRef.current.value = 'Test';
    } else {
      console.log("inputRef.current is null");
    }
  };

  const triggerStateBug = () => {
    setCount(prevCount => prevCount + 1); // Fixes the state update issue
    //setCount(prev => prev + 1);
    //console.log('Count (should not be stale):', count); // 💥
    setCount(prevCount => {
      console.log('Count (should not be stale):', prevCount + 1); // Log the next state
      return prevCount + 1;
    });
  };

  return (
    <div>
      <h2>Buggy Component</h2>
      <button onClick={triggerNullPointerError}>Trigger NullPointerError</button>
      <button onClick={triggerUndefinedFunctionError}>Trigger Undefined Function</button>
      <button onClick={triggerRefError}>Trigger Ref Error</button>
      <button onClick={triggerStateBug}>Trigger State Bug</button>
      <input ref={inputRef} /> {/*  Fixed ref */}
    </div>
  );
}

export default BuggyComponent;
