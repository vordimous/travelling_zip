import { useEffect, useState } from "react";
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

export default function App() {
  const [config, setConfig] = useState(defaultConfigInputs);
  const [snapshot, setSnapshot] = useState(null);
  const [status, setStatus] = useState("idle");
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;

    setStatus("loading");
    setError("");

    fetchSimulation(defaultConfig)
      .then((nextSnapshot) => {
        if (!active) {
          return;
        }
        setSnapshot(nextSnapshot);
        setConfig(
          Object.fromEntries(
            Object.entries(nextSnapshot.config).map(([key, value]) => [
              key,
              String(value),
            ])
          )
        );
        setStatus("success");
      })
      .catch((nextError) => {
        if (!active) {
          return;
        }
        setStatus("error");
        setError(nextError.message);
      });

    return () => {
      active = false;
    };
  }, []);

  const handleSubmit = async (event) => {
    event.preventDefault();
    setStatus("loading");
    setError("");

    try {
      const numericConfig = Object.fromEntries(
        Object.entries(config).map(([key, value]) => [key, Number(value)])
      );
      const nextSnapshot = await fetchSimulation(numericConfig);
      setSnapshot(nextSnapshot);
      setConfig(
        Object.fromEntries(
          Object.entries(nextSnapshot.config).map(([key, value]) => [
            key,
            String(value),
          ])
        )
      );
      setStatus("success");
    } catch (nextError) {
      setStatus("error");
      setError(nextError.message);
    }
  };

  const handleChange = (event) => {
    const { name, value } = event.target;
    setConfig((current) => ({
      ...current,
      [name]: value,
    }));
  };

  return (
    <main className="app-shell">
      <section className="hero panel">
        <div>
          <p className="eyebrow">Starter Frontend</p>
          <h1>Traveling Zip Simulator</h1>
          <p className="lede">
            A small React surface that talks to whichever backend
            implementation is currently running on port 3001.
          </p>
        </div>
        <div className="hero-meta">
          <span className={`status-pill status-${status}`}>{status}</span>
          {snapshot ? (
            <span className="implementation">{snapshot.implementation}</span>
          ) : null}
        </div>
      </section>

      <section className="panel">
        <div className="panel-header">
          <h2>Simulation Controls</h2>
          <span>POST /api/simulation</span>
        </div>
        <form className="control-grid" onSubmit={handleSubmit}>
          {Object.entries(config).map(([key, value]) => (
            <label key={key}>
              <span>{key}</span>
              <input
                min="0"
                name={key}
                onChange={handleChange}
                type="number"
                value={value}
              />
            </label>
          ))}
          <button type="submit">Run Simulation</button>
        </form>
        {error ? <p className="error">{error}</p> : null}
      </section>

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
            Start one backend on port 3001, then run the frontend and submit the
            form.
          </p>
        </section>
      )}
    </main>
  );
}
