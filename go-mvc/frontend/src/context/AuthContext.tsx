"use client";

import React, { createContext, useContext, useEffect, useState } from "react";
import { User, LoginPayload, RegisterPayload } from "@/types/auth";
import {
  fetchCurrentUser,
  getStoredToken,
  setStoredToken,
  clearStoredToken,
  loginApi,
  registerApi,
} from "@/lib/api";

interface AuthContextType {
  user: User | null;
  token: string | null;
  isLoading: boolean;
  login: (payload: LoginPayload) => Promise<void>;
  register: (payload: RegisterPayload) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const existingToken = getStoredToken();
    if (existingToken) {
      setToken(existingToken);
      fetchCurrentUser()
        .then((userData) => setUser(userData))
        .catch(() => {
          clearStoredToken();
          setToken(null);
          setUser(null);
        })
        .finally(() => setIsLoading(false));
    } else {
      setIsLoading(false);
    }
  }, []);

  const login = async (payload: LoginPayload) => {
    const authData = await loginApi(payload);
    setStoredToken(authData.token);
    setToken(authData.token);
    setUser(authData.user);
  };

  const register = async (payload: RegisterPayload) => {
    const authData = await registerApi(payload);
    setStoredToken(authData.token);
    setToken(authData.token);
    setUser(authData.user);
  };

  const logout = () => {
    clearStoredToken();
    setToken(null);
    setUser(null);
  };

  return (
    <AuthContext.Provider value={{ user, token, isLoading, login, register, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextType {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
