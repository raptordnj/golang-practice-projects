export interface Employee {
  id: number;
  first_name: string;
  last_name: string;
  email: string;
  phone?: string;
  department: string;
  position: string;
  salary: number;
  hire_date: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateEmployeePayload {
  first_name: string;
  last_name: string;
  email: string;
  phone?: string;
  department: string;
  position: string;
  salary: number;
  hire_date: string;
}

export interface UpdateEmployeePayload {
  first_name?: string;
  last_name?: string;
  email?: string;
  phone?: string;
  department?: string;
  position?: string;
  salary?: number;
  hire_date?: string;
  is_active?: boolean;
}

export interface PaginatedResponse<T> {
  success: boolean;
  message: string;
  data: T[];
  page: number;
  page_size: number;
  total_count: number;
  total_pages: number;
}

export interface SingleResponse<T> {
  success: boolean;
  message: string;
  data: T;
}
