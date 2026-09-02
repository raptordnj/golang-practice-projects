"use client";

import * as React from "react";
import { Employee } from "@/types/employee";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Mail, Phone, Calendar, Edit2, Trash2, Building2, Briefcase } from "lucide-react";

interface EmployeeCardProps {
  employee: Employee;
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

export function EmployeeCard({ employee, onEdit, onDelete }: EmployeeCardProps) {
  const initials = `${employee.first_name?.[0] || ""}${employee.last_name?.[0] || ""}`.toUpperCase();
  const gradientIndex = (employee.id || 0) % AVATAR_GRADIENTS.length;
  const gradient = AVATAR_GRADIENTS[gradientIndex];

  return (
    <div className="group relative flex flex-col justify-between rounded-2xl border border-slate-200/80 bg-white/90 backdrop-blur-md p-5 shadow-xs transition-all duration-300 hover:shadow-lg hover:border-slate-300 hover:-translate-y-1">
      <div>
        {/* Card Header: Avatar, Name & Status */}
        <div className="flex items-start justify-between gap-3 mb-4">
          <div className="flex items-center gap-3">
            <div
              className={`h-12 w-12 rounded-2xl bg-gradient-to-tr ${gradient} flex items-center justify-center text-white font-bold text-base shadow-sm group-hover:scale-105 transition-transform`}
            >
              {initials}
            </div>
            <div>
              <h4 className="font-bold text-slate-900 leading-tight group-hover:text-indigo-600 transition-colors">
                {employee.first_name} {employee.last_name}
              </h4>
              <p className="text-xs font-medium text-slate-500 flex items-center gap-1 mt-0.5">
                <Briefcase className="h-3 w-3 text-slate-400" />
                {employee.position}
              </p>
            </div>
          </div>

          <Badge variant={employee.is_active ? "success" : "secondary"}>
            <span
              className={`h-1.5 w-1.5 rounded-full ${
                employee.is_active ? "bg-emerald-500 animate-pulse" : "bg-slate-400"
              }`}
            />
            {employee.is_active ? "Active" : "Inactive"}
          </Badge>
        </div>

        {/* Department & Salary pill row */}
        <div className="flex items-center gap-2 mb-4">
          <span className="inline-flex items-center gap-1 rounded-lg bg-indigo-50/80 px-2.5 py-1 text-xs font-semibold text-indigo-700">
            <Building2 className="h-3 w-3 text-indigo-500" />
            {employee.department}
          </span>
          <span className="inline-flex items-center rounded-lg bg-slate-100 px-2.5 py-1 text-xs font-bold text-slate-800">
            ${employee.salary?.toLocaleString() || "0"}
            <span className="text-[10px] text-slate-500 font-normal ml-0.5">/yr</span>
          </span>
        </div>

        {/* Contact Info */}
        <div className="space-y-1.5 py-2 border-t border-slate-100 text-xs text-slate-600">
          <a
            href={`mailto:${employee.email}`}
            className="flex items-center gap-2 truncate hover:text-indigo-600 transition-colors"
          >
            <Mail className="h-3.5 w-3.5 text-slate-400 shrink-0" />
            <span className="truncate">{employee.email}</span>
          </a>

          {employee.phone && (
            <a
              href={`tel:${employee.phone}`}
              className="flex items-center gap-2 truncate hover:text-indigo-600 transition-colors"
            >
              <Phone className="h-3.5 w-3.5 text-slate-400 shrink-0" />
              <span>{employee.phone}</span>
            </a>
          )}

          {employee.hire_date && (
            <div className="flex items-center gap-2 text-slate-400">
              <Calendar className="h-3.5 w-3.5 shrink-0" />
              <span>Joined {new Date(employee.hire_date).toLocaleDateString(undefined, { year: "numeric", month: "short", day: "numeric" })}</span>
            </div>
          )}
        </div>
      </div>

      {/* Card Actions */}
      <div className="flex items-center justify-end gap-2 pt-4 mt-3 border-t border-slate-100">
        <Button
          variant="outline"
          size="sm"
          onClick={() => onEdit(employee)}
          className="h-8 rounded-lg text-xs hover:bg-indigo-50 hover:text-indigo-600 hover:border-indigo-200"
        >
          <Edit2 className="h-3.5 w-3.5 mr-1" />
          Edit
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={() => onDelete(employee)}
          className="h-8 rounded-lg text-xs text-rose-600 hover:bg-rose-50 hover:text-rose-700 hover:border-rose-200"
        >
          <Trash2 className="h-3.5 w-3.5 mr-1" />
          Delete
        </Button>
      </div>
    </div>
  );
}
