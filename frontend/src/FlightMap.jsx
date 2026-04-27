import { useMemo } from "react";

// FlightMap renders the Nest, every hospital, and every flight in the most
// recent simulation snapshot. SVG with an auto-fit viewBox; flight legs are
// drawn at low opacity so overlapping routes remain readable.
//
// Snapshot shape (from POST /api/simulation):
//   hospitals: [{ name, northM, eastM }]
//   flights:   [{ launchTime, hospitalNames, orderIds }]
// The Nest is implicit at (0, 0) and is the start + end of every flight.

const HOSPITAL_RADIUS = 4;
const NEST_RADIUS = 7;
const FLIGHT_OPACITY = 0.25;

function svgCoord(node) {
  // SVG y grows downward; north grows upward, so flip vertical axis.
  return { x: node.eastM, y: -node.northM };
}

export default function FlightMap({ snapshot }) {
  const layout = useMemo(() => {
    if (!snapshot) {
      return null;
    }
    const points = [
      svgCoord({ northM: 0, eastM: 0 }),
      ...snapshot.hospitals.map(svgCoord),
    ];
    const xs = points.map((point) => point.x);
    const ys = points.map((point) => point.y);
    const minX = Math.min(...xs);
    const maxX = Math.max(...xs);
    const minY = Math.min(...ys);
    const maxY = Math.max(...ys);
    const width = maxX - minX;
    const height = maxY - minY;
    const margin = Math.max(width, height) * 0.06 || 1;
    return {
      viewBox: `${minX - margin} ${minY - margin} ${width + 2 * margin} ${height + 2 * margin}`,
      span: Math.max(width, height) || 1,
    };
  }, [snapshot]);

  if (!snapshot || !layout) {
    return null;
  }

  const hospitalsByName = new Map(
    snapshot.hospitals.map((hospital) => [hospital.name, hospital])
  );
  // Scale stroke + radius to the dataset so they read at any zoom.
  const stroke = layout.span * 0.0015;
  const radius = layout.span * 0.006;
  const labelSize = layout.span * 0.012;

  return (
    <section className="panel">
      <div className="panel-header">
        <h2>Flight Map</h2>
        <span>{snapshot.flights.length} flights</span>
      </div>
      <svg
        className="flight-map"
        role="img"
        aria-label="Pre-execution map of planned flight paths from the Nest to hospitals"
        viewBox={layout.viewBox}
        preserveAspectRatio="xMidYMid meet"
      >
        <g className="flight-paths">
          {snapshot.flights.map((flight, index) => {
            const stops = flight.hospitalNames
              .map((name) => hospitalsByName.get(name))
              .filter(Boolean);
            if (stops.length === 0) {
              return null;
            }
            const route = [
              { northM: 0, eastM: 0 },
              ...stops,
              { northM: 0, eastM: 0 },
            ].map(svgCoord);
            const points = route.map((p) => `${p.x},${p.y}`).join(" ");
            return (
              <polyline
                key={`${flight.launchTime}-${index}`}
                points={points}
                fill="none"
                stroke="#3b82f6"
                strokeWidth={stroke}
                strokeOpacity={FLIGHT_OPACITY}
                strokeLinejoin="round"
                strokeLinecap="round"
              />
            );
          })}
        </g>
        <g className="hospitals">
          {snapshot.hospitals.map((hospital) => {
            const point = svgCoord(hospital);
            return (
              <g key={hospital.name}>
                <circle
                  cx={point.x}
                  cy={point.y}
                  r={radius || HOSPITAL_RADIUS}
                  fill="#1f2937"
                />
                <text
                  x={point.x}
                  y={point.y - radius * 1.6}
                  fontSize={labelSize}
                  textAnchor="middle"
                  fill="#1f2937"
                >
                  {hospital.name}
                </text>
              </g>
            );
          })}
        </g>
        <g className="nest">
          <circle
            cx={0}
            cy={0}
            r={(radius || NEST_RADIUS) * 1.6}
            fill="#dc2626"
          />
          <text
            x={0}
            y={-radius * 2.2}
            fontSize={labelSize * 1.1}
            textAnchor="middle"
            fontWeight="600"
            fill="#dc2626"
          >
            Nest
          </text>
        </g>
      </svg>
    </section>
  );
}
