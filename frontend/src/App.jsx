import { useEffect, useState } from "react";
import ConfigModal from "./ConfigModal";
import DataTable from "./DataTable";
import FlightMap from "./FlightMap";

const defaultConfig = {
  numZips: 10,
  maxPackagesPerZip: 3,
  zipSpeedMps: 30,
  zipMaxCumulativeRangeM: 160000,
};

const defaultConfigInputs = Object.fromEntries(
  Object.entries(defaultConfig).map(([key, value]) => [key, String(value)])
);

async function fetchSimulation(config) {
  const response = await fetch("/api/simulation", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ config }),
  });

  if (!response.ok) {
    const payload = await response.json().catch(() => null);
    throw new Error(payload?.error ?? "Backend request failed");
  }

  return response.json();
}

function configInputsToNumeric(inputs) {
  return Object.fromEntries(
    Object.entries(inputs).map(([key, value]) => [key, Number(value)])
  );
}

function configToInputs(config) {
  return Object.fromEntries(
    Object.entries(config).map(([key, value]) => [key, String(value)])
  );
}

function SummaryStat({ label, value }) {
  return (
    <div className="summary-stat">
      <span className="summary-label">{label}</span>
      <span className="summary-value">{value}</span>
    </div>
  );
}

export default function App() {
  const [config, setConfig] = useState(defaultConfigInputs);
  const [snapshot, setSnapshot] = useState(null);
  const [status, setStatus] = useState("idle");
  const [error, setError] = useState("");
  const [modalOpen, setModalOpen] = useState(false);

  const runSimulation = async (numericConfig) => {
    setStatus("loading");
    setError("");
    try {
      const nextSnapshot = await fetchSimulation(numericConfig);
      setSnapshot(nextSnapshot);
      setConfig(configToInputs(nextSnapshot.config));
      setStatus("success");
    } catch (nextError) {
      setStatus("error");
      setError(nextError.message);
    }
  };

  useEffect(() => {
    let active = true;
    setStatus("loading");
    setError("");
    fetchSimulation(defaultConfig)
      .then((nextSnapshot) => {
        if (!active) return;
        setSnapshot(nextSnapshot);
        setConfig(configToInputs(nextSnapshot.config));
        setStatus("success");
      })
      .catch((nextError) => {
        if (!active) return;
        setStatus("error");
        setError(nextError.message);
      });
    return () => {
      active = false;
    };
  }, []);

  const handleEditConfig = () => setModalOpen(true);

  const handleRunSimulation = () => {
    runSimulation(configInputsToNumeric(config));
  };

  const handleConfigChange = (event) => {
    const { name, value } = event.target;
    setConfig((current) => ({ ...current, [name]: value }));
  };

  const handleConfigSubmit = async (event) => {
    event.preventDefault();
    await runSimulation(configInputsToNumeric(config));
    setModalOpen(false);
  };

  return (
    <>
      <header className="app-header">
        <div className="app-header-title">
          <h1>Traveling Zip</h1>
          <span className={`status-pill status-${status}`}>{status}</span>
          {snapshot ? (
            <span className="implementation">{snapshot.implementation}</span>
          ) : null}
        </div>
        <div className="app-header-actions">
          <button type="button" onClick={handleEditConfig}>
            Edit Config
          </button>
          <button type="button" onClick={handleRunSimulation}>
            Run Simulation
          </button>
        </div>
      </header>

      <main className="app-shell">
        {error ? (
          <section className="panel">
            <p className="error">{error}</p>
          </section>
        ) : null}

        {snapshot ? (
          <section className="panel summary-bar">
            <SummaryStat label="Hospitals" value={snapshot.hospitals.length} />
            <SummaryStat label="Orders" value={snapshot.orders.length} />
            <SummaryStat label="Flights" value={snapshot.flights.length} />
            <SummaryStat
              label="Unfulfilled"
              value={snapshot.unfulfilledOrders.length}
            />
          </section>
        ) : null}

        {snapshot ? <FlightMap snapshot={snapshot} /> : null}

        {snapshot ? (
          <div className="grid">
            <DataTable
              columns={[
                { key: "name", label: "Hospital" },
                { key: "northM", label: "North (m)" },
                { key: "eastM", label: "East (m)" },
              ]}
              emptyMessage="No hospitals loaded."
              rows={snapshot.hospitals}
              title="Hospitals"
            />
            <DataTable
              columns={[
                { key: "id", label: "ID" },
                { key: "time", label: "Time" },
                { key: "hospitalName", label: "Hospital" },
                { key: "priority", label: "Priority" },
              ]}
              emptyMessage="No orders loaded."
              rows={snapshot.orders}
              title="Orders"
            />
            <DataTable
              columns={[
                { key: "launchTime", label: "Launch Time" },
                { key: "hospitalNames", label: "Hospital Names" },
                { key: "orderIds", label: "Order IDs" },
              ]}
              emptyMessage="No flights launched yet."
              rows={snapshot.flights.map((flight) => ({
                ...flight,
                hospitalNames: flight.hospitalNames.join(" -> "),
                orderIds: flight.orderIds.join(", "),
              }))}
              title="Flights"
            />
            <DataTable
              columns={[
                { key: "id", label: "ID" },
                { key: "time", label: "Time" },
                { key: "hospitalName", label: "Hospital" },
                { key: "priority", label: "Priority" },
              ]}
              emptyMessage="No unfulfilled orders."
              rows={snapshot.unfulfilledOrders}
              title="Unfulfilled Orders"
            />
          </div>
        ) : (
          <section className="panel">
            <p className="empty">
              Start one backend on port 3001, then click Run Simulation.
            </p>
          </section>
        )}
      </main>

      <ConfigModal
        open={modalOpen}
        config={config}
        onChange={handleConfigChange}
        onSubmit={handleConfigSubmit}
        onClose={() => setModalOpen(false)}
        error={error}
      />
    </>
  );
}
