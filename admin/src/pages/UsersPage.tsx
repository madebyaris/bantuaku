import { useEffect, useState } from "react";
import {
  Plus,
  Edit,
  Trash2,
  Shield,
  User,
  Ban,
  CheckCircle,
  Crown,
  MoreVertical,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { api } from "@/lib/api";
import { toast } from "@/components/ui/toaster";
import { formatDateShort } from "@/lib/utils";
import { cn } from "@/lib/utils";
import { useAuthStore } from "@/state/auth";

interface User {
  id: string;
  email: string;
  role: string;
  status: string;
  store_name?: string;
  industry?: string;
  subscription_plan?: string;
  subscription_status?: string;
  created_at: string;
  updated_at?: string;
}

const BUSINESS_CATEGORIES = [
  { value: "retail", label: "Retail / Toko" },
  { value: "food", label: "Makanan & Minuman" },
  { value: "fashion", label: "Fashion & Pakaian" },
  { value: "beauty", label: "Kecantikan & Skincare" },
  { value: "electronics", label: "Elektronik & Gadget" },
  { value: "other", label: "Lainnya" },
];

const initialFormState = {
  email: "",
  password: "",
  role: "user",
  storeName: "",
  industry: "",
};

export function UsersPage() {
  const currentUser = useAuthStore((state) => state.user);
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [actionUser, setActionUser] = useState<User | null>(null);
  const [formData, setFormData] = useState(initialFormState);

  useEffect(() => {
    loadUsers();
  }, [page]);

  // Debug: Log when editingUser changes to verify data is loaded
  useEffect(() => {
    if (editingUser) {
      console.log("Editing user:", editingUser);
      console.log("Form data will be set to:", {
        email: editingUser.email,
        role: editingUser.role,
        storeName: editingUser.store_name,
        industry: editingUser.industry,
      });
    }
  }, [editingUser]);

  async function loadUsers() {
    try {
      setLoading(true);
      const response = await api.admin.users.list(page, 20);
      setUsers(response.users ?? []);
      setTotal(response.pagination?.total ?? 0);
    } catch (error) {
      toast({
        title: "Error",
        description:
          error instanceof Error ? error.message : "Failed to load users",
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  }

  async function handleCreate() {
    if (!formData.storeName.trim()) {
      toast({
        title: "Nama toko wajib diisi",
        description: "Lengkapi Nama Toko sebelum membuat user.",
        variant: "destructive",
      });
      return;
    }

    try {
      await api.admin.users.create({
        email: formData.email,
        password: formData.password,
        role: formData.role,
        store_name: formData.storeName,
        industry: formData.industry,
      });
      toast({
        title: "Success",
        description: "User created successfully",
        variant: "success",
      });
      setShowCreateModal(false);
      setFormData(initialFormState);
      loadUsers();
    } catch (error) {
      toast({
        title: "Error",
        description:
          error instanceof Error ? error.message : "Failed to create user",
        variant: "destructive",
      });
    }
  }

  async function handleUpdate() {
    if (!editingUser) return;
    if (!formData.storeName.trim()) {
      toast({
        title: "Nama toko wajib diisi",
        description: "Lengkapi Nama Toko sebelum mengupdate user.",
        variant: "destructive",
      });
      return;
    }
    try {
      await api.admin.users.update(editingUser.id, {
        email: formData.email,
        store_name: formData.storeName,
        industry: formData.industry,
      });
      toast({
        title: "Success",
        description: "User updated successfully",
        variant: "success",
      });
      setEditingUser(null);
      setFormData(initialFormState);
      loadUsers();
    } catch (error) {
      toast({
        title: "Error",
        description:
          error instanceof Error ? error.message : "Failed to update user",
        variant: "destructive",
      });
    }
  }

  async function handleSuspend(id: string) {
    // Prevent admins from suspending themselves
    if (currentUser && currentUser.id === id) {
      toast({
        title: "Error",
        description: "You cannot suspend your own account",
        variant: "destructive",
      });
      return;
    }
    if (!confirm("Are you sure you want to suspend this user?")) return;
    try {
      await api.admin.users.updateStatus(id, "suspended");
      toast({
        title: "Success",
        description: "User suspended",
        variant: "success",
      });
      loadUsers();
      setActionUser(null);
    } catch (error) {
      toast({
        title: "Error",
        description:
          error instanceof Error ? error.message : "Failed to suspend user",
        variant: "destructive",
      });
    }
  }

  async function handleUnsuspend(id: string) {
    try {
      await api.admin.users.updateStatus(id, "active");
      toast({
        title: "Success",
        description: "User activated",
        variant: "success",
      });
      loadUsers();
      setActionUser(null);
    } catch (error) {
      toast({
        title: "Error",
        description:
          error instanceof Error ? error.message : "Failed to activate user",
        variant: "destructive",
      });
    }
  }

  async function handleDelete(id: string) {
    // Prevent admins from deleting themselves
    if (currentUser && currentUser.id === id) {
      toast({
        title: "Error",
        description: "You cannot delete your own account",
        variant: "destructive",
      });
      return;
    }
    if (!confirm("Are you sure you want to delete this user?")) return;
    try {
      await api.admin.users.delete(id);
      toast({
        title: "Success",
        description: "User deleted",
        variant: "success",
      });
      loadUsers();
      setActionUser(null);
    } catch (error) {
      toast({
        title: "Error",
        description:
          error instanceof Error ? error.message : "Failed to delete user",
        variant: "destructive",
      });
    }
  }

  const getRoleBadge = (role: string) => {
    const colors = {
      super_admin: "bg-purple-500/20 text-purple-400 border-purple-500/20",
      admin: "bg-emerald-500/20 text-emerald-400 border-emerald-500/20",
      user: "bg-slate-500/20 text-slate-400 border-slate-500/20",
    };
    return colors[role as keyof typeof colors] || colors.user;
  };

  const getStatusBadge = (status: string) => {
    const styles = {
      active: {
        wrapper:
          "bg-emerald-500/15 text-emerald-200 border border-emerald-500/30",
        label: "Active",
        icon: <CheckCircle className="w-3.5 h-3.5" />,
      },
      suspended: {
        wrapper: "bg-red-500/15 text-red-200 border border-red-500/30",
        label: "Suspended",
        icon: <Ban className="w-3.5 h-3.5" />,
      },
      deleted: {
        wrapper: "bg-slate-500/15 text-slate-200 border border-slate-500/30",
        label: "Deleted",
        icon: <Trash2 className="w-3.5 h-3.5" />,
      },
    };
    return styles[status as keyof typeof styles] || styles.active;
  };

  const getSubscriptionBadge = (plan?: string, status?: string) => {
    const planColors = {
      free: "bg-slate-500/20 text-slate-400 border-slate-500/20",
      pro: "bg-emerald-500/20 text-emerald-400 border-emerald-500/20",
      enterprise: "bg-purple-500/20 text-purple-400 border-purple-500/20",
    };
    const planLabel = plan ? plan.charAt(0).toUpperCase() + plan.slice(1) : "Free";
    const color = planColors[plan as keyof typeof planColors] || planColors.free;
    
    return {
      wrapper: `inline-flex px-2 py-1 rounded-full text-xs font-medium border ${color}`,
      label: planLabel,
      status: status || "",
    };
  };

  return (
    <div className="space-y-6 animate-fade-in-up">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-display font-bold text-slate-100 mb-2">
            User Management
          </h1>
          <p className="text-slate-400">Kelola pengguna dan peran akses</p>
        </div>
        <Button
          onClick={() => {
            setFormData(initialFormState);
            setShowCreateModal(true);
          }}
          className="bg-gradient-to-r from-emerald-600 to-emerald-400 hover:from-emerald-500 hover:to-emerald-300 text-black"
        >
          <Plus className="w-4 h-4 mr-2" />
          Add User
        </Button>
      </div>

      <Card className="p-6 hover-card-effect">
        {loading ? (
          <div className="text-center py-12 text-slate-400">Loading...</div>
        ) : users.length === 0 ? (
          <div className="text-center py-12 text-slate-400">
            Belum ada pengguna.
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-white/10">
                  <th className="text-left py-3 px-4 text-sm font-semibold text-slate-300">
                    Email
                  </th>
                  <th className="text-left py-3 px-4 text-sm font-semibold text-slate-300">
                    Role
                  </th>
                  <th className="text-left py-3 px-4 text-sm font-semibold text-slate-300">
                    Status
                  </th>
                  <th className="text-left py-3 px-4 text-sm font-semibold text-slate-300">
                    Subscription
                  </th>
                  <th className="text-left py-3 px-4 text-sm font-semibold text-slate-300">
                    Created
                  </th>
                  <th className="text-right py-3 px-4 text-sm font-semibold text-slate-300">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody>
                {users.map((user) => (
                  <tr
                    key={user.id}
                    className="border-b border-white/5 hover:bg-white/5 transition-colors"
                  >
                    <td className="py-3 px-4">
                      <div className="flex items-center gap-2">
                        {user.role === "admin" ||
                        user.role === "super_admin" ? (
                          <Shield className="w-4 h-4 text-emerald-400" />
                        ) : (
                          <User className="w-4 h-4 text-slate-500" />
                        )}
                        <span
                          className={cn(
                            "text-slate-200",
                            user.status === "suspended" &&
                              "line-through opacity-50"
                          )}
                        >
                          {user.email}
                        </span>
                      </div>
                    </td>
                    <td className="py-3 px-4">
                      <span
                        className={cn(
                          "inline-flex px-2 py-1 rounded-full text-xs font-medium border",
                          getRoleBadge(user.role)
                        )}
                      >
                        {user.role}
                      </span>
                    </td>
                    <td className="py-3 px-4">
                      {(() => {
                        const statusBadge = getStatusBadge(user.status);
                        return (
                          <span
                            className={cn(
                              "inline-flex items-center gap-2 px-3 py-1 rounded-full text-[11px] font-semibold uppercase tracking-wide",
                              statusBadge.wrapper
                            )}
                          >
                            {statusBadge.icon}
                            <span>{statusBadge.label}</span>
                          </span>
                        );
                      })()}
                    </td>
                    <td className="py-3 px-4">
                      {(() => {
                        const subBadge = getSubscriptionBadge(user.subscription_plan, user.subscription_status);
                        return (
                          <div className="flex flex-col gap-1">
                            <span className={cn(subBadge.wrapper)}>
                              {subBadge.label}
                            </span>
                            {subBadge.status && (
                              <span className="text-xs text-slate-500">
                                {subBadge.status}
                              </span>
                            )}
                          </div>
                        );
                      })()}
                    </td>
                    <td className="py-3 px-4 text-sm text-slate-400">
                      {formatDateShort(user.created_at)}
                    </td>
                    <td className="py-3 px-4">
                      <div className="flex items-center justify-end gap-2">
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => setActionUser(user)}
                          className="h-8 w-8 text-slate-400 hover:text-emerald-400"
                        >
                          <MoreVertical className="w-4 h-4" />
                        </Button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {total > 20 && (
          <div className="flex items-center justify-between mt-4 pt-4 border-t border-white/10">
            <p className="text-sm text-slate-400">
              Showing {(page - 1) * 20 + 1} to {Math.min(page * 20, total)} of{" "}
              {total} users
            </p>
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
              >
                Previous
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setPage((p) => p + 1)}
                disabled={page * 20 >= total}
              >
                Next
              </Button>
            </div>
          </div>
        )}
      </Card>

      {/* Create Modal */}
      {showCreateModal && (
        <div
          className="fixed inset-0 z-[100] bg-black/85 backdrop-blur-sm"
          onClick={() => setShowCreateModal(false)}
        >
          <div className="flex min-h-full items-center justify-center p-4">
            <Card
              className="w-full max-w-md p-6 bg-[#0a0a0a] border border-white/10 shadow-2xl"
              onClick={(e) => e.stopPropagation()}
            >
              <h2 className="text-xl font-bold text-slate-100 mb-6">
                Create New User
              </h2>
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="email" className="text-slate-200">
                    Email
                  </Label>
                  <Input
                    id="email"
                    type="email"
                    placeholder="user@example.com"
                    value={formData.email}
                    onChange={(e) =>
                      setFormData({ ...formData, email: e.target.value })
                    }
                    className="bg-white/5 border-white/10 text-slate-100 placeholder:text-slate-500 focus:border-emerald-500/50 focus:ring-emerald-500/20"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="password" className="text-slate-200">
                    Password
                  </Label>
                  <Input
                    id="password"
                    type="password"
                    placeholder="Minimal 6 karakter"
                    value={formData.password}
                    onChange={(e) =>
                      setFormData({ ...formData, password: e.target.value })
                    }
                    className="bg-white/5 border-white/10 text-slate-100 placeholder:text-slate-500 focus:border-emerald-500/50 focus:ring-emerald-500/20"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="role" className="text-slate-200">
                    Role
                  </Label>
                  <Select
                    value={formData.role}
                    onValueChange={(value) =>
                      setFormData({ ...formData, role: value })
                    }
                  >
                    <SelectTrigger className="bg-white/5 border-white/10 text-slate-100 focus:border-emerald-500/50 focus:ring-emerald-500/20">
                      <SelectValue placeholder="Pilih role..." />
                    </SelectTrigger>
                    <SelectContent className="bg-[#0a0a0a] border-white/10">
                      <SelectItem
                        value="user"
                        className="text-slate-100 focus:bg-emerald-500/20 focus:text-emerald-400"
                      >
                        User
                      </SelectItem>
                      <SelectItem
                        value="admin"
                        className="text-slate-100 focus:bg-emerald-500/20 focus:text-emerald-400"
                      >
                        Admin
                      </SelectItem>
                      <SelectItem
                        value="super_admin"
                        className="text-slate-100 focus:bg-emerald-500/20 focus:text-emerald-400"
                      >
                        Super Admin
                      </SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="storeName" className="text-slate-200">
                    Nama Toko
                  </Label>
                  <Input
                    id="storeName"
                    value={formData.storeName}
                    onChange={(e) =>
                      setFormData({ ...formData, storeName: e.target.value })
                    }
                    placeholder="Contoh: Toko Berkah Jaya"
                    className="bg-white/5 border-white/10 text-slate-100 placeholder:text-slate-500 focus:border-emerald-500/50 focus:ring-emerald-500/20"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="industry" className="text-slate-200">
                    Kategori Bisnis
                  </Label>
                  <Select
                    value={formData.industry}
                    onValueChange={(value) =>
                      setFormData({ ...formData, industry: value })
                    }
                  >
                    <SelectTrigger className="bg-white/5 border-white/10 text-slate-100 focus:border-emerald-500/50 focus:ring-emerald-500/20">
                      <SelectValue placeholder="Pilih kategori..." />
                    </SelectTrigger>
                    <SelectContent className="bg-[#0a0a0a] border-white/10">
                      {BUSINESS_CATEGORIES.map((category) => (
                        <SelectItem
                          key={category.value}
                          value={category.value}
                          className="text-slate-100 focus:bg-emerald-500/20 focus:text-emerald-400"
                        >
                          {category.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="flex gap-3 pt-4">
                  <Button
                    onClick={handleCreate}
                    className="flex-1 bg-gradient-to-r from-emerald-600 to-emerald-400 hover:from-emerald-500 hover:to-emerald-300 text-black font-semibold"
                  >
                    Create User
                  </Button>
                  <Button
                    variant="outline"
                    onClick={() => setShowCreateModal(false)}
                    className="border-white/10 text-slate-300 hover:bg-white/5"
                  >
                    Cancel
                  </Button>
                </div>
              </div>
            </Card>
          </div>
        </div>
      )}

      {/* Edit Modal */}
      {editingUser && (
        <div
          className="fixed inset-0 z-[100] bg-black/85 backdrop-blur-sm"
          onClick={() => setEditingUser(null)}
        >
          <div className="flex min-h-full items-center justify-center p-4">
            <Card
              className="w-full max-w-md p-6 bg-[#0a0a0a] border border-white/10 shadow-2xl"
              onClick={(e) => e.stopPropagation()}
            >
              <h2 className="text-xl font-bold text-slate-100 mb-6">
                Edit User
              </h2>
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="edit-email" className="text-slate-200">
                    Email
                  </Label>
                  <Input
                    id="edit-email"
                    type="email"
                    value={formData.email}
                    onChange={(e) =>
                      setFormData({ ...formData, email: e.target.value })
                    }
                    className="bg-white/5 border-white/10 text-slate-100 placeholder:text-slate-500 focus:border-emerald-500/50 focus:ring-emerald-500/20"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="edit-role" className="text-slate-200">
                    Role
                  </Label>
                  <Select
                    value={formData.role}
                    onValueChange={(value) =>
                      setFormData({ ...formData, role: value })
                    }
                  >
                    <SelectTrigger className="bg-white/5 border-white/10 text-slate-100 focus:border-emerald-500/50 focus:ring-emerald-500/20">
                      <SelectValue placeholder="Pilih role..." />
                    </SelectTrigger>
                    <SelectContent className="bg-[#0a0a0a] border-white/10">
                      <SelectItem
                        value="user"
                        className="text-slate-100 focus:bg-emerald-500/20 focus:text-emerald-400"
                      >
                        User
                      </SelectItem>
                      <SelectItem
                        value="admin"
                        className="text-slate-100 focus:bg-emerald-500/20 focus:text-emerald-400"
                      >
                        Admin
                      </SelectItem>
                      <SelectItem
                        value="super_admin"
                        className="text-slate-100 focus:bg-emerald-500/20 focus:text-emerald-400"
                      >
                        Super Admin
                      </SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="edit-storeName" className="text-slate-200">
                    Nama Toko
                  </Label>
                  <Input
                    id="edit-storeName"
                    value={formData.storeName}
                    onChange={(e) =>
                      setFormData({ ...formData, storeName: e.target.value })
                    }
                    placeholder="Contoh: Toko Berkah Jaya"
                    className="bg-white/5 border-white/10 text-slate-100 placeholder:text-slate-500 focus:border-emerald-500/50 focus:ring-emerald-500/20"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="edit-industry" className="text-slate-200">
                    Kategori Bisnis
                  </Label>
                  <Select
                    value={formData.industry || undefined}
                    onValueChange={(value) =>
                      setFormData({ ...formData, industry: value })
                    }
                  >
                    <SelectTrigger className="bg-white/5 border-white/10 text-slate-100 focus:border-emerald-500/50 focus:ring-emerald-500/20">
                      <SelectValue placeholder="Pilih kategori...">
                        {formData.industry
                          ? BUSINESS_CATEGORIES.find(
                              (c) => c.value === formData.industry
                            )?.label || formData.industry
                          : null}
                      </SelectValue>
                    </SelectTrigger>
                    <SelectContent className="bg-[#0a0a0a] border-white/10">
                      {BUSINESS_CATEGORIES.map((category) => (
                        <SelectItem
                          key={category.value}
                          value={category.value}
                          className="text-slate-100 focus:bg-emerald-500/20 focus:text-emerald-400"
                        >
                          {category.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="flex gap-3 pt-4">
                  <Button
                    onClick={handleUpdate}
                    className="flex-1 bg-gradient-to-r from-emerald-600 to-emerald-400 hover:from-emerald-500 hover:to-emerald-300 text-black font-semibold"
                  >
                    Update User
                  </Button>
                  <Button
                    variant="outline"
                    onClick={() => setEditingUser(null)}
                    className="border-white/10 text-slate-300 hover:bg-white/5"
                  >
                    Cancel
                  </Button>
                </div>
              </div>
            </Card>
          </div>
        </div>
      )}

      {/* Actions Modal */}
      {actionUser && (
        <div
          className="fixed inset-0 z-[100] bg-black/85 backdrop-blur-sm"
          onClick={() => setActionUser(null)}
        >
          <div className="flex min-h-full items-center justify-center p-4">
            <Card
              className="w-full max-w-sm p-6 bg-[#0a0a0a] border border-white/10 shadow-2xl"
              onClick={(e) => e.stopPropagation()}
            >
              <div className="mb-5">
                <h2 className="text-xl font-bold text-slate-100">
                  User Actions
                </h2>
                <p className="text-sm text-slate-400 truncate mt-1">
                  {actionUser.email}
                </p>
              </div>
              <div className="space-y-2">
                <Button
                  variant="outline"
                  className="w-full justify-start gap-3 text-slate-200 border-white/10 hover:bg-white/5 hover:text-emerald-400 hover:border-emerald-500/30"
                  onClick={async () => {
                    // Fetch full user data including company info
                    try {
                      const userData = await api.admin.users.get(actionUser.id);
                      console.log("Fetched user data:", userData); // Debug log
                      setEditingUser(userData);
                      const newFormData = {
                        email: userData.email,
                        password: "",
                        role: userData.role,
                        storeName: userData.store_name || "",
                        industry: userData.industry || "",
                      };
                      console.log("Setting formData:", newFormData); // Debug log
                      setFormData(newFormData);
                      setActionUser(null);
                    } catch (error) {
                      console.error("Error fetching user data:", error);
                      // Fallback to actionUser data if fetch fails
                      setEditingUser(actionUser);
                      setFormData({
                        email: actionUser.email,
                        password: "",
                        role: actionUser.role,
                        storeName: actionUser.store_name || "",
                        industry: actionUser.industry || "",
                      });
                      setActionUser(null);
                    }
                  }}
                >
                  <Edit className="w-4 h-4" />
                  Edit user data & role
                </Button>
                {actionUser.status === "active" ? (
                  <Button
                    variant="outline"
                    className="w-full justify-start gap-3 text-red-400 border-white/10 hover:bg-red-500/10 hover:border-red-500/30"
                    onClick={() => handleSuspend(actionUser.id)}
                    disabled={currentUser?.id === actionUser.id}
                    title={currentUser?.id === actionUser.id ? "You cannot suspend your own account" : ""}
                  >
                    <Ban className="w-4 h-4" />
                    Suspend user
                  </Button>
                ) : (
                  <Button
                    variant="outline"
                    className="w-full justify-start gap-3 text-emerald-400 border-white/10 hover:bg-emerald-500/10 hover:border-emerald-500/30"
                    onClick={() => handleUnsuspend(actionUser.id)}
                  >
                    <CheckCircle className="w-4 h-4" />
                    Activate user
                  </Button>
                )}
                <Button
                  variant="outline"
                  className="w-full justify-start gap-3 text-red-400 border-white/10 hover:bg-red-500/10 hover:border-red-500/30"
                  onClick={() => handleDelete(actionUser.id)}
                  disabled={currentUser?.id === actionUser.id}
                  title={currentUser?.id === actionUser.id ? "You cannot delete your own account" : ""}
                >
                  <Trash2 className="w-4 h-4" />
                  Delete user
                </Button>
              </div>
              <Button
                variant="ghost"
                className="w-full mt-4 text-slate-400 hover:text-slate-200 hover:bg-white/5"
                onClick={() => setActionUser(null)}
              >
                Close
              </Button>
            </Card>
          </div>
        </div>
      )}
    </div>
  );
}
