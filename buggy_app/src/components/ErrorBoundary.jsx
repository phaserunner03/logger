import React from 'react';

class ErrorBoundary extends React.Component {
  constructor(props) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error) {
    return { hasError: true, error };
  }

  componentDidCatch(error, errorInfo) {
    console.error("Error caught by ErrorBoundary:", error);

    // Send error to backend
    console.log(error);

    fetch('http://localhost:4000/log-error', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        message: error.toString(),
        stack: errorInfo.componentStack,
      }),
    }).catch(err => console.error('Failed to log error:', err));
  }

  render() {
    if (this.state.hasError) {
      return <h3>Something went wrong. Check logs.</h3>;
    }
    return this.props.children;
  }
}

export default ErrorBoundary;
