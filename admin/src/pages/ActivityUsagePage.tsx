import { useEffect, useMemo, useState } from "react";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { api } from "@/lib/api";
import { toast } from "@/components/ui/toaster";
import { formatDateShort } from "@/lib/utils";
import { useAuthStore } from "@/state/auth";

type ActivityRow = {
  date: string;
  user_id?: string;
  company_id?: string;
  action_type: string;
  count: number;
};

type TokenAggRow = {
  date: string;
  user_id?: string;
  company_id?: string;
  model: string;
  provider: string;
  prompt_tokens: number;
  completion_tokens: number;
  total_tokens: number;
};

export function ActivityUsagePage() {
  const [activity, setActivity] = useState<ActivityRow[]>([]);
  const [tokenAgg, setTokenAgg] = useState<TokenAggRow[]>([]);
  const [loading, setLoading] = useState(false);
  const [live, setLive] = useState<{
    chat: number;
    upload: number;
    rag: number;
    tokens: number;
    ts?: string;
  }>({
    chat: 0,
    upload: 0,
    rag: 0,
    tokens: 0,
  });

  const [filters, setFilters] = useState({
    company_id: "",
    user_id: "",
    action_type: "",
    model: "",
    provider: "",
    start_date: "",
    end_date: "",
  });
  const today = useMemo(() => new Date().toISOString().slice(0, 10), []);

  const dateRangeLabel = useMemo(() => {
    if (filters.start_date && filters.end_date)
      return `${filters.start_date} → ${filters.end_date}`;
    return "Last 30 days (default)";
  }, [filters.start_date, filters.end_date]);

  const summaries = useMemo(() => {
    const actionTotals: Record<string, number> = {};
    let totalActivity = 0;
    activity.forEach((a) => {
      actionTotals[a.action_type] =
        (actionTotals[a.action_type] || 0) + a.count;
      totalActivity += a.count;
    });

    let totalTokens = 0;
    tokenAgg.forEach((t) => {
      totalTokens += t.total_tokens;
    });

    return { actionTotals, totalActivity, totalTokens };
  }, [activity, tokenAgg]);

  useEffect(() => {
    loadData();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Live feed via SSE
  useEffect(() => {
    const { token } = useAuthStore.getState();
    if (!token) return;
    const source = new EventSource(
      `/api/v1/admin/live-usage?token=${encodeURIComponent(token)}`
    );

    source.onmessage = (event) => {
      try {
        const parsed = JSON.parse(event.data);
        if (parsed.error) return;
        setLive({
          chat: parsed.chat_messages ?? 0,
          upload: parsed.file_uploads ?? 0,
          rag: parsed.rag_queries ?? 0,
          tokens: parsed.total_tokens ?? 0,
          ts: parsed.refreshed_at_utc,
        });
      } catch {
        // ignore parse errors
      }
    };

    source.onerror = () => {
      source.close();
    };

    return () => {
      source.close();
    };
  }, []);

  async function loadData() {
    try {
      setLoading(true);
      const activityRes = await api.admin.activityAggregates.list({
        company_id: filters.company_id || undefined,
        user_id: filters.user_id || undefined,
        action_type: filters.action_type || undefined,
        start_date: filters.start_date || undefined,
        end_date: filters.end_date || undefined,
        page: 1,
        limit: 50,
      });
      const tokenRes = await api.admin.tokenUsageAggregates.list({
        company_id: filters.company_id || undefined,
        user_id: filters.user_id || undefined,
        model: filters.model || undefined,
        provider: filters.provider || undefined,
        start_date: filters.start_date || undefined,
        end_date: filters.end_date || undefined,
        page: 1,
        limit: 50,
      });
      setActivity(activityRes.data ?? []);
      setTokenAgg(tokenRes.data ?? []);
    } catch (error) {
      toast({
        title: "Gagal memuat data",
        description: error instanceof Error ? error.message : "Request gagal",
        variant: "destructive",
      });
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="space-y-6 animate-fade-in-up">
      <div>
        <h1 className="text-3xl font-display font-bold text-slate-100 mb-2">
          Activity & Token Usage
        </h1>
        <p className="text-slate-400">
          Pantau aktivitas pengguna dan konsumsi token per hari.
        </p>
      </div>

      <Card className="p-4 hover-card-effect">
        <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-4 gap-4">
          <Input
            placeholder="Company ID"
            value={filters.company_id}
            onChange={(e) =>
              setFilters((f) => ({ ...f, company_id: e.target.value }))
            }
            className="bg-white/5 border-white/10"
          />
          <Input
            placeholder="User ID"
            value={filters.user_id}
            onChange={(e) =>
              setFilters((f) => ({ ...f, user_id: e.target.value }))
            }
            className="bg-white/5 border-white/10"
          />
          <Input
            placeholder="Action type (e.g. activity.chat.message)"
            value={filters.action_type}
            onChange={(e) =>
              setFilters((f) => ({ ...f, action_type: e.target.value }))
            }
            className="bg-white/5 border-white/10"
          />
          <Input
            placeholder="Model (token usage)"
            value={filters.model}
            onChange={(e) =>
              setFilters((f) => ({ ...f, model: e.target.value }))
            }
            className="bg-white/5 border-white/10"
          />
          <Input
            placeholder="Provider (token usage)"
            value={filters.provider}
            onChange={(e) =>
              setFilters((f) => ({ ...f, provider: e.target.value }))
            }
            className="bg-white/5 border-white/10"
          />
          <div className="flex gap-2">
            <Input
              type="date"
              value={filters.start_date}
              max={today}
              onChange={(e) =>
                setFilters((f) => {
                  const value = e.target.value;
                  // Clamp start date to today (no future)
                  const clamped = value && value > today ? today : value || "";
                  // If end date is before new start, align it
                  const end =
                    f.end_date && f.end_date < clamped ? clamped : f.end_date;
                  return { ...f, start_date: clamped, end_date: end };
                })
              }
              className="bg-white/5 border-white/10"
            />
            <Input
              type="date"
              value={filters.end_date}
              min={filters.start_date || today}
              onChange={(e) =>
                setFilters((f) => {
                  const value = e.target.value;
                  const minDate = f.start_date || today;
                  // Clamp end date to not be before today or start_date
                  let clamped = value;
                  if (clamped && clamped < minDate) clamped = minDate;
                  if (clamped && clamped < today) clamped = today;
                  return { ...f, end_date: clamped || "" };
                })
              }
              className="bg-white/5 border-white/10"
            />
          </div>
        </div>
        <div className="flex items-center justify-between mt-4">
          <div className="text-xs text-slate-500">{dateRangeLabel}</div>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => {
                setFilters({
                  company_id: "",
                  user_id: "",
                  action_type: "",
                  model: "",
                  provider: "",
                  start_date: "",
                  end_date: "",
                });
                setTimeout(loadData, 0);
              }}
            >
              Reset
            </Button>
            <Button size="sm" onClick={loadData} disabled={loading}>
              {loading ? "Loading..." : "Apply Filters"}
            </Button>
          </div>
        </div>
      </Card>

      <div className="grid grid-cols-1 xl:grid-cols-2 gap-6">
        <Card className="p-6 hover-card-effect">
          <div className="flex items-center justify-between mb-4">
            <div>
              <h2 className="text-xl font-semibold text-slate-100">
                System Stats (Selected Range)
              </h2>
              <p className="text-sm text-slate-400">Totals from aggregates</p>
            </div>
            <span className="text-xs text-slate-500">
              {activity.length + tokenAgg.length} rows
            </span>
          </div>
          <div className="grid grid-cols-2 md:grid-cols-3 gap-3">
            <Stat label="Total Activity" value={summaries.totalActivity} />
            <Stat label="Total Tokens" value={summaries.totalTokens} />
            <div className="p-3 rounded-lg bg-white/5 border border-white/5">
              <div className="text-xs text-slate-400 mb-2">Top Actions</div>
              <div className="space-y-1">
                {Object.entries(summaries.actionTotals)
                  .sort((a, b) => b[1] - a[1])
                  .slice(0, 3)
                  .map(([action, count]) => (
                    <div
                      key={action}
                      className="flex items-center justify-between text-xs text-slate-300"
                    >
                      <span className="truncate mr-2">{action}</span>
                      <span className="text-slate-400">{count}</span>
                    </div>
                  ))}
                {Object.keys(summaries.actionTotals).length === 0 && (
                  <div className="text-xs text-slate-500">Tidak ada data</div>
                )}
              </div>
            </div>
          </div>
        </Card>

        <Card className="p-6 hover-card-effect">
          <div className="flex items-center justify-between mb-4">
            <div>
              <h2 className="text-xl font-semibold text-slate-100">
                Live Counters (Today)
              </h2>
              <p className="text-sm text-slate-400">SSE feed, auto-updating</p>
            </div>
            <span className="text-xs text-slate-500">
              {live.ts ? `Updated: ${formatDateShort(live.ts)}` : ""}
            </span>
          </div>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
            <Stat label="Chats" value={live.chat} />
            <Stat label="Uploads" value={live.upload} />
            <Stat label="RAG Queries" value={live.rag} />
            <Stat label="Tokens" value={live.tokens} />
          </div>
        </Card>

        <Card className="p-6 hover-card-effect">
          <div className="flex items-center justify-between mb-4">
            <div>
              <h2 className="text-xl font-semibold text-slate-100">
                Activity Aggregates
              </h2>
              <p className="text-sm text-slate-400">
                Grouped per day, user/company, action
              </p>
            </div>
            <span className="text-xs text-slate-500">
              {activity.length} rows
            </span>
          </div>
          {activity.length === 0 ? (
            <div className="text-slate-500 text-sm">Tidak ada data.</div>
          ) : (
            <div className="space-y-2">
              {activity.map((row, idx) => (
                <div
                  key={`${row.date}-${row.action_type}-${idx}`}
                  className="p-3 rounded-lg bg-white/5 border border-white/5"
                >
                  <div className="flex items-center justify-between">
                    <div className="text-sm text-slate-200 font-medium">
                      {row.action_type}
                    </div>
                    <div className="text-xs text-slate-500">
                      {formatDateShort(row.date)}
                    </div>
                  </div>
                  <div className="text-xs text-slate-400 mt-1 space-y-1">
                    {row.company_id && <div>Company: {row.company_id}</div>}
                    {row.user_id && <div>User: {row.user_id}</div>}
                    <div>Count: {row.count}</div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </Card>

        <Card className="p-6 hover-card-effect">
          <div className="flex items-center justify-between mb-4">
            <div>
              <h2 className="text-xl font-semibold text-slate-100">
                Token Usage Aggregates
              </h2>
              <p className="text-sm text-slate-400">
                Grouped per day, user/company, model/provider
              </p>
            </div>
            <span className="text-xs text-slate-500">
              {tokenAgg.length} rows
            </span>
          </div>
          {tokenAgg.length === 0 ? (
            <div className="text-slate-500 text-sm">Tidak ada data.</div>
          ) : (
            <div className="space-y-2">
              {tokenAgg.map((row, idx) => (
                <div
                  key={`${row.date}-${row.model}-${row.provider}-${idx}`}
                  className="p-3 rounded-lg bg-white/5 border border-white/5"
                >
                  <div className="flex items-center justify-between">
                    <div className="text-sm text-slate-200 font-medium">
                      {row.model} · {row.provider}
                    </div>
                    <div className="text-xs text-slate-500">
                      {formatDateShort(row.date)}
                    </div>
                  </div>
                  <div className="text-xs text-slate-400 mt-1 space-y-1">
                    {row.company_id && <div>Company: {row.company_id}</div>}
                    {row.user_id && <div>User: {row.user_id}</div>}
                    <div>
                      Prompt: {row.prompt_tokens} · Completion:{" "}
                      {row.completion_tokens} · Total: {row.total_tokens}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </Card>
      </div>
    </div>
  );
}

function Stat({ label, value }: { label: string; value: number }) {
  return (
    <div className="p-3 rounded-lg bg-white/5 border border-white/5">
      <div className="text-xs text-slate-400 mb-1">{label}</div>
      <div className="text-lg font-semibold text-slate-100">{value}</div>
    </div>
  );
}
