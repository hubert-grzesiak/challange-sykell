import React, { useState, useMemo, useEffect, useCallback } from "react";
import axios from "axios";
import { BrowserRouter as Router, Routes, Route, Link } from "react-router-dom";
import {
  QueryClient,
  QueryClientProvider,
  useQuery,
  useMutation,
} from "@tanstack/react-query";
import { type ColumnDef } from "@tanstack/react-table";
import { ArrowUpDown, Trash2, Play, Pause, RefreshCw } from "lucide-react";

import { Button } from "./components/ui/button";
import { Input } from "./components/ui/input";
import { DataTable } from "./components/Table";
import Details from "./components/Details";
import Status from "./components/Status";

const queryClient = new QueryClient();

interface AnalysisResult {
  id: number;
  url: string;
  html_version: string;
  title: string;
  status: "queued" | "running" | "done" | "error" | "stopped";
  internal_links: number;
  external_links: number;
  inaccessible_links: number;
  has_login_form: boolean;
  h1_count: number;
  h2_count: number;
  h3_count: number;
  h4_count: number;
  h5_count: number;
  h6_count: number;
}

export const columns: ColumnDef<AnalysisResult>[] = [
  {
    accessorKey: "status",
    header: "Status",
    cell: ({ row }) => <Status status={row.getValue("status")} />,
  },
  {
    id: "url",
    accessorKey: "url",
    header: ({ column }) => (
      <Button
        variant="ghost"
        onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>
        URL
        <ArrowUpDown className="ml-2 h-4 w-4" />
      </Button>
    ),
    cell: ({ row }) => {
      const url = row.getValue("url") as string;
      return (
        <Link
          to={`/details/${encodeURIComponent(url)}`}
          state={{ data: row.original }}
          className="font-medium text-blue-600 hover:underline">
          {url}
        </Link>
      );
    },
  },
  {
    accessorKey: "title",
    header: "Title",
  },
  {
    accessorKey: "h1_count",
    header: "H1",
  },
  {
    accessorKey: "h2_count",
    header: "H2",
  },
  {
    accessorKey: "internal_links",
    header: "Internal Links",
  },
  {
    accessorKey: "external_links",
    header: "External Links",
  },
  {
    accessorKey: "inaccessible_links",
    header: "Broken Links",
  },
  {
    accessorKey: "has_login_form",
    header: "Login Form",
    cell: ({ row }) => (row.getValue("has_login_form") ? "Yes" : "No"),
  },
];

function AppContent() {
  const [url, setUrl] = useState("");
  const [selectedRows, setSelectedRows] = useState<AnalysisResult[]>([]);
  const [globalFilter, setGlobalFilter] = useState("");
  const [apiKey, setApiKey] = useState(
    () => localStorage.getItem("apiKey") || ""
  );
  const [isPolling, setIsPolling] = useState(false);

  useEffect(() => {
    if (apiKey) {
      localStorage.setItem("apiKey", apiKey);
    } else {
      localStorage.removeItem("apiKey");
    }
  }, [apiKey]);

  const apiClient = useMemo(() => {
    const client = axios.create({ baseURL: "/api" });
    client.interceptors.request.use(
      (config) => {
        if (apiKey) {
          config.headers.Authorization = `Bearer ${apiKey}`;
        }
        return config;
      },
      (error) => Promise.reject(error)
    );
    return client;
  }, [apiKey]);

  const {
    data: results = [],
    isLoading,
    error,
  } = useQuery({
    queryKey: ["analyses", apiKey],
    queryFn: async () => {
      const { data } = await apiClient.get<AnalysisResult[]>("/analyses");
      return data;
    },
    enabled: !!apiKey,
    refetchInterval: isPolling ? 5000 : false,
    refetchOnWindowFocus: true,
  });

  const analyzeMutation = useMutation({
    mutationFn: (newUrl: string) => apiClient.post("/analyze", { url: newUrl }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["analyses"] });
      setUrl("");
    },
  });

  const rerunMutation = useMutation({
    mutationFn: (ids: number[]) =>
      Promise.all(
        ids.map((id) => apiClient.post("/api/analyze/rerun", { id }))
      ),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["analyses"] });
      setSelectedRows([]);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (ids: number[]) =>
      Promise.all(ids.map((id) => apiClient.delete(`/analyses/${id}`))),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["analyses"] });
      setSelectedRows([]);
    },
  });

  const handleSelectedRowsChange = useCallback((rows: AnalysisResult[]) => {
    setSelectedRows(rows);
  }, []);

  const handleRerunAnalysis = () => {
    const idsToRerun = selectedRows.map((row) => row.id);
    rerunMutation.mutate(idsToRerun);
  };

  const handleDelete = () => {
    const idsToDelete = selectedRows.map((row) => row.id);
    deleteMutation.mutate(idsToDelete);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!url) return;
    analyzeMutation.mutate(url);
  };

  return (
    <>
      <header className="bg-slate-900 text-white shadow-md">
        <div className="container mx-auto flex h-16 items-center justify-between px-4">
          <h1 className="text-xl font-bold">Website Analyzer</h1>
          <Input
            type="password"
            value={apiKey}
            onChange={(e) => setApiKey(e.target.value)}
            placeholder="Enter API Key"
            className="w-64 bg-slate-800 text-white border-slate-600"
          />
        </div>
      </header>

      <div className="container mx-auto px-4 py-8">
        <main>
          <Routes>
            <Route
              path="/"
              element={
                <div className="space-y-4">
                  <form onSubmit={handleSubmit} className="flex gap-2">
                    <Input
                      type="text"
                      value={url}
                      onChange={(e) => setUrl(e.target.value)}
                      placeholder="Enter a website URL"
                      className="flex-grow"
                    />
                    <Button
                      type="submit"
                      disabled={!apiKey || analyzeMutation.isPending}>
                      {analyzeMutation.isPending ? (
                        <div className="w-5 h-5 border-2 border-t-transparent rounded-full animate-spin" />
                      ) : (
                        "Analyze"
                      )}
                    </Button>
                  </form>

                  <div className="flex gap-2 flex-wrap items-center">
                    <Button
                      onClick={() => setIsPolling(!isPolling)}
                      variant="outline"
                      disabled={!apiKey}>
                      {isPolling ? (
                        <Pause className="mr-2 h-4 w-4" />
                      ) : (
                        <Play className="mr-2 h-4 w-4" />
                      )}
                      {isPolling ? "Stop Polling" : "Start Polling"}
                    </Button>
                    <Button
                      variant="outline"
                      disabled={
                        selectedRows.length === 0 || rerunMutation.isPending
                      }
                      onClick={handleRerunAnalysis}>
                      <RefreshCw className="mr-2 h-4 w-4" /> Re-run
                    </Button>
                    <Button
                      variant="destructive"
                      disabled={
                        selectedRows.length === 0 || deleteMutation.isPending
                      }
                      onClick={handleDelete}>
                      <Trash2 className="mr-2 h-4 w-4" /> Delete
                    </Button>
                  </div>

                  {isLoading && <p>Loading initial data...</p>}
                  {error && (
                    <p className="text-red-500">
                      Error: {(error as Error).message}
                    </p>
                  )}

                  <DataTable
                    columns={columns}
                    data={results}
                    onSelectedRowsChange={handleSelectedRowsChange}
                    setGlobalFilter={setGlobalFilter}
                    globalFilter={globalFilter}
                  />
                </div>
              }
            />
            <Route path="/details/:url" element={<Details />} />
          </Routes>
        </main>
      </div>
    </>
  );
}

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <Router>
        <AppContent />
      </Router>
    </QueryClientProvider>
  );
}

export default App;
