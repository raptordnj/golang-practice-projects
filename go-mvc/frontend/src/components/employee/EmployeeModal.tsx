"use client";

import * as React from "react";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Employee, CreateEmployeePayload, UpdateEmployeePayload } from "@/types/employee";
import { Briefcase, Building2, Calendar, DollarSign, Mail, Phone, User, Loader2 } from "lucide-react";

interface EmployeeModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  employee?: Employee | null;
  onSubmit: (data: CreateEmployeePayload | UpdateEmployeePayload) => Promise<void>;
}

const DEPARTMENTS = [
  "Engineering",
  "Product",
  "Design",
  "Marketing",
  "Sales",
  "Customer Support",
  "Human Resources",
  "Finance",
];

export function EmployeeModal({
  open,
  onOpenChange,
  employee,
  onSubmit,
}: EmployeeModalProps) {
  const isEditing = !!employee;
  const [loading, setLoading] = React.useState(false);
  const [formData, setFormData] = React.useState({
    first_name: "",
    last_name: "",
    email: "",
    phone: "",
    department: "Engineering",
    position: "",
    salary: "",
    hire_date: new Date().toISOString().split("T")[0],
    is_active: true,
  });

  React.useEffect(() => {
    if (employee) {
      setFormData({
        first_name: employee.first_name,
        last_name: employee.last_name,
        email: employee.email,
        phone: employee.phone || "",
        department: employee.department,
        position: employee.position,
        salary: employee.salary.toString(),
        hire_date: employee.hire_date ? employee.hire_date.substring(0, 10) : "",
        is_active: employee.is_active,
      });
    } else {
      setFormData({
        first_name: "",
        last_name: "",
        email: "",
        phone: "",
        department: "Engineering",
        position: "",
        salary: "",
        hire_date: new Date().toISOString().split("T")[0],
        is_active: true,
      });
    }
  }, [employee, open]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      const payload = {
        ...formData,
        salary: parseFloat(formData.salary) || 0,
      };
      await onSubmit(payload);
      onOpenChange(false);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent onClose={() => onOpenChange(false)} className="max-w-xl">
        <DialogHeader>
          <div className="flex items-center gap-3 mb-1">
            <div className="h-10 w-10 rounded-xl bg-gradient-to-tr from-indigo-600 to-pink-500 flex items-center justify-center text-white shadow-md shadow-indigo-500/20">
              <User className="h-5 w-5" />
            </div>
            <div>
              <DialogTitle>{isEditing ? "Update Employee" : "Add New Employee"}</DialogTitle>
              <DialogDescription>
                {isEditing
                  ? "Make updates to the employee profile below."
                  : "Fill in the details to register a new member to the organization."}
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="text-xs font-semibold text-slate-700 mb-1.5 flex items-center gap-1.5">
                <User className="h-3.5 w-3.5 text-slate-400" /> First Name *
              </label>
              <Input
                required
                placeholder="John"
                value={formData.first_name}
                onChange={(e) => setFormData({ ...formData, first_name: e.target.value })}
              />
            </div>

            <div>
              <label className="text-xs font-semibold text-slate-700 mb-1.5 flex items-center gap-1.5">
                <User className="h-3.5 w-3.5 text-slate-400" /> Last Name *
              </label>
              <Input
                required
                placeholder="Doe"
                value={formData.last_name}
                onChange={(e) => setFormData({ ...formData, last_name: e.target.value })}
              />
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="text-xs font-semibold text-slate-700 mb-1.5 flex items-center gap-1.5">
                <Mail className="h-3.5 w-3.5 text-slate-400" /> Work Email *
              </label>
              <Input
                required
                type="email"
                placeholder="john.doe@company.com"
                value={formData.email}
                onChange={(e) => setFormData({ ...formData, email: e.target.value })}
              />
            </div>

            <div>
              <label className="text-xs font-semibold text-slate-700 mb-1.5 flex items-center gap-1.5">
                <Phone className="h-3.5 w-3.5 text-slate-400" /> Phone Number
              </label>
              <Input
                type="tel"
                placeholder="+1 (555) 019-2834"
                value={formData.phone}
                onChange={(e) => setFormData({ ...formData, phone: e.target.value })}
              />
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="text-xs font-semibold text-slate-700 mb-1.5 flex items-center gap-1.5">
                <Building2 className="h-3.5 w-3.5 text-slate-400" /> Department *
              </label>
              <select
                value={formData.department}
                onChange={(e) => setFormData({ ...formData, department: e.target.value })}
                className="flex h-10 w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm text-slate-900 shadow-xs transition-colors focus-visible:outline-none focus-visible:border-indigo-500 focus-visible:ring-2 focus-visible:ring-indigo-500/20"
              >
                {DEPARTMENTS.map((dept) => (
                  <option key={dept} value={dept}>
                    {dept}
                  </option>
                ))}
              </select>
            </div>

            <div>
              <label className="text-xs font-semibold text-slate-700 mb-1.5 flex items-center gap-1.5">
                <Briefcase className="h-3.5 w-3.5 text-slate-400" /> Job Position *
              </label>
              <Input
                required
                placeholder="Senior Engineer"
                value={formData.position}
                onChange={(e) => setFormData({ ...formData, position: e.target.value })}
              />
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="text-xs font-semibold text-slate-700 mb-1.5 flex items-center gap-1.5">
                <DollarSign className="h-3.5 w-3.5 text-slate-400" /> Annual Salary (USD) *
              </label>
              <Input
                required
                type="number"
                min="0"
                step="1000"
                placeholder="95000"
                value={formData.salary}
                onChange={(e) => setFormData({ ...formData, salary: e.target.value })}
              />
            </div>

            <div>
              <label className="text-xs font-semibold text-slate-700 mb-1.5 flex items-center gap-1.5">
                <Calendar className="h-3.5 w-3.5 text-slate-400" /> Hire Date
              </label>
              <Input
                type="date"
                value={formData.hire_date}
                onChange={(e) => setFormData({ ...formData, hire_date: e.target.value })}
              />
            </div>
          </div>

          {isEditing && (
            <div className="flex items-center gap-3 pt-2">
              <label className="text-xs font-semibold text-slate-700">Status:</label>
              <button
                type="button"
                onClick={() => setFormData({ ...formData, is_active: !formData.is_active })}
                className={`flex items-center gap-2 px-3 py-1.5 rounded-lg text-xs font-semibold border transition-all cursor-pointer ${
                  formData.is_active
                    ? "bg-emerald-50 text-emerald-700 border-emerald-200"
                    : "bg-slate-100 text-slate-600 border-slate-200"
                }`}
              >
                <span className={`h-2 w-2 rounded-full ${formData.is_active ? "bg-emerald-500 animate-pulse" : "bg-slate-400"}`} />
                {formData.is_active ? "Active Member" : "Inactive"}
              </button>
            </div>
          )}

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={loading}
            >
              Cancel
            </Button>
            <Button type="submit" variant="gradient" disabled={loading}>
              {loading && <Loader2 className="h-4 w-4 animate-spin mr-1.5" />}
              {isEditing ? "Save Changes" : "Create Employee"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
