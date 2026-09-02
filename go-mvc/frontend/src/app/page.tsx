"use client";

import * as React from "react";
import { Employee, CreateEmployeePayload, UpdateEmployeePayload } from "@/types/employee";
import { fetchEmployees, createEmployee, updateEmployee, deleteEmployee } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { StatsCards } from "@/components/employee/StatsCards";
import { EmployeeCard } from "@/components/employee/EmployeeCard";
import { EmployeeTable } from "@/components/employee/EmployeeTable";
import { EmployeeModal } from "@/components/employee/EmployeeModal";
import { DeleteConfirmModal } from "@/components/employee/DeleteConfirmModal";
import { useToast } from "@/components/ui/toast";
import {
  Plus,
  Search,
  LayoutGrid,
  Table as TableIcon,
  RefreshCw,
  Sparkles,
  Users,
  Filter,
  CheckCircle,
  Database,
  Building2,
} from "lucide-react";

export default function Home() {
  const { toast } = useToast();

  const [employees, setEmployees] = React.useState<Employee[]>([]);
  const [totalCount, setTotalCount] = React.useState<number>(0);
  const [loading, setLoading] = React.useState<boolean>(true);
  const [backendOnline, setBackendOnline] = React.useState<boolean>(true);

  // Search & Filters
  const [searchQuery, setSearchQuery] = React.useState<string>("");
  const [selectedDept, setSelectedDept] = React.useState<string>("All");
  const [selectedStatus, setSelectedStatus] = React.useState<"all" | "active" | "inactive">("all");
  const [viewMode, setViewMode] = React.useState<"grid" | "table">("grid");

  // Modals state
  const [isModalOpen, setIsModalOpen] = React.useState<boolean>(false);
  const [editingEmployee, setEditingEmployee] = React.useState<Employee | null>(null);
  const [deletingEmployee, setDeletingEmployee] = React.useState<Employee | null>(null);
  const [isDeleteModalOpen, setIsDeleteModalOpen] = React.useState<boolean>(false);

  // Load employees from API
  const loadEmployees = React.useCallback(async () => {
    setLoading(true);
    try {
      const res = await fetchEmployees(1, 100);
      setEmployees(res.data || []);
      setTotalCount(res.total_count || 0);
      setBackendOnline(true);
    } catch (err: any) {
      setBackendOnline(false);
      toast({
        title: "Connection Error",
        description: err.message || "Failed to reach Beego API server at http://localhost:8080",
        type: "error",
      });
    } finally {
      setLoading(false);
    }
  }, [toast]);

  React.useEffect(() => {
    loadEmployees();
  }, [loadEmployees]);

  // Handle Create or Update
  const handleSaveEmployee = async (data: CreateEmployeePayload | UpdateEmployeePayload) => {
    try {
      if (editingEmployee) {
        const updated = await updateEmployee(editingEmployee.id, data);
        setEmployees((prev) => prev.map((e) => (e.id === updated.id ? updated : e)));
        toast({
          title: "Employee Updated",
          description: `${updated.first_name} ${updated.last_name}'s record has been updated.`,
          type: "success",
        });
      } else {
        const created = await createEmployee(data as CreateEmployeePayload);
        setEmployees((prev) => [created, ...prev]);
        setTotalCount((prev) => prev + 1);
        toast({
          title: "Employee Added",
          description: `${created.first_name} ${created.last_name} was successfully registered.`,
          type: "success",
        });
      }
    } catch (err: any) {
      toast({
        title: "Operation Failed",
        description: err.message || "An error occurred while saving employee data.",
        type: "error",
      });
      throw err;
    }
  };

  // Handle Delete
  const handleDeleteConfirm = async () => {
    if (!deletingEmployee) return;
    try {
      await deleteEmployee(deletingEmployee.id);
      setEmployees((prev) => prev.filter((e) => e.id !== deletingEmployee.id));
      setTotalCount((prev) => Math.max(0, prev - 1));
      toast({
        title: "Employee Removed",
        description: `${deletingEmployee.first_name} ${deletingEmployee.last_name} was deleted.`,
        type: "success",
      });
    } catch (err: any) {
      toast({
        title: "Delete Failed",
        description: err.message || "Failed to delete employee record.",
        type: "error",
      });
    }
  };

  // Departments list for dropdown
  const departments = React.useMemo(() => {
    const list = Array.from(new Set(employees.map((e) => e.department).filter(Boolean)));
    return ["All", ...list];
  }, [employees]);

  // Filtered employees
  const filteredEmployees = React.useMemo(() => {
    return employees.filter((emp) => {
      const matchesSearch =
        `${emp.first_name} ${emp.last_name}`.toLowerCase().includes(searchQuery.toLowerCase()) ||
        emp.email?.toLowerCase().includes(searchQuery.toLowerCase()) ||
        emp.position?.toLowerCase().includes(searchQuery.toLowerCase()) ||
        emp.department?.toLowerCase().includes(searchQuery.toLowerCase());

      const matchesDept = selectedDept === "All" || emp.department === selectedDept;

      const matchesStatus =
        selectedStatus === "all"
          ? true
          : selectedStatus === "active"
          ? emp.is_active
          : !emp.is_active;

      return matchesSearch && matchesDept && matchesStatus;
    });
  }, [employees, searchQuery, selectedDept, selectedStatus]);

  return (
    <div className="min-h-screen bg-slate-50 bg-mesh-pattern font-sans text-slate-900 pb-16">
      {/* Top Navigation Bar */}
      <header className="sticky top-0 z-40 border-b border-slate-200/80 bg-white/80 backdrop-blur-md">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div className="flex h-16 items-center justify-between">
            {/* Logo and Brand */}
            <div className="flex items-center gap-3">
              <div className="h-10 w-10 rounded-2xl bg-gradient-to-tr from-indigo-600 via-purple-600 to-pink-500 flex items-center justify-center text-white shadow-md shadow-indigo-500/25">
                <Sparkles className="h-5 w-5" />
              </div>
              <div>
                <div className="flex items-center gap-2">
                  <h1 className="text-lg font-black tracking-tight text-slate-900">
                    WorkPulse
                  </h1>
                  <span className="rounded-md bg-indigo-50 px-2 py-0.5 text-[10px] font-bold uppercase tracking-wider text-indigo-700">
                    v2.0
                  </span>
                </div>
                <p className="text-xs text-slate-500 hidden sm:block">
                  Beego v2 &bull; MySQL Enterprise CRUD
                </p>
              </div>
            </div>

            {/* Quick Actions & Status Badge */}
            <div className="flex items-center gap-3">
              {/* Backend Status indicator */}
              <div
                className={`hidden md:inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold border ${
                  backendOnline
                    ? "bg-emerald-50 text-emerald-700 border-emerald-200/70"
                    : "bg-rose-50 text-rose-700 border-rose-200/70"
                }`}
              >
                <Database className="h-3 w-3" />
                <span
                  className={`h-1.5 w-1.5 rounded-full ${
                    backendOnline ? "bg-emerald-500 animate-pulse" : "bg-rose-500"
                  }`}
                />
                {backendOnline ? "API Live (:8080)" : "Backend Offline"}
              </div>

              <Button
                variant="outline"
                size="sm"
                onClick={loadEmployees}
                disabled={loading}
                className="h-9 px-3 rounded-xl hover:bg-slate-100"
                title="Refresh Records"
              >
                <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} />
                <span className="hidden sm:inline">Refresh</span>
              </Button>

              <Button
                variant="gradient"
                size="sm"
                onClick={() => {
                  setEditingEmployee(null);
                  setIsModalOpen(true);
                }}
                className="h-9 px-4 rounded-xl"
              >
                <Plus className="h-4 w-4" />
                <span>Add Employee</span>
              </Button>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content Area */}
      <main className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 pt-8">
        {/* KPI Stats Section */}
        <StatsCards employees={employees} totalCount={totalCount} />

        {/* Toolbar: Search, Filters & View Toggle */}
        <div className="mb-6 flex flex-col md:flex-row items-stretch md:items-center justify-between gap-3 bg-white/80 backdrop-blur-md p-3 rounded-2xl border border-slate-200/80 shadow-xs">
          {/* Search Bar */}
          <div className="relative flex-1">
            <Search className="absolute left-3.5 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400" />
            <Input
              type="text"
              placeholder="Search by name, email, department, or role..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-9 bg-slate-50/70 border-slate-200 focus-visible:bg-white"
            />
          </div>

          {/* Department Filter */}
          <div className="flex items-center gap-2">
            <div className="relative">
              <Building2 className="absolute left-3 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-slate-400 pointer-events-none" />
              <select
                value={selectedDept}
                onChange={(e) => setSelectedDept(e.target.value)}
                className="h-10 pl-8 pr-8 rounded-xl border border-slate-200 bg-slate-50/70 text-xs font-semibold text-slate-700 shadow-xs focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-indigo-500/20"
              >
                {departments.map((dept) => (
                  <option key={dept} value={dept}>
                    {dept === "All" ? "All Departments" : dept}
                  </option>
                ))}
              </select>
            </div>

            {/* Status Pills */}
            <div className="flex rounded-xl bg-slate-100 p-1 text-xs font-semibold">
              <button
                onClick={() => setSelectedStatus("all")}
                className={`px-3 py-1.5 rounded-lg transition-all cursor-pointer ${
                  selectedStatus === "all"
                    ? "bg-white text-slate-900 shadow-xs"
                    : "text-slate-500 hover:text-slate-900"
                }`}
              >
                All
              </button>
              <button
                onClick={() => setSelectedStatus("active")}
                className={`px-3 py-1.5 rounded-lg transition-all cursor-pointer ${
                  selectedStatus === "active"
                    ? "bg-white text-emerald-700 shadow-xs"
                    : "text-slate-500 hover:text-slate-900"
                }`}
              >
                Active
              </button>
              <button
                onClick={() => setSelectedStatus("inactive")}
                className={`px-3 py-1.5 rounded-lg transition-all cursor-pointer ${
                  selectedStatus === "inactive"
                    ? "bg-white text-rose-700 shadow-xs"
                    : "text-slate-500 hover:text-slate-900"
                }`}
              >
                Inactive
              </button>
            </div>

            {/* Grid / Table Toggle */}
            <div className="hidden sm:flex rounded-xl bg-slate-100 p-1 text-slate-600">
              <button
                onClick={() => setViewMode("grid")}
                className={`p-1.5 rounded-lg transition-all cursor-pointer ${
                  viewMode === "grid"
                    ? "bg-white text-indigo-600 shadow-xs"
                    : "text-slate-400 hover:text-slate-700"
                }`}
                title="Grid View"
              >
                <LayoutGrid className="h-4 w-4" />
              </button>
              <button
                onClick={() => setViewMode("table")}
                className={`p-1.5 rounded-lg transition-all cursor-pointer ${
                  viewMode === "table"
                    ? "bg-white text-indigo-600 shadow-xs"
                    : "text-slate-400 hover:text-slate-700"
                }`}
                title="Table View"
              >
                <TableIcon className="h-4 w-4" />
              </button>
            </div>
          </div>
        </div>

        {/* Dynamic Display Area */}
        {loading ? (
          /* Loading Skeletons */
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-5">
            {[1, 2, 3, 4, 5, 6].map((i) => (
              <div
                key={i}
                className="h-56 rounded-2xl border border-slate-200/70 bg-white/70 p-5 shadow-xs animate-pulse flex flex-col justify-between"
              >
                <div className="flex items-center gap-3">
                  <div className="h-12 w-12 rounded-2xl bg-slate-200" />
                  <div className="space-y-2 flex-1">
                    <div className="h-4 w-32 bg-slate-200 rounded" />
                    <div className="h-3 w-20 bg-slate-100 rounded" />
                  </div>
                </div>
                <div className="space-y-2">
                  <div className="h-3 w-full bg-slate-100 rounded" />
                  <div className="h-3 w-2/3 bg-slate-100 rounded" />
                </div>
                <div className="h-8 bg-slate-100 rounded-xl" />
              </div>
            ))}
          </div>
        ) : filteredEmployees.length === 0 ? (
          /* Empty State */
          <div className="flex flex-col items-center justify-center rounded-3xl border border-dashed border-slate-300 bg-white/60 p-12 text-center backdrop-blur-md">
            <div className="h-16 w-16 rounded-3xl bg-indigo-50 flex items-center justify-center text-indigo-500 mb-4 shadow-inner">
              <Users className="h-8 w-8" />
            </div>
            <h3 className="text-base font-bold text-slate-900">
              No employees found
            </h3>
            <p className="text-sm text-slate-500 mt-1 max-w-sm">
              {searchQuery || selectedDept !== "All" || selectedStatus !== "all"
                ? "Try adjusting your search filters to find what you're looking for."
                : "Your organization doesn't have any registered personnel yet. Add your first employee to get started!"}
            </p>
            <Button
              variant="gradient"
              onClick={() => {
                setEditingEmployee(null);
                setIsModalOpen(true);
              }}
              className="mt-6 rounded-xl"
            >
              <Plus className="h-4 w-4 mr-1.5" />
              Add Employee
            </Button>
          </div>
        ) : viewMode === "grid" ? (
          /* Grid View */
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-5">
            {filteredEmployees.map((employee) => (
              <EmployeeCard
                key={employee.id}
                employee={employee}
                onEdit={(emp) => {
                  setEditingEmployee(emp);
                  setIsModalOpen(true);
                }}
                onDelete={(emp) => {
                  setDeletingEmployee(emp);
                  setIsDeleteModalOpen(true);
                }}
              />
            ))}
          </div>
        ) : (
          /* Table View */
          <EmployeeTable
            employees={filteredEmployees}
            onEdit={(emp) => {
              setEditingEmployee(emp);
              setIsModalOpen(true);
            }}
            onDelete={(emp) => {
              setDeletingEmployee(emp);
              setIsDeleteModalOpen(true);
            }}
          />
        )}
      </main>

      {/* Modals */}
      <EmployeeModal
        open={isModalOpen}
        onOpenChange={setIsModalOpen}
        employee={editingEmployee}
        onSubmit={handleSaveEmployee}
      />

      <DeleteConfirmModal
        open={isDeleteModalOpen}
        onOpenChange={setIsDeleteModalOpen}
        employee={deletingEmployee}
        onConfirm={handleDeleteConfirm}
      />
    </div>
  );
}
