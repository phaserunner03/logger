useEffect(() => {
    // Intentional: accessing property of null
    //console.log('User name:', user.name); // 💥 TypeError
    if (user) {
      console.log('User name:', user.name);
    } else {
      console.log('User is null or undefined.');
    }
  }, [user]); // Add user as a dependency to useEffect