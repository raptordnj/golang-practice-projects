import {
  Employee,
  CreateEmployeePayload,
  UpdateEmployeePayload,
  PaginatedResponse,
  SingleResponse,
} from "@/types/employee";

const BASE_URL = "/api/v1/employees";

export async function fetchEmployees(
  page: number = 1,
  pageSize: number = 20
): Promise<PaginatedResponse<Employee>> {
  const res = await fetch(`${BASE_URL}?page=${page}&page_size=${pageSize}`, {
    cache: "no-store",
  });
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.message || `Failed to fetch employees: ${res.statusText}`);
  }
  return res.json();
}

export async function fetchEmployeeById(id: number): Promise<Employee> {
  const res = await fetch(`${BASE_URL}/${id}`, { cache: "no-store" });
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
  const res = await fetch(BASE_URL, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
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
  const res = await fetch(`${BASE_URL}/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || "Failed to update employee");
  }
  return json.data;
}

export async function deleteEmployee(id: number): Promise<void> {
  const res = await fetch(`${BASE_URL}/${id}`, {
    method: "DELETE",
  });
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || "Failed to delete employee");
  }
}
