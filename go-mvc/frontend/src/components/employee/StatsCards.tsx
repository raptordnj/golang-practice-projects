"use client";

import * as React from "react";
import { Employee } from "@/types/employee";
import { Users, UserCheck, DollarSign, Building2, TrendingUp } from "lucide-react";

interface StatsCardsProps {
  employees: Employee[];
  totalCount: number;
}

export function StatsCards({ employees, totalCount }: StatsCardsProps) {
  const activeCount = employees.filter((e) => e.is_active).length;
  const avgSalary =
    employees.length > 0
      ? Math.round(
          employees.reduce((acc, curr) => acc + (curr.salary || 0), 0) /
            employees.length
        )
      : 0;

  const uniqueDepartments = new Set(employees.map((e) => e.department)).size;

  const stats = [
    {
      title: "Total Staff",
      value: totalCount.toLocaleString(),
      subtitle: "Registered members",
      icon: Users,
      gradient: "from-indigo-500 to-blue-600",
      bgGlow: "bg-indigo-500/10",
      textColor: "text-indigo-600",
    },
    {
      title: "Active Personnel",
      value: activeCount.toLocaleString(),
      subtitle: `${totalCount > 0 ? Math.round((activeCount / totalCount) * 100) : 0}% active rate`,
      icon: UserCheck,
      gradient: "from-emerald-500 to-teal-600",
      bgGlow: "bg-emerald-500/10",
      textColor: "text-emerald-600",
    },
    {
      title: "Average Salary",
      value: `$${avgSalary.toLocaleString()}`,
      subtitle: "Annual compensation",
      icon: DollarSign,
      gradient: "from-amber-500 to-orange-600",
      bgGlow: "bg-amber-500/10",
      textColor: "text-amber-600",
    },
    {
      title: "Departments",
      value: uniqueDepartments.toString(),
      subtitle: "Active workgroups",
      icon: Building2,
      gradient: "from-purple-500 to-pink-600",
      bgGlow: "bg-purple-500/10",
      textColor: "text-purple-600",
    },
  ];

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
      {stats.map((stat, idx) => {
        const Icon = stat.icon;
        return (
          <div
            key={idx}
            className="relative overflow-hidden rounded-2xl border border-slate-200/80 bg-white/80 backdrop-blur-md p-5 shadow-xs transition-all duration-300 hover:shadow-md hover:-translate-y-0.5 group"
          >
            <div className="flex items-center justify-between">
              <span className="text-xs font-semibold uppercase tracking-wider text-slate-500">
                {stat.title}
              </span>
              <div
                className={`h-9 w-9 rounded-xl flex items-center justify-center bg-gradient-to-br ${stat.gradient} text-white shadow-xs group-hover:scale-105 transition-transform`}
              >
                <Icon className="h-4.5 w-4.5" />
              </div>
            </div>

            <div className="mt-3 flex items-baseline gap-2">
              <span className="text-2xl font-black tracking-tight text-slate-900">
                {stat.value}
              </span>
            </div>

            <div className="mt-1 flex items-center gap-1.5 text-xs text-slate-500">
              <TrendingUp className="h-3 w-3 text-emerald-500" />
              <span>{stat.subtitle}</span>
            </div>
          </div>
        );
      })}
    </div>
  );
}
