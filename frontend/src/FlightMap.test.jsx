import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import FlightMap from "./FlightMap";

const sampleSnapshot = {
  hospitals: [
    { name: "Alpha", northM: 3000, eastM: 4000 },
    { name: "Beta", northM: -6000, eastM: 8000 },
    { name: "Gamma", northM: 0, eastM: 10000 },
  ],
  flights: [
    {
      launchTime: 60,
      hospitalNames: ["Alpha"],
      orderIds: ["1"],
    },
    {
      launchTime: 120,
      hospitalNames: ["Beta", "Gamma"],
      orderIds: ["2", "3"],
    },
  ],
};

describe("FlightMap", () => {
  it("returns nothing when no snapshot is provided", () => {
    const { container } = render(<FlightMap snapshot={null} />);
    expect(container).toBeEmptyDOMElement();
  });

  it("renders the heading with flight count", () => {
    render(<FlightMap snapshot={sampleSnapshot} />);
    expect(
      screen.getByRole("heading", { name: /flight map/i })
    ).toBeInTheDocument();
    expect(screen.getByText(/2 flights/i)).toBeInTheDocument();
  });

  it("renders a polyline per flight and a circle per hospital plus the Nest", () => {
    const { container } = render(<FlightMap snapshot={sampleSnapshot} />);

    const polylines = container.querySelectorAll("polyline");
    expect(polylines).toHaveLength(sampleSnapshot.flights.length);

    const circles = container.querySelectorAll("circle");
    // one per hospital + the Nest
    expect(circles.length).toBe(sampleSnapshot.hospitals.length + 1);

    for (const hospital of sampleSnapshot.hospitals) {
      expect(screen.getByText(hospital.name)).toBeInTheDocument();
    }
    expect(screen.getByText(/^Nest$/)).toBeInTheDocument();
  });

  it("renders an empty SVG (no polylines) when there are no flights", () => {
    const empty = { ...sampleSnapshot, flights: [] };
    const { container } = render(<FlightMap snapshot={empty} />);
    expect(container.querySelectorAll("polyline")).toHaveLength(0);
    expect(screen.getByText(/0 flights/i)).toBeInTheDocument();
  });
});
