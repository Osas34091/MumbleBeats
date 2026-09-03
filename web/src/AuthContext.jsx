import React, { createContext, useState, useContext, useEffect } from 'react';

const AuthContext = createContext({});

export function AuthProvider({ children }) {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [username, setUsername] = useState(null);
  const [isSetupComplete, setIsSetupComplete] = useState(true);
  const [isLoading, setIsLoading] = useState(true);

  const checkAuth = async () => {
    try {
      // First check if setup is needed
      const setupRes = await fetch('/api/setup/status');
      if (setupRes.ok) {
        const setupData = await setupRes.json();
        setIsSetupComplete(setupData.is_setup);
        
        if (!setupData.is_setup) {
          setIsLoading(false);
          return; // Stop here if setup isn't complete
        }
      }

      // If setup is complete, check if logged in
      const meRes = await fetch('/api/me');
      if (meRes.ok) {
        const meData = await meRes.json();
        setIsAuthenticated(true);
        setUsername(meData.username);
      } else {
        setIsAuthenticated(false);
        setUsername(null);
      }
    } catch (err) {
      console.error("Auth check error", err);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    checkAuth();
  }, []);

  const login = async (username, password) => {
    const res = await fetch('/api/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password })
    });
    
    if (res.ok) {
      await checkAuth();
      return true;
    }
    return false;
  };

  const logout = async () => {
    await fetch('/api/logout', { method: 'POST' });
    setIsAuthenticated(false);
    setUsername(null);
  };

  return (
    <AuthContext.Provider value={{ isAuthenticated, username, isSetupComplete, isLoading, login, logout, checkAuth }}>
      {children}
    </AuthContext.Provider>
  );
}

export const useAuth = () => useContext(AuthContext);
