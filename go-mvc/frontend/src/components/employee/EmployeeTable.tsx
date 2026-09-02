"use client";

import * as React from "react";
import { Employee } from "@/types/employee";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Edit2, Trash2, Mail, Building2 } from "lucide-react";

interface EmployeeTableProps {
  employees: Employee[];
  onEdit: (employee: Employee) => void;
  onDelete: (employee: Employee) => void;
}

const AVATAR_GRADIENTS = [
  "from-indigo-500 to-purple-600",
  "from-pink-500 to-rose-600",
  "from-cyan-500 to-blue-600",
  "from-emerald-500 to-teal-600",
  "from-amber-500 to-orange-600",
  "from-violet-500 to-fuchsia-600",
];

export function EmployeeTable({
  employees,
  onEdit,
  onDelete,
}: EmployeeTableProps) {
  return (
    <div className="overflow-hidden rounded-2xl border border-slate-200/80 bg-white/90 backdrop-blur-md shadow-xs">
      <div className="overflow-x-auto">
        <table className="w-full text-left text-sm text-slate-600">
          <thead className="bg-slate-50/80 border-b border-slate-100 text-xs font-semibold uppercase tracking-wider text-slate-500">
            <tr>
              <th className="px-6 py-4">Employee</th>
              <th className="px-6 py-4">Department & Role</th>
              <th className="px-6 py-4">Salary</th>
              <th className="px-6 py-4">Status</th>
              <th className="px-6 py-4">Hire Date</th>
              <th className="px-6 py-4 text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100">
            {employees.map((employee) => {
              const initials = `${employee.first_name?.[0] || ""}${employee.last_name?.[0] || ""}`.toUpperCase();
              const gradientIndex = (employee.id || 0) % AVATAR_GRADIENTS.length;
              const gradient = AVATAR_GRADIENTS[gradientIndex];

              return (
                <tr
                  key={employee.id}
                  className="transition-colors hover:bg-indigo-50/30 group"
                >
                  {/* Name + Avatar */}
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center gap-3">
                      <div
                        className={`h-10 w-10 rounded-xl bg-gradient-to-tr ${gradient} flex items-center justify-center text-white font-bold text-xs shadow-xs shrink-0`}
                      >
                        {initials}
                      </div>
                      <div>
                        <div className="font-semibold text-slate-900 group-hover:text-indigo-600 transition-colors">
                          {employee.first_name} {employee.last_name}
                        </div>
                        <div className="text-xs text-slate-400 flex items-center gap-1 mt-0.5">
                          <Mail className="h-3 w-3" />
                          <span>{employee.email}</span>
                        </div>
                      </div>
                    </div>
                  </td>

                  {/* Department & Role */}
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="font-medium text-slate-800">{employee.position}</div>
                    <div className="inline-flex items-center gap-1 text-xs text-indigo-600 mt-0.5">
                      <Building2 className="h-3 w-3" />
                      <span>{employee.department}</span>
                    </div>
                  </td>

                  {/* Salary */}
                  <td className="px-6 py-4 whitespace-nowrap font-semibold text-slate-900">
                    ${employee.salary?.toLocaleString() || "0"}
                    <span className="text-xs font-normal text-slate-400 ml-0.5">/yr</span>
                  </td>

                  {/* Status */}
                  <td className="px-6 py-4 whitespace-nowrap">
                    <Badge variant={employee.is_active ? "success" : "secondary"}>
                      <span
                        className={`h-1.5 w-1.5 rounded-full ${
                          employee.is_active ? "bg-emerald-500 animate-pulse" : "bg-slate-400"
                        }`}
                      />
                      {employee.is_active ? "Active" : "Inactive"}
                    </Badge>
                  </td>

                  {/* Hire Date */}
                  <td className="px-6 py-4 whitespace-nowrap text-xs text-slate-500">
                    {employee.hire_date
                      ? new Date(employee.hire_date).toLocaleDateString(undefined, {
                          year: "numeric",
                          month: "short",
                          day: "numeric",
                        })
                      : "—"}
                  </td>

                  {/* Actions */}
                  <td className="px-6 py-4 whitespace-nowrap text-right">
                    <div className="flex items-center justify-end gap-1.5">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => onEdit(employee)}
                        className="h-8 w-8 p-0 rounded-lg text-slate-500 hover:text-indigo-600 hover:bg-indigo-50"
                        title="Edit Employee"
                      >
                        <Edit2 className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => onDelete(employee)}
                        className="h-8 w-8 p-0 rounded-lg text-slate-400 hover:text-rose-600 hover:bg-rose-50"
                        title="Delete Employee"
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}
