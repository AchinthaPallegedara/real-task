# 🚀 Frontend Integration Guide - Next.js 15 with Task Management API

A comprehensive guide for building a modern, real-time task management frontend using **Next.js 15** with Server-Side Rendering (SSR), integrated with the Go-based Task Management API.

## 📋 Table of Contents

1. [Project Overview](#project-overview)
2. [Prerequisites](#prerequisites)
3. [Project Setup](#project-setup)
4. [Architecture](#architecture)
5. [Authentication Implementation](#authentication-implementation)
6. [API Integration](#api-integration)
7. [Real-time Features](#real-time-features)
8. [Pages & Components](#pages--components)
9. [State Management](#state-management)
10. [Styling & UI](#styling--ui)
11. [Deployment](#deployment)
12. [Best Practices](#best-practices)

---

## 🎯 Project Overview

This frontend application will provide:

- **🔐 Complete Authentication System** - Login, registration, email verification, 2FA
- **📋 Project & Task Management** - Full CRUD operations with real-time updates
- **👥 Collaboration Features** - User invitations, real-time presence, typing indicators
- **⚡ Real-time Updates** - WebSocket integration for live collaboration
- **📱 Responsive Design** - Mobile-first design with modern UI
- **🚀 SSR Performance** - Next.js 15 App Router with optimized loading

---

## 🛠️ Prerequisites

- **Node.js 18.18+** (required for Next.js 15)
- **npm** or **yarn** package manager
- **Go API Server** running on `http://localhost:8080`
- Basic knowledge of React, TypeScript, and Next.js

---

## 🏗️ Project Setup

### 1. Create Next.js 15 Project

```bash
# Create new Next.js 15 app with TypeScript
npx create-next-app@latest task-manager-frontend --typescript --tailwind --eslint --app --src-dir --import-alias "@/*"

cd task-manager-frontend
```

### 2. Install Required Dependencies

```bash
# Core dependencies
npm install @tanstack/react-query axios socket.io-client
npm install react-hook-form @hookform/resolvers zod
npm install @headlessui/react @heroicons/react
npm install react-hot-toast sonner
npm install qrcode react-qr-code
npm install js-cookie @types/js-cookie

# Development dependencies
npm install -D @types/qrcode
```

### 3. Environment Configuration

Create `.env.local`:

```env
# API Configuration
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
NEXT_PUBLIC_WS_URL=ws://localhost:8080

# App Configuration
NEXT_PUBLIC_APP_NAME="Task Manager"
NEXT_PUBLIC_APP_URL=http://localhost:3000

# Development
NODE_ENV=development
```

### 4. TypeScript Configuration

Update `tsconfig.json`:

```json
{
  "compilerOptions": {
    "target": "ES2017",
    "lib": ["dom", "dom.iterable", "ES6"],
    "allowJs": true,
    "skipLibCheck": true,
    "strict": true,
    "noEmit": true,
    "esModuleInterop": true,
    "module": "esnext",
    "moduleResolution": "bundler",
    "resolveJsonModule": true,
    "isolatedModules": true,
    "jsx": "preserve",
    "incremental": true,
    "plugins": [
      {
        "name": "next"
      }
    ],
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"],
      "@/components/*": ["./src/components/*"],
      "@/lib/*": ["./src/lib/*"],
      "@/types/*": ["./src/types/*"],
      "@/hooks/*": ["./src/hooks/*"],
      "@/utils/*": ["./src/utils/*"]
    }
  },
  "include": ["next-env.d.ts", "**/*.ts", "**/*.tsx", ".next/types/**/*.ts"],
  "exclude": ["node_modules"]
}
```

---

## 🏛️ Architecture

### Project Structure

```
src/
├── app/                    # Next.js 15 App Router
│   ├── (auth)/            # Auth route group
│   │   ├── login/
│   │   ├── register/
│   │   ├── verify-email/
│   │   └── reset-password/
│   ├── (dashboard)/       # Protected route group
│   │   ├── dashboard/
│   │   ├── projects/
│   │   ├── tasks/
│   │   └── settings/
│   ├── globals.css
│   ├── layout.tsx
│   ├── loading.tsx
│   ├── not-found.tsx
│   └── page.tsx
├── components/            # Reusable components
│   ├── auth/             # Authentication components
│   ├── dashboard/        # Dashboard components
│   ├── projects/         # Project components
│   ├── tasks/           # Task components
│   ├── ui/              # Base UI components
│   └── layout/          # Layout components
├── lib/                  # Core library functions
│   ├── api.ts           # API client
│   ├── auth.ts          # Authentication logic
│   ├── websocket.ts     # WebSocket client
│   ├── validators.ts    # Zod validation schemas
│   └── utils.ts         # Utility functions
├── hooks/               # Custom React hooks
│   ├── useAuth.ts
│   ├── useWebSocket.ts
│   ├── useProjects.ts
│   └── useTasks.ts
├── types/               # TypeScript definitions
│   ├── auth.ts
│   ├── api.ts
│   ├── project.ts
│   └── task.ts
├── providers/           # Context providers
│   ├── AuthProvider.tsx
│   ├── QueryProvider.tsx
│   └── WebSocketProvider.tsx
└── middleware.ts        # Next.js middleware for auth
```

---

## 🔐 Authentication Implementation

### 1. Types Definition

Create `src/types/auth.ts`:

```typescript
export interface User {
  id: number;
  email: string;
  name?: string;
  email_verified: boolean;
  two_factor_enabled: boolean;
  is_active: boolean;
  last_login_at?: string;
}

export interface AuthState {
  user: User | null;
  token: string | null;
  isLoading: boolean;
  isAuthenticated: boolean;
}

export interface LoginRequest {
  email: string;
  password: string;
  two_factor_token?: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  message: string;
  token: string;
  user: User;
}

export interface TwoFactorSetup {
  secret: string;
  qr_code: string;
  backup_codes?: string[];
}
```

### 2. API Client

Create `src/lib/api.ts`:

```typescript
import axios, { AxiosInstance, AxiosResponse } from "axios";
import { toast } from "sonner";

const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL || "http://localhost:8080";

// Create axios instance
const api: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    "Content-Type": "application/json",
  },
});

// Token management
let authToken: string | null = null;

export const setAuthToken = (token: string | null) => {
  authToken = token;
  if (token) {
    api.defaults.headers.common["Authorization"] = `Bearer ${token}`;
    if (typeof window !== "undefined") {
      localStorage.setItem("auth_token", token);
    }
  } else {
    delete api.defaults.headers.common["Authorization"];
    if (typeof window !== "undefined") {
      localStorage.removeItem("auth_token");
    }
  }
};

// Initialize token from localStorage on client side
if (typeof window !== "undefined") {
  const savedToken = localStorage.getItem("auth_token");
  if (savedToken) {
    setAuthToken(savedToken);
  }
}

// Response interceptor for error handling
api.interceptors.response.use(
  (response: AxiosResponse) => response,
  (error) => {
    if (error.response?.status === 401) {
      setAuthToken(null);
      if (typeof window !== "undefined") {
        window.location.href = "/login";
      }
    }

    // Show error toast
    const message = error.response?.data?.error || "An error occurred";
    toast.error(message);

    return Promise.reject(error);
  }
);

export default api;

// Auth API functions
export const authAPI = {
  // Authentication
  register: (data: RegisterRequest) => api.post("/register", data),

  login: (data: LoginRequest) => api.post("/login", data),

  verifyEmail: (token: string) => api.post("/api/auth/verify-email", { token }),

  resendVerification: (email: string) =>
    api.post("/api/auth/resend-verification", { email }),

  requestPasswordReset: (email: string) =>
    api.post("/api/auth/request-password-reset", { email }),

  resetPassword: (token: string, new_password: string) =>
    api.post("/api/auth/reset-password", { token, new_password }),

  getAuthStatus: () => api.get("/api/auth/status"),

  changePassword: (current_password: string, new_password: string) =>
    api.post("/api/auth/change-password", { current_password, new_password }),

  // Two-Factor Authentication
  setup2FA: () => api.post("/api/auth/2fa/setup"),

  confirm2FA: (token: string) => api.post("/api/auth/2fa/confirm", { token }),

  disable2FA: (password: string) =>
    api.post("/api/auth/2fa/disable", { password }),

  verify2FA: (token: string) => api.post("/api/auth/2fa/verify", { token }),
};

// Projects API
export const projectsAPI = {
  getAll: () => api.get("/api/projects"),
  getById: (id: number) => api.get(`/api/projects/${id}`),
  create: (data: any) => api.post("/api/projects", data),
  update: (id: number, data: any) => api.put(`/api/projects/${id}`, data),
  delete: (id: number) => api.delete(`/api/projects/${id}`),

  // Collaboration
  invite: (projectId: number, data: any) =>
    api.post(`/api/projects/${projectId}/invite`, data),
  getCollaborators: (projectId: number) =>
    api.get(`/api/projects/${projectId}/collaborators`),
  removeCollaborator: (projectId: number, userId: number) =>
    api.delete(`/api/projects/${projectId}/collaborators/${userId}`),
};

// Tasks API
export const tasksAPI = {
  getForProject: (projectId: number) =>
    api.get(`/api/projects/${projectId}/tasks`),
  getById: (id: number) => api.get(`/api/tasks/${id}`),
  create: (projectId: number, data: any) =>
    api.post(`/api/projects/${projectId}/tasks`, data),
  update: (id: number, data: any) => api.put(`/api/tasks/${id}`, data),
  delete: (id: number) => api.delete(`/api/tasks/${id}`),
};
```

### 3. Authentication Hook

Create `src/hooks/useAuth.ts`:

```typescript
"use client";

import {
  createContext,
  useContext,
  useEffect,
  useState,
  ReactNode,
} from "react";
import { User, AuthState, LoginRequest, RegisterRequest } from "@/types/auth";
import { authAPI, setAuthToken } from "@/lib/api";
import { toast } from "sonner";
import { useRouter } from "next/navigation";

interface AuthContextType extends AuthState {
  login: (data: LoginRequest) => Promise<boolean>;
  register: (data: RegisterRequest) => Promise<boolean>;
  logout: () => void;
  refreshUser: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [state, setState] = useState<AuthState>({
    user: null,
    token: null,
    isLoading: true,
    isAuthenticated: false,
  });

  const router = useRouter();

  // Initialize auth state
  useEffect(() => {
    const initAuth = async () => {
      const token = localStorage.getItem("auth_token");
      if (token) {
        setAuthToken(token);
        try {
          const response = await authAPI.getAuthStatus();
          setState({
            user: response.data.user,
            token,
            isLoading: false,
            isAuthenticated: true,
          });
        } catch (error) {
          // Token is invalid
          setAuthToken(null);
          setState({
            user: null,
            token: null,
            isLoading: false,
            isAuthenticated: false,
          });
        }
      } else {
        setState((prev) => ({ ...prev, isLoading: false }));
      }
    };

    initAuth();
  }, []);

  const login = async (data: LoginRequest): Promise<boolean> => {
    try {
      const response = await authAPI.login(data);
      const { token, user } = response.data;

      setAuthToken(token);
      setState({
        user,
        token,
        isLoading: false,
        isAuthenticated: true,
      });

      toast.success("Login successful!");
      return true;
    } catch (error: any) {
      if (error.response?.data?.two_factor_required) {
        return false; // Indicate 2FA is required
      }
      throw error;
    }
  };

  const register = async (data: RegisterRequest): Promise<boolean> => {
    try {
      await authAPI.register(data);
      toast.success(
        "Registration successful! Please check your email for verification."
      );
      return true;
    } catch (error) {
      throw error;
    }
  };

  const logout = () => {
    setAuthToken(null);
    setState({
      user: null,
      token: null,
      isLoading: false,
      isAuthenticated: false,
    });
    router.push("/login");
    toast.success("Logged out successfully");
  };

  const refreshUser = async () => {
    try {
      const response = await authAPI.getAuthStatus();
      setState((prev) => ({
        ...prev,
        user: response.data.user,
      }));
    } catch (error) {
      console.error("Failed to refresh user:", error);
    }
  };

  return (
    <AuthContext.Provider
      value={{
        ...state,
        login,
        register,
        logout,
        refreshUser,
      }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
};
```

### 4. Middleware for Route Protection

Create `src/middleware.ts`:

```typescript
import { NextRequest, NextResponse } from "next/server";

// Define protected routes
const protectedRoutes = ["/dashboard", "/projects", "/tasks", "/settings"];
const authRoutes = ["/login", "/register", "/verify-email", "/reset-password"];

export function middleware(request: NextRequest) {
  const token =
    request.cookies.get("auth_token")?.value ||
    request.headers.get("authorization")?.replace("Bearer ", "");

  const { pathname } = request.nextUrl;

  // Check if the current route is protected
  const isProtectedRoute = protectedRoutes.some((route) =>
    pathname.startsWith(route)
  );

  // Check if the current route is an auth route
  const isAuthRoute = authRoutes.some((route) => pathname.startsWith(route));

  // If user is not authenticated and trying to access protected route
  if (isProtectedRoute && !token) {
    const loginUrl = new URL("/login", request.url);
    loginUrl.searchParams.set("redirect", pathname);
    return NextResponse.redirect(loginUrl);
  }

  // If user is authenticated and trying to access auth routes
  if (isAuthRoute && token) {
    return NextResponse.redirect(new URL("/dashboard", request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/((?!api|_next/static|_next/image|favicon.ico).*)"],
};
```

---

## 🌐 API Integration

### 1. React Query Setup

Create `src/providers/QueryProvider.tsx`:

```typescript
"use client";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ReactNode, useState } from "react";

export const QueryProvider = ({ children }: { children: ReactNode }) => {
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            staleTime: 60 * 1000, // 1 minute
            retry: 1,
          },
        },
      })
  );

  return (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
};
```

### 2. Projects Hook

Create `src/hooks/useProjects.ts`:

```typescript
"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { projectsAPI } from "@/lib/api";
import { toast } from "sonner";

export const useProjects = () => {
  return useQuery({
    queryKey: ["projects"],
    queryFn: () => projectsAPI.getAll().then((res) => res.data),
  });
};

export const useProject = (id: number) => {
  return useQuery({
    queryKey: ["projects", id],
    queryFn: () => projectsAPI.getById(id).then((res) => res.data),
    enabled: !!id,
  });
};

export const useCreateProject = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: projectsAPI.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["projects"] });
      toast.success("Project created successfully!");
    },
    onError: () => {
      toast.error("Failed to create project");
    },
  });
};

export const useUpdateProject = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: any }) =>
      projectsAPI.update(id, data),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: ["projects"] });
      queryClient.invalidateQueries({ queryKey: ["projects", id] });
      toast.success("Project updated successfully!");
    },
    onError: () => {
      toast.error("Failed to update project");
    },
  });
};

export const useDeleteProject = () => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: projectsAPI.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["projects"] });
      toast.success("Project deleted successfully!");
    },
    onError: () => {
      toast.error("Failed to delete project");
    },
  });
};
```

### 3. Tasks Hook

Create `src/hooks/useTasks.ts`:

```typescript
"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { tasksAPI } from "@/lib/api";
import { toast } from "sonner";

export const useTasks = (projectId: number) => {
  return useQuery({
    queryKey: ["tasks", projectId],
    queryFn: () => tasksAPI.getForProject(projectId).then((res) => res.data),
    enabled: !!projectId,
  });
};

export const useTask = (id: number) => {
  return useQuery({
    queryKey: ["tasks", "detail", id],
    queryFn: () => tasksAPI.getById(id).then((res) => res.data),
    enabled: !!id,
  });
};

export const useCreateTask = (projectId: number) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: any) => tasksAPI.create(projectId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["tasks", projectId] });
      toast.success("Task created successfully!");
    },
    onError: () => {
      toast.error("Failed to create task");
    },
  });
};

export const useUpdateTask = (projectId: number) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: any }) =>
      tasksAPI.update(id, data),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: ["tasks", projectId] });
      queryClient.invalidateQueries({ queryKey: ["tasks", "detail", id] });
      toast.success("Task updated successfully!");
    },
    onError: () => {
      toast.error("Failed to update task");
    },
  });
};

export const useDeleteTask = (projectId: number) => {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: tasksAPI.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["tasks", projectId] });
      toast.success("Task deleted successfully!");
    },
    onError: () => {
      toast.error("Failed to delete task");
    },
  });
};
```

---

## ⚡ Real-time Features

### 1. WebSocket Provider

Create `src/providers/WebSocketProvider.tsx`:

```typescript
"use client";

import {
  createContext,
  useContext,
  useEffect,
  useState,
  ReactNode,
} from "react";
import { useAuth } from "@/hooks/useAuth";
import { toast } from "sonner";

interface WebSocketMessage {
  type: string;
  project_id?: number;
  user_id?: number;
  data?: any;
  timestamp?: number;
}

interface WebSocketContextType {
  isConnected: boolean;
  sendMessage: (message: WebSocketMessage) => void;
  subscribe: (callback: (message: WebSocketMessage) => void) => () => void;
}

const WebSocketContext = createContext<WebSocketContextType | undefined>(
  undefined
);

export const WebSocketProvider = ({ children }: { children: ReactNode }) => {
  const [socket, setSocket] = useState<WebSocket | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [subscribers, setSubscribers] = useState<
    ((message: WebSocketMessage) => void)[]
  >([]);
  const { token, isAuthenticated } = useAuth();

  useEffect(() => {
    if (!isAuthenticated || !token) return;

    const WS_URL = process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8080";
    const wsUrl = `${WS_URL}/api/ws`;

    const ws = new WebSocket(wsUrl, [], {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });

    ws.onopen = () => {
      console.log("WebSocket connected");
      setIsConnected(true);
      setSocket(ws);
    };

    ws.onmessage = (event) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data);

        // Notify all subscribers
        subscribers.forEach((callback) => callback(message));

        // Show notifications for certain message types
        if (message.type === "task_assigned" && message.data?.assignee_id) {
          toast.info(`You've been assigned a new task: ${message.data.title}`);
        }
      } catch (error) {
        console.error("Failed to parse WebSocket message:", error);
      }
    };

    ws.onclose = () => {
      console.log("WebSocket disconnected");
      setIsConnected(false);
      setSocket(null);
    };

    ws.onerror = (error) => {
      console.error("WebSocket error:", error);
      toast.error("Real-time connection error");
    };

    return () => {
      ws.close();
    };
  }, [isAuthenticated, token]);

  const sendMessage = (message: WebSocketMessage) => {
    if (socket && isConnected) {
      socket.send(JSON.stringify(message));
    }
  };

  const subscribe = (callback: (message: WebSocketMessage) => void) => {
    setSubscribers((prev) => [...prev, callback]);

    // Return unsubscribe function
    return () => {
      setSubscribers((prev) => prev.filter((cb) => cb !== callback));
    };
  };

  return (
    <WebSocketContext.Provider value={{ isConnected, sendMessage, subscribe }}>
      {children}
    </WebSocketContext.Provider>
  );
};

export const useWebSocket = () => {
  const context = useContext(WebSocketContext);
  if (context === undefined) {
    throw new Error("useWebSocket must be used within a WebSocketProvider");
  }
  return context;
};
```

### 2. Real-time Updates Hook

Create `src/hooks/useRealTimeUpdates.ts`:

```typescript
"use client";

import { useEffect } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useWebSocket } from "@/providers/WebSocketProvider";

export const useRealTimeUpdates = (projectId?: number) => {
  const { subscribe } = useWebSocket();
  const queryClient = useQueryClient();

  useEffect(() => {
    const unsubscribe = subscribe((message) => {
      // Only handle messages for the current project or global messages
      if (projectId && message.project_id && message.project_id !== projectId) {
        return;
      }

      switch (message.type) {
        case "task_created":
        case "task_updated":
        case "task_deleted":
          // Invalidate tasks query to refresh the list
          if (message.project_id) {
            queryClient.invalidateQueries({
              queryKey: ["tasks", message.project_id],
            });
          }
          break;

        case "task_assigned":
          // Invalidate tasks query and show notification
          if (message.project_id) {
            queryClient.invalidateQueries({
              queryKey: ["tasks", message.project_id],
            });
          }
          break;

        case "user_joined_project":
        case "user_left_project":
          // Invalidate project collaborators
          if (message.project_id) {
            queryClient.invalidateQueries({
              queryKey: ["projects", message.project_id, "collaborators"],
            });
          }
          break;

        case "project_updated":
          // Invalidate specific project and projects list
          if (message.project_id) {
            queryClient.invalidateQueries({
              queryKey: ["projects", message.project_id],
            });
            queryClient.invalidateQueries({
              queryKey: ["projects"],
            });
          }
          break;
      }
    });

    return unsubscribe;
  }, [subscribe, queryClient, projectId]);
};
```

---

## 📱 Pages & Components

### 1. Root Layout

Update `src/app/layout.tsx`:

```typescript
import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import { AuthProvider } from "@/hooks/useAuth";
import { QueryProvider } from "@/providers/QueryProvider";
import { WebSocketProvider } from "@/providers/WebSocketProvider";
import { Toaster } from "sonner";

const inter = Inter({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "Task Manager - Real-time Collaboration",
  description: "Modern task management with real-time collaboration features",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className={inter.className}>
        <QueryProvider>
          <AuthProvider>
            <WebSocketProvider>
              {children}
              <Toaster position="top-right" richColors />
            </WebSocketProvider>
          </AuthProvider>
        </QueryProvider>
      </body>
    </html>
  );
}
```

### 2. Login Page

Create `src/app/(auth)/login/page.tsx`:

```typescript
"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useAuth } from "@/hooks/useAuth";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import { EyeIcon, EyeSlashIcon } from "@heroicons/react/24/outline";

const loginSchema = z.object({
  email: z.string().email("Invalid email address"),
  password: z.string().min(6, "Password must be at least 6 characters"),
  two_factor_token: z.string().optional(),
});

type LoginFormData = z.infer<typeof loginSchema>;

export default function LoginPage() {
  const [showPassword, setShowPassword] = useState(false);
  const [requires2FA, setRequires2FA] = useState(false);
  const [isLoading, setIsLoading] = useState(false);

  const { login } = useAuth();
  const router = useRouter();
  const searchParams = useSearchParams();
  const redirectTo = searchParams.get("redirect") || "/dashboard";

  const form = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      email: "",
      password: "",
      two_factor_token: "",
    },
  });

  const onSubmit = async (data: LoginFormData) => {
    setIsLoading(true);
    try {
      const success = await login(data);
      if (success) {
        router.push(redirectTo);
      } else {
        // 2FA required
        setRequires2FA(true);
      }
    } catch (error: any) {
      if (error.response?.data?.two_factor_required) {
        setRequires2FA(true);
      }
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div>
          <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
            Sign in to your account
          </h2>
          <p className="mt-2 text-center text-sm text-gray-600">
            Or{" "}
            <Link
              href="/register"
              className="font-medium text-indigo-600 hover:text-indigo-500">
              create a new account
            </Link>
          </p>
        </div>

        <form className="mt-8 space-y-6" onSubmit={form.handleSubmit(onSubmit)}>
          <div className="space-y-4">
            <div>
              <label
                htmlFor="email"
                className="block text-sm font-medium text-gray-700">
                Email address
              </label>
              <input
                {...form.register("email")}
                type="email"
                autoComplete="email"
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                placeholder="you@example.com"
              />
              {form.formState.errors.email && (
                <p className="mt-1 text-sm text-red-600">
                  {form.formState.errors.email.message}
                </p>
              )}
            </div>

            <div>
              <label
                htmlFor="password"
                className="block text-sm font-medium text-gray-700">
                Password
              </label>
              <div className="mt-1 relative">
                <input
                  {...form.register("password")}
                  type={showPassword ? "text" : "password"}
                  autoComplete="current-password"
                  className="block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 pr-10"
                  placeholder="Password"
                />
                <button
                  type="button"
                  className="absolute inset-y-0 right-0 pr-3 flex items-center"
                  onClick={() => setShowPassword(!showPassword)}>
                  {showPassword ? (
                    <EyeSlashIcon className="h-5 w-5 text-gray-400" />
                  ) : (
                    <EyeIcon className="h-5 w-5 text-gray-400" />
                  )}
                </button>
              </div>
              {form.formState.errors.password && (
                <p className="mt-1 text-sm text-red-600">
                  {form.formState.errors.password.message}
                </p>
              )}
            </div>

            {requires2FA && (
              <div>
                <label
                  htmlFor="two_factor_token"
                  className="block text-sm font-medium text-gray-700">
                  Two-Factor Authentication Code
                </label>
                <input
                  {...form.register("two_factor_token")}
                  type="text"
                  className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                  placeholder="Enter 6-digit code"
                  maxLength={6}
                />
                <p className="mt-1 text-xs text-gray-500">
                  Enter the 6-digit code from your authenticator app or use a
                  backup code.
                </p>
              </div>
            )}
          </div>

          <div className="flex items-center justify-between">
            <Link
              href="/reset-password"
              className="text-sm text-indigo-600 hover:text-indigo-500">
              Forgot your password?
            </Link>
          </div>

          <button
            type="submit"
            disabled={isLoading}
            className="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed">
            {isLoading ? "Signing in..." : "Sign in"}
          </button>
        </form>
      </div>
    </div>
  );
}
```

### 3. Dashboard Page

Create `src/app/(dashboard)/dashboard/page.tsx`:

```typescript
"use client";

import { useAuth } from "@/hooks/useAuth";
import { useProjects } from "@/hooks/useProjects";
import { useTasks } from "@/hooks/useTasks";
import { useRealTimeUpdates } from "@/hooks/useRealTimeUpdates";
import Link from "next/link";
import { PlusIcon, FolderIcon, ClockIcon } from "@heroicons/react/24/outline";

export default function DashboardPage() {
  const { user } = useAuth();
  const { data: projects, isLoading: projectsLoading } = useProjects();

  // Enable real-time updates for dashboard
  useRealTimeUpdates();

  // Get recent tasks from all projects
  const recentTasks =
    projects
      ?.flatMap(
        (project: any) =>
          project.tasks?.slice(0, 3).map((task: any) => ({
            ...task,
            project_name: project.name,
          })) || []
      )
      .slice(0, 5) || [];

  if (projectsLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-indigo-500"></div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">
          Welcome back, {user?.name || user?.email}!
        </h1>
        <p className="mt-2 text-gray-600">
          Here's what's happening with your projects today.
        </p>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <div className="bg-white p-6 rounded-lg shadow">
          <div className="flex items-center">
            <FolderIcon className="h-8 w-8 text-indigo-500" />
            <div className="ml-4">
              <h3 className="text-lg font-semibold text-gray-900">
                {projects?.length || 0}
              </h3>
              <p className="text-gray-600">Total Projects</p>
            </div>
          </div>
        </div>

        <div className="bg-white p-6 rounded-lg shadow">
          <div className="flex items-center">
            <ClockIcon className="h-8 w-8 text-yellow-500" />
            <div className="ml-4">
              <h3 className="text-lg font-semibold text-gray-900">
                {
                  recentTasks.filter((task: any) => task.status === "pending")
                    .length
                }
              </h3>
              <p className="text-gray-600">Pending Tasks</p>
            </div>
          </div>
        </div>

        <div className="bg-white p-6 rounded-lg shadow">
          <div className="flex items-center">
            <ClockIcon className="h-8 w-8 text-green-500" />
            <div className="ml-4">
              <h3 className="text-lg font-semibold text-gray-900">
                {
                  recentTasks.filter((task: any) => task.status === "completed")
                    .length
                }
              </h3>
              <p className="text-gray-600">Completed Tasks</p>
            </div>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        {/* Recent Projects */}
        <div className="bg-white p-6 rounded-lg shadow">
          <div className="flex justify-between items-center mb-6">
            <h2 className="text-xl font-semibold text-gray-900">
              Recent Projects
            </h2>
            <Link
              href="/projects"
              className="text-indigo-600 hover:text-indigo-500 text-sm font-medium">
              View all
            </Link>
          </div>

          <div className="space-y-4">
            {projects?.slice(0, 5).map((project: any) => (
              <div
                key={project.id}
                className="flex items-center justify-between p-3 border rounded-lg hover:bg-gray-50">
                <div>
                  <h3 className="font-medium text-gray-900">{project.name}</h3>
                  <p className="text-sm text-gray-500">{project.description}</p>
                </div>
                <Link
                  href={`/projects/${project.id}`}
                  className="text-indigo-600 hover:text-indigo-500 text-sm font-medium">
                  View
                </Link>
              </div>
            ))}

            {(!projects || projects.length === 0) && (
              <div className="text-center py-8">
                <FolderIcon className="mx-auto h-12 w-12 text-gray-400" />
                <p className="mt-2 text-gray-500">No projects yet</p>
                <Link
                  href="/projects/new"
                  className="mt-4 inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700">
                  <PlusIcon className="h-4 w-4 mr-2" />
                  Create Project
                </Link>
              </div>
            )}
          </div>
        </div>

        {/* Recent Tasks */}
        <div className="bg-white p-6 rounded-lg shadow">
          <div className="flex justify-between items-center mb-6">
            <h2 className="text-xl font-semibold text-gray-900">
              Recent Tasks
            </h2>
            <Link
              href="/tasks"
              className="text-indigo-600 hover:text-indigo-500 text-sm font-medium">
              View all
            </Link>
          </div>

          <div className="space-y-4">
            {recentTasks.map((task: any) => (
              <div
                key={task.id}
                className="flex items-center justify-between p-3 border rounded-lg hover:bg-gray-50">
                <div>
                  <h3 className="font-medium text-gray-900">{task.title}</h3>
                  <p className="text-sm text-gray-500">{task.project_name}</p>
                </div>
                <span
                  className={`px-2 py-1 text-xs font-medium rounded-full ${
                    task.status === "completed"
                      ? "bg-green-100 text-green-800"
                      : task.status === "in_progress"
                      ? "bg-yellow-100 text-yellow-800"
                      : "bg-gray-100 text-gray-800"
                  }`}>
                  {task.status.replace("_", " ")}
                </span>
              </div>
            ))}

            {recentTasks.length === 0 && (
              <div className="text-center py-8">
                <ClockIcon className="mx-auto h-12 w-12 text-gray-400" />
                <p className="mt-2 text-gray-500">No tasks yet</p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
```

---

## 🎨 Styling & UI

### 1. Tailwind Configuration

Update `tailwind.config.js`:

```javascript
/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./src/pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/components/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: {
          50: "#eff6ff",
          500: "#3b82f6",
          600: "#2563eb",
          700: "#1d4ed8",
        },
      },
      animation: {
        "fade-in": "fadeIn 0.5s ease-in-out",
        "slide-up": "slideUp 0.3s ease-out",
      },
      keyframes: {
        fadeIn: {
          "0%": { opacity: "0" },
          "100%": { opacity: "1" },
        },
        slideUp: {
          "0%": { transform: "translateY(10px)", opacity: "0" },
          "100%": { transform: "translateY(0)", opacity: "1" },
        },
      },
    },
  },
  plugins: [require("@tailwindcss/forms")],
};
```

### 2. Global Styles

Update `src/app/globals.css`:

```css
@tailwind base;
@tailwind components;
@tailwind utilities;

@layer base {
  html {
    font-family: "Inter", system-ui, sans-serif;
  }
}

@layer components {
  .btn-primary {
    @apply inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed;
  }

  .btn-secondary {
    @apply inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500;
  }

  .form-input {
    @apply block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500;
  }

  .form-label {
    @apply block text-sm font-medium text-gray-700 mb-1;
  }

  .card {
    @apply bg-white rounded-lg shadow-sm border border-gray-200 p-6;
  }

  .status-badge {
    @apply px-2 py-1 text-xs font-medium rounded-full;
  }

  .status-pending {
    @apply bg-gray-100 text-gray-800;
  }

  .status-in-progress {
    @apply bg-yellow-100 text-yellow-800;
  }

  .status-completed {
    @apply bg-green-100 text-green-800;
  }
}

/* Custom scrollbar */
::-webkit-scrollbar {
  width: 6px;
}

::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 3px;
}

::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
}

::-webkit-scrollbar-thumb:hover {
  background: #a1a1a1;
}

/* Loading animations */
.animate-pulse-slow {
  animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}

/* Real-time indicators */
.typing-indicator {
  @apply flex items-center space-x-1 text-sm text-gray-500;
}

.typing-indicator .dot {
  @apply w-1 h-1 bg-gray-400 rounded-full animate-bounce;
}

.typing-indicator .dot:nth-child(2) {
  animation-delay: 0.1s;
}

.typing-indicator .dot:nth-child(3) {
  animation-delay: 0.2s;
}

/* Connection status */
.connection-status {
  @apply flex items-center space-x-2 text-sm;
}

.connection-online {
  @apply text-green-600;
}

.connection-offline {
  @apply text-red-600;
}

.connection-dot {
  @apply w-2 h-2 rounded-full;
}

.connection-dot.online {
  @apply bg-green-500;
}

.connection-dot.offline {
  @apply bg-red-500;
}
```

---

## 🚀 Deployment

### 1. Production Build

```bash
# Build the application
npm run build

# Start production server
npm start
```

### 2. Environment Variables for Production

Create `.env.production`:

```env
NEXT_PUBLIC_API_BASE_URL=https://your-api-domain.com
NEXT_PUBLIC_WS_URL=wss://your-api-domain.com
NEXT_PUBLIC_APP_NAME="Task Manager"
NEXT_PUBLIC_APP_URL=https://your-domain.com
NODE_ENV=production
```

### 3. Vercel Deployment

Create `vercel.json`:

```json
{
  "builds": [
    {
      "src": "package.json",
      "use": "@vercel/next"
    }
  ],
  "env": {
    "NEXT_PUBLIC_API_BASE_URL": "https://your-api-domain.com",
    "NEXT_PUBLIC_WS_URL": "wss://your-api-domain.com"
  }
}
```

### 4. Docker Deployment

Create `Dockerfile`:

```dockerfile
FROM node:18-alpine AS base

# Install dependencies only when needed
FROM base AS deps
WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm ci

# Rebuild the source code only when needed
FROM base AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN npm run build

# Production image, copy all the files and run next
FROM base AS runner
WORKDIR /app

ENV NODE_ENV production

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

COPY --from=builder /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs

EXPOSE 3000

ENV PORT 3000

CMD ["node", "server.js"]
```

---

## ✅ Best Practices

### 1. Code Organization

- **Components**: Keep components small and focused
- **Hooks**: Extract complex logic into custom hooks
- **Types**: Define comprehensive TypeScript interfaces
- **API**: Centralize API calls in dedicated modules

### 2. Performance Optimization

- **Server Components**: Use React Server Components where possible
- **Code Splitting**: Implement dynamic imports for large components
- **Image Optimization**: Use Next.js Image component
- **Caching**: Implement proper caching strategies

### 3. Security

- **Token Storage**: Store JWT tokens securely
- **Input Validation**: Validate all user inputs
- **CORS**: Configure CORS properly for production
- **Environment Variables**: Never expose sensitive data

### 4. User Experience

- **Loading States**: Show loading indicators
- **Error Handling**: Provide meaningful error messages
- **Real-time Feedback**: Show real-time updates
- **Offline Support**: Handle network failures gracefully

### 5. Testing

```bash
# Install testing dependencies
npm install -D @testing-library/react @testing-library/jest-dom jest jest-environment-jsdom

# Create test files
# - __tests__/components/
# - __tests__/hooks/
# - __tests__/pages/
```

---

## 📚 Additional Resources

### Documentation Links

- [Next.js 15 Documentation](https://nextjs.org/docs)
- [React Query Documentation](https://tanstack.com/query/latest)
- [Tailwind CSS Documentation](https://tailwindcss.com/docs)
- [WebSocket API Documentation](https://developer.mozilla.org/en-US/docs/Web/API/WebSocket)

### Recommended Tools

- **VS Code Extensions**: ES7+ React/Redux/React-Native snippets, Tailwind CSS IntelliSense
- **Browser Extensions**: React Developer Tools, Redux DevTools
- **Development**: Thunder Client for API testing

---

This comprehensive guide provides everything needed to build a modern, real-time task management frontend with Next.js 15 that integrates seamlessly with your Go API backend. The implementation includes authentication, real-time features, responsive design, and production-ready deployment configurations.
