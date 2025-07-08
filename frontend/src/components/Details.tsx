import React from "react";
import { useLocation, Link } from "react-router-dom";
import { Doughnut } from "react-chartjs-2";
import { Chart as ChartJS, ArcElement, Tooltip, Legend } from "chart.js";

ChartJS.register(ArcElement, Tooltip, Legend);

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
  broken_links?: string[];
  h1_count: number;
  h2_count: number;
  h3_count: number;
  h4_count: number;
  h5_count: number;
  h6_count: number;
}

const Details: React.FC = () => {
  const location = useLocation();
  const data = location.state?.data as AnalysisResult;

  if (!data) {
    return (
      <div className="text-center p-8">
        <p>No analysis data found.</p>
        <Link
          to="/"
          className="text-blue-600 hover:underline mt-4 inline-block">
          Back to Dashboard
        </Link>
      </div>
    );
  }

  const chartData = {
    labels: ["Internal Links", "External Links"],
    datasets: [
      {
        label: "# of Links",
        data: [data.internal_links, data.external_links],
        backgroundColor: ["rgba(54, 162, 235, 0.6)", "rgba(255, 99, 132, 0.6)"],
        borderColor: ["rgba(54, 162, 235, 1)", "rgba(255, 99, 132, 1)"],
        borderWidth: 1,
      },
    ],
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold break-all">{data.title}</h1>
        <a
          href={data.url}
          target="_blank"
          rel="noopener noreferrer"
          className="text-blue-500 hover:underline break-all">
          {data.url}
        </a>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Wykres */}
        <div className="p-4 border rounded-lg shadow-sm">
          <h2 className="text-xl font-semibold mb-4">Links Distribution</h2>
          <div style={{ maxWidth: "400px", margin: "0 auto" }}>
            <Doughnut data={chartData} />
          </div>
        </div>

        <div className="p-4 border rounded-lg shadow-sm md:col-span-2">
          <h2 className="text-xl font-semibold mb-4">Page Details</h2>
          <div className="grid grid-cols-2 md:grid-cols-3 gap-6">
            <div className="space-y-2">
              <div>
                <p className="text-sm font-medium text-gray-500">
                  HTML Version
                </p>
                <p className="text-lg font-semibold">{data.html_version}</p>
              </div>
              <div>
                <p className="text-sm font-medium text-gray-500">
                  Has Login Form
                </p>
                <p className="text-lg font-semibold">
                  {data.has_login_form ? "Yes" : "No"}
                </p>
              </div>
            </div>
            <div className="space-y-2 col-span-2 md:col-span-1">
              <p className="text-sm font-medium text-gray-500">
                Headings Count
              </p>
              <div className="grid grid-cols-3 gap-x-4 gap-y-2">
                <p>
                  H1: <span className="font-bold">{data.h1_count}</span>
                </p>
                <p>
                  H2: <span className="font-bold">{data.h2_count}</span>
                </p>
                <p>
                  H3: <span className="font-bold">{data.h3_count}</span>
                </p>
                <p>
                  H4: <span className="font-bold">{data.h4_count}</span>
                </p>
                <p>
                  H5: <span className="font-bold">{data.h5_count}</span>
                </p>
                <p>
                  H6: <span className="font-bold">{data.h6_count}</span>
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <Link to="/" className="text-blue-600 hover:underline mt-6 inline-block">
        &larr; Back to Dashboard
      </Link>
    </div>
  );
};

export default Details;
