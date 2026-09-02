"use client";

import * as React from "react";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Employee } from "@/types/employee";
import { AlertTriangle, Loader2 } from "lucide-react";

interface DeleteConfirmModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  employee: Employee | null;
  onConfirm: () => Promise<void>;
}

export function DeleteConfirmModal({
  open,
  onOpenChange,
  employee,
  onConfirm,
}: DeleteConfirmModalProps) {
  const [loading, setLoading] = React.useState(false);

  if (!employee) return null;

  const handleConfirm = async () => {
    setLoading(true);
    try {
      await onConfirm();
      onOpenChange(false);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent onClose={() => onOpenChange(false)} className="max-w-md">
        <DialogHeader>
          <div className="flex items-center gap-3">
            <div className="h-10 w-10 rounded-xl bg-rose-50 flex items-center justify-center text-rose-600 border border-rose-100">
              <AlertTriangle className="h-5 w-5" />
            </div>
            <div>
              <DialogTitle>Delete Employee</DialogTitle>
              <DialogDescription>
                This action cannot be undone.
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>

        <div className="p-4 rounded-xl bg-slate-50 border border-slate-100 text-sm">
          Are you sure you want to remove{" "}
          <span className="font-semibold text-slate-900">
            {employee.first_name} {employee.last_name}
          </span>{" "}
          ({employee.position} &bull; {employee.department}) from the database?
        </div>

        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={loading}
          >
            Cancel
          </Button>
          <Button
            type="button"
            variant="destructive"
            onClick={handleConfirm}
            disabled={loading}
          >
            {loading && <Loader2 className="h-4 w-4 animate-spin mr-1.5" />}
            Delete Record
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
