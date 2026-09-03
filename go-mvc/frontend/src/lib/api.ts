import {
  Employee,
  CreateEmployeePayload,
  UpdateEmployeePayload,
  PaginatedResponse,
  SingleResponse,
} from "@/types/employee";
import { AuthResponse, LoginPayload, RegisterPayload, User } from "@/types/auth";

const EMPLOYEES_URL = "/api/v1/employees";
const AUTH_URL = "/api/v1/auth";

const TOKEN_KEY = "workpulse_auth_token";

export function getStoredToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(TOKEN_KEY);
}

export function setStoredToken(token: string): void {
  if (typeof window !== "undefined") {
    localStorage.setItem(TOKEN_KEY, token);
  }
}

export function clearStoredToken(): void {
  if (typeof window !== "undefined") {
    localStorage.removeItem(TOKEN_KEY);
  }
}

function getAuthHeaders(): HeadersInit {
  const token = getStoredToken();
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }
  return headers;
}

export async function loginApi(payload: LoginPayload): Promise<AuthResponse> {
  const res = await fetch(`${AUTH_URL}/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || "Failed to log in");
  }
  return json.data;
}

export async function registerApi(payload: RegisterPayload): Promise<AuthResponse> {
  const res = await fetch(`${AUTH_URL}/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || "Failed to register");
  }
  return json.data;
}

export async function fetchCurrentUser(): Promise<User> {
  const res = await fetch(`${AUTH_URL}/me`, {
    headers: getAuthHeaders(),
    cache: "no-store",
  });
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || "Failed to fetch current user");
  }
  return json.data;
}

export async function fetchEmployees(
  page: number = 1,
  pageSize: number = 20
): Promise<PaginatedResponse<Employee>> {
  const res = await fetch(`${EMPLOYEES_URL}?page=${page}&page_size=${pageSize}`, {
    headers: getAuthHeaders(),
    cache: "no-store",
  });
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.message || `Failed to fetch employees: ${res.statusText}`);
  }
  return res.json();
}

export async function fetchEmployeeById(id: number): Promise<Employee> {
  const res = await fetch(`${EMPLOYEES_URL}/${id}`, {
    headers: getAuthHeaders(),
    cache: "no-store",
  });
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.message || `Failed to fetch employee: ${res.statusText}`);
  }
  const json: SingleResponse<Employee> = await res.json();
  return json.data;
}

export async function createEmployee(
  payload: CreateEmployeePayload
): Promise<Employee> {
  const res = await fetch(EMPLOYEES_URL, {
    method: "POST",
    headers: getAuthHeaders(),
    body: JSON.stringify(payload),
  });
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || "Failed to create employee");
  }
  return json.data;
}

export async function updateEmployee(
  id: number,
  payload: UpdateEmployeePayload
): Promise<Employee> {
  const res = await fetch(`${EMPLOYEES_URL}/${id}`, {
    method: "PUT",
    headers: getAuthHeaders(),
    body: JSON.stringify(payload),
  });
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || "Failed to update employee");
  }
  return json.data;
}

export async function deleteEmployee(id: number): Promise<void> {
  const res = await fetch(`${EMPLOYEES_URL}/${id}`, {
    method: "DELETE",
    headers: getAuthHeaders(),
  });
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || "Failed to delete employee");
  }
}
