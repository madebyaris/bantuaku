import { useEffect, useState } from "react";
import {
  Plus,
  Edit,
  Trash2,
  CheckCircle,
  XCircle,
  MessageSquare,
  Upload,
  RefreshCw,
  HardDrive,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { api } from "@/lib/api";
import { toast } from "@/components/ui/toaster";
import { cn } from "@/lib/utils";

interface SubscriptionPlan {
  id: string;
  name: string;
  display_name: string;
  price_monthly: number;
  price_yearly?: number;
  currency: string;
  max_stores?: number;
  max_products?: number;
  max_chats_per_month?: number;
  max_file_uploads_per_month?: number;
  max_file_size_mb?: number;
  max_forecast_refreshes_per_month?: number;
  features: Record<string, boolean>;
  is_active: boolean;
  created_at: string;
  updated_at?: string;
}

const initialFormState = {
  name: "",
  display_name: "",
  price_monthly: 0,
  price_yearly: undefined as number | undefined,
  currency: "IDR",
  max_stores: undefined as number | undefined,
  max_products: undefined as number | undefined,
  max_chats_per_month: undefined as number | undefined,
  max_file_uploads_per_month: undefined as number | undefined,
  max_file_size_mb: 10,
  max_forecast_refreshes_per_month: undefined as number | undefined,
  features: {} as Record<string, boolean>,
  is_active: true,
};

export function SubscriptionPlansPage() {
  const [plans, setPlans] = useState<SubscriptionPlan[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [editingPlan, setEditingPlan] = useState<SubscriptionPlan | null>(null);
  const [formData, setFormData] = useState(initialFormState);

  useEffect(() => {
    loadPlans();
  }, []);

  async function loadPlans() {
    try {
      setLoading(true);
      const response = await api.admin.plans.list(1, 100);
      setPlans(response.plans ?? []);
    } catch (error) {
      toast({
        title: "Error",
        description:
          error instanceof Error ? error.message : "Failed to load plans",
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  }

  async function handleCreate() {
    if (!formData.name.trim()) {
      toast({
        title: "Error",
        description: "Plan name is required",
        variant: "destructive",
      });
      return;
    }
    if (!formData.display_name.trim()) {
      toast({
        title: "Error",
        description: "Display name is required",
        variant: "destructive",
      });
      return;
    }

    try {
      await api.admin.plans.create({
        name: formData.name,
        display_name: formData.display_name,
        price_monthly: formData.price_monthly,
        price_yearly: formData.price_yearly,
        currency: formData.currency,
        max_stores: formData.max_stores,
        max_products: formData.max_products,
        max_chats_per_month: formData.max_chats_per_month,
        max_file_uploads_per_month: formData.max_file_uploads_per_month,
        max_file_size_mb: formData.max_file_size_mb,
        max_forecast_refreshes_per_month: formData.max_forecast_refreshes_per_month,
        features: formData.features,
      });
      toast({
        title: "Success",
        description: "Plan created successfully",
        variant: "success",
      });
      setShowCreateModal(false);
      setFormData(initialFormState);
      loadPlans();
    } catch (error) {
      toast({
        title: "Error",
        description:
          error instanceof Error ? error.message : "Failed to create plan",
        variant: "destructive",
      });
    }
  }

  async function handleUpdate() {
    if (!editingPlan) return;
    if (!formData.display_name.trim()) {
      toast({
        title: "Error",
        description: "Display name is required",
        variant: "destructive",
      });
      return;
    }

    try {
      await api.admin.plans.update(editingPlan.id, {
        display_name: formData.display_name,
        price_monthly: formData.price_monthly,
        price_yearly: formData.price_yearly,
        currency: formData.currency,
        max_stores: formData.max_stores,
        max_products: formData.max_products,
        max_chats_per_month: formData.max_chats_per_month,
        max_file_uploads_per_month: formData.max_file_uploads_per_month,
        max_file_size_mb: formData.max_file_size_mb,
        max_forecast_refreshes_per_month: formData.max_forecast_refreshes_per_month,
        features: formData.features,
        is_active: formData.is_active,
      });
      toast({
        title: "Success",
        description: "Plan updated successfully",
        variant: "success",
      });
      setEditingPlan(null);
      setFormData(initialFormState);
      loadPlans();
    } catch (error) {
      toast({
        title: "Error",
        description:
          error instanceof Error ? error.message : "Failed to update plan",
        variant: "destructive",
      });
    }
  }

  async function handleDelete(id: string, name: string) {
    if (!confirm(`Are you sure you want to deactivate the "${name}" plan?`)) return;
    try {
      await api.admin.plans.delete(id);
      toast({
        title: "Success",
        description: "Plan deactivated",
        variant: "success",
      });
      loadPlans();
    } catch (error) {
      toast({
        title: "Error",
        description:
          error instanceof Error ? error.message : "Failed to deactivate plan",
        variant: "destructive",
      });
    }
  }

  function formatPrice(price: number, currency: string) {
    return new Intl.NumberFormat("id-ID", {
      style: "currency",
      currency: currency || "IDR",
      minimumFractionDigits: 0,
    }).format(price);
  }

  function formatLimit(value: number | undefined | null) {
    if (value === null || value === undefined) return "∞";
    return value.toLocaleString("id-ID");
  }

  function openEditModal(plan: SubscriptionPlan) {
    setEditingPlan(plan);
    setFormData({
      name: plan.name,
      display_name: plan.display_name,
      price_monthly: plan.price_monthly,
      price_yearly: plan.price_yearly,
      currency: plan.currency,
      max_stores: plan.max_stores,
      max_products: plan.max_products,
      max_chats_per_month: plan.max_chats_per_month,
      max_file_uploads_per_month: plan.max_file_uploads_per_month,
      max_file_size_mb: plan.max_file_size_mb ?? 10,
      max_forecast_refreshes_per_month: plan.max_forecast_refreshes_per_month,
      features: plan.features || {},
      is_active: plan.is_active,
    });
  }

  return (
    <div className="space-y-6 animate-fade-in-up">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-display font-bold text-slate-100 mb-2">
            Subscription Plans
          </h1>
          <p className="text-slate-400">Kelola paket langganan dan batas penggunaan</p>
        </div>
        <Button
          onClick={() => {
            setFormData(initialFormState);
            setShowCreateModal(true);
          }}
          className="bg-gradient-to-r from-emerald-600 to-emerald-400 hover:from-emerald-500 hover:to-emerald-300 text-black"
        >
          <Plus className="w-4 h-4 mr-2" />
          Add Plan
        </Button>
      </div>

      {/* Plans Grid */}
      {loading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {[1, 2, 3].map((i) => (
            <Card key={i} className="p-6 hover-card-effect animate-pulse">
              <div className="h-48 bg-white/5 rounded-lg" />
            </Card>
          ))}
        </div>
      ) : plans.length === 0 ? (
        <Card className="p-12 text-center">
          <p className="text-slate-400">Belum ada subscription plan.</p>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {plans.map((plan) => (
            <Card
              key={plan.id}
              className={cn(
                "p-6 hover-card-effect relative",
                !plan.is_active && "opacity-60"
              )}
            >
              {/* Status Badge */}
              <div className="absolute top-4 right-4">
                {plan.is_active ? (
                  <span className="inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium bg-emerald-500/20 text-emerald-400 border border-emerald-500/20">
                    <CheckCircle className="w-3 h-3" />
                    Active
                  </span>
                ) : (
                  <span className="inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium bg-red-500/20 text-red-400 border border-red-500/20">
                    <XCircle className="w-3 h-3" />
                    Inactive
                  </span>
                )}
              </div>

              {/* Plan Header */}
              <div className="mb-4">
                <h3 className="text-xl font-bold text-slate-100">{plan.display_name}</h3>
                <p className="text-sm text-slate-500">{plan.name}</p>
              </div>

              {/* Pricing */}
              <div className="mb-6">
                <div className="text-3xl font-bold text-emerald-400">
                  {formatPrice(plan.price_monthly, plan.currency)}
                </div>
                <p className="text-sm text-slate-400">per bulan</p>
                {plan.price_yearly && (
                  <p className="text-sm text-slate-500">
                    {formatPrice(plan.price_yearly, plan.currency)} / tahun
                  </p>
                )}
              </div>

              {/* Limits */}
              <div className="space-y-3 mb-6">
                <div className="flex items-center gap-3 text-sm">
                  <MessageSquare className="w-4 h-4 text-blue-400" />
                  <span className="text-slate-300">
                    <span className="font-semibold">{formatLimit(plan.max_chats_per_month)}</span> chat/bulan
                  </span>
                </div>
                <div className="flex items-center gap-3 text-sm">
                  <Upload className="w-4 h-4 text-purple-400" />
                  <span className="text-slate-300">
                    <span className="font-semibold">{formatLimit(plan.max_file_uploads_per_month)}</span> upload/bulan
                  </span>
                </div>
                <div className="flex items-center gap-3 text-sm">
                  <HardDrive className="w-4 h-4 text-amber-400" />
                  <span className="text-slate-300">
                    Max <span className="font-semibold">{formatLimit(plan.max_file_size_mb)}</span> MB/file
                  </span>
                </div>
                <div className="flex items-center gap-3 text-sm">
                  <RefreshCw className="w-4 h-4 text-emerald-400" />
                  <span className="text-slate-300">
                    <span className="font-semibold">{formatLimit(plan.max_forecast_refreshes_per_month)}</span> refresh/bulan
                  </span>
                </div>
              </div>

              {/* Actions */}
              <div className="flex gap-2 pt-4 border-t border-white/10">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => openEditModal(plan)}
                  className="flex-1 border-white/10 text-slate-300 hover:bg-white/5"
                >
                  <Edit className="w-4 h-4 mr-2" />
                  Edit
                </Button>
                {plan.is_active && (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => handleDelete(plan.id, plan.display_name)}
                    className="border-red-500/20 text-red-400 hover:bg-red-500/10"
                  >
                    <Trash2 className="w-4 h-4" />
                  </Button>
                )}
              </div>
            </Card>
          ))}
        </div>
      )}

      {/* Create Modal */}
      {showCreateModal && (
        <PlanFormModal
          title="Create New Plan"
          formData={formData}
          setFormData={setFormData}
          onSave={handleCreate}
          onClose={() => setShowCreateModal(false)}
          isCreate={true}
        />
      )}

      {/* Edit Modal */}
      {editingPlan && (
        <PlanFormModal
          title={`Edit ${editingPlan.display_name}`}
          formData={formData}
          setFormData={setFormData}
          onSave={handleUpdate}
          onClose={() => {
            setEditingPlan(null);
            setFormData(initialFormState);
          }}
          isCreate={false}
        />
      )}
    </div>
  );
}

// Form Modal Component
interface PlanFormModalProps {
  title: string;
  formData: typeof initialFormState;
  setFormData: React.Dispatch<React.SetStateAction<typeof initialFormState>>;
  onSave: () => void;
  onClose: () => void;
  isCreate: boolean;
}

function PlanFormModal({
  title,
  formData,
  setFormData,
  onSave,
  onClose,
  isCreate,
}: PlanFormModalProps) {
  return (
    <div
      className="fixed inset-0 z-[100] bg-black/85 backdrop-blur-sm overflow-y-auto"
      onClick={onClose}
    >
      <div className="flex min-h-full items-center justify-center p-4">
        <Card
          className="w-full max-w-2xl p-6 bg-[#0a0a0a] border border-white/10 shadow-2xl my-8"
          onClick={(e) => e.stopPropagation()}
        >
          <h2 className="text-xl font-bold text-slate-100 mb-6">{title}</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Basic Info */}
            {isCreate && (
              <div className="space-y-2">
                <Label htmlFor="name" className="text-slate-200">
                  Plan Name (slug)
                </Label>
                <Input
                  id="name"
                  placeholder="e.g., starter, business"
                  value={formData.name}
                  onChange={(e) =>
                    setFormData({ ...formData, name: e.target.value.toLowerCase().replace(/\s+/g, '_') })
                  }
                  className="bg-white/5 border-white/10 text-slate-100"
                />
              </div>
            )}
            <div className="space-y-2">
              <Label htmlFor="display_name" className="text-slate-200">
                Display Name
              </Label>
              <Input
                id="display_name"
                placeholder="e.g., Starter Plan"
                value={formData.display_name}
                onChange={(e) =>
                  setFormData({ ...formData, display_name: e.target.value })
                }
                className="bg-white/5 border-white/10 text-slate-100"
              />
            </div>

            {/* Pricing */}
            <div className="space-y-2">
              <Label htmlFor="price_monthly" className="text-slate-200">
                Price Monthly (IDR)
              </Label>
              <Input
                id="price_monthly"
                type="number"
                placeholder="0"
                value={formData.price_monthly || ""}
                onChange={(e) =>
                  setFormData({ ...formData, price_monthly: Number(e.target.value) || 0 })
                }
                className="bg-white/5 border-white/10 text-slate-100"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="price_yearly" className="text-slate-200">
                Price Yearly (IDR) - Optional
              </Label>
              <Input
                id="price_yearly"
                type="number"
                placeholder="Optional"
                value={formData.price_yearly || ""}
                onChange={(e) =>
                  setFormData({ ...formData, price_yearly: e.target.value ? Number(e.target.value) : undefined })
                }
                className="bg-white/5 border-white/10 text-slate-100"
              />
            </div>

            {/* Limits Section */}
            <div className="col-span-full mt-4">
              <h3 className="text-lg font-semibold text-slate-200 mb-3 flex items-center gap-2">
                <span className="w-8 h-[1px] bg-emerald-500" />
                Usage Limits
                <span className="flex-1 h-[1px] bg-white/10" />
              </h3>
              <p className="text-xs text-slate-500 mb-4">Leave empty for unlimited</p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="max_chats" className="text-slate-200 flex items-center gap-2">
                <MessageSquare className="w-4 h-4 text-blue-400" />
                Chats per Month
              </Label>
              <Input
                id="max_chats"
                type="number"
                placeholder="Unlimited"
                value={formData.max_chats_per_month ?? ""}
                onChange={(e) =>
                  setFormData({ ...formData, max_chats_per_month: e.target.value ? Number(e.target.value) : undefined })
                }
                className="bg-white/5 border-white/10 text-slate-100"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="max_uploads" className="text-slate-200 flex items-center gap-2">
                <Upload className="w-4 h-4 text-purple-400" />
                File Uploads per Month
              </Label>
              <Input
                id="max_uploads"
                type="number"
                placeholder="Unlimited"
                value={formData.max_file_uploads_per_month ?? ""}
                onChange={(e) =>
                  setFormData({ ...formData, max_file_uploads_per_month: e.target.value ? Number(e.target.value) : undefined })
                }
                className="bg-white/5 border-white/10 text-slate-100"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="max_file_size" className="text-slate-200 flex items-center gap-2">
                <HardDrive className="w-4 h-4 text-amber-400" />
                Max File Size (MB)
              </Label>
              <Input
                id="max_file_size"
                type="number"
                placeholder="10"
                value={formData.max_file_size_mb ?? ""}
                onChange={(e) =>
                  setFormData({ ...formData, max_file_size_mb: e.target.value ? Number(e.target.value) : 10 })
                }
                className="bg-white/5 border-white/10 text-slate-100"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="max_forecasts" className="text-slate-200 flex items-center gap-2">
                <RefreshCw className="w-4 h-4 text-emerald-400" />
                Forecast Refreshes per Month
              </Label>
              <Input
                id="max_forecasts"
                type="number"
                placeholder="Unlimited"
                value={formData.max_forecast_refreshes_per_month ?? ""}
                onChange={(e) =>
                  setFormData({ ...formData, max_forecast_refreshes_per_month: e.target.value ? Number(e.target.value) : undefined })
                }
                className="bg-white/5 border-white/10 text-slate-100"
              />
            </div>

            {/* Status Toggle (edit only) */}
            {!isCreate && (
              <div className="col-span-full">
                <Label className="text-slate-200 flex items-center gap-3 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={formData.is_active}
                    onChange={(e) =>
                      setFormData({ ...formData, is_active: e.target.checked })
                    }
                    className="w-4 h-4 rounded border-white/10 bg-white/5 text-emerald-500 focus:ring-emerald-500/20"
                  />
                  Plan is Active
                </Label>
              </div>
            )}
          </div>

          {/* Actions */}
          <div className="flex gap-3 pt-6 mt-6 border-t border-white/10">
            <Button
              onClick={onSave}
              className="flex-1 bg-gradient-to-r from-emerald-600 to-emerald-400 hover:from-emerald-500 hover:to-emerald-300 text-black font-semibold"
            >
              {isCreate ? "Create Plan" : "Update Plan"}
            </Button>
            <Button
              variant="outline"
              onClick={onClose}
              className="border-white/10 text-slate-300 hover:bg-white/5"
            >
              Cancel
            </Button>
          </div>
        </Card>
      </div>
    </div>
  );
}
