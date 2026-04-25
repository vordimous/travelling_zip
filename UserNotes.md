# User Notes

## Overview

Briefly describe what you changed and what you chose to focus on.

## How To Run

List the exact commands needed to run your solution.

Include:

- which backend language you chose
- any required versions or environment assumptions
- any setup steps beyond the defaults in the repo

## Design Decisions

Describe the most important decisions you made while working on the exercise.

Useful things to cover:

- product or UX choices
- architectural or code organization choices
- scheduling approach
- frontend/backend interaction
- tradeoffs between simplicity, correctness, extensibility, and time

## Prioritization

What did you choose not to do?

If you had another 2-4 hours, what would you improve next?

## AI Usage

Describe how you used AI tools, if at all.

Useful things to cover:

- what AI helped with
- what you validated yourself
- where you changed or rejected AI-generated suggestions

## Known Issues / Limitations

List any known bugs, incomplete areas, or rough edges in your solution.

## Feedback on the Test

Share any feedback on the exercise itself.

Useful things to cover:

- what felt clear or unclear
- what parts of the starter code or prompt were most helpful
- what you would change to improve the exercise

## Ideation

The core of the solution is a graph that models the delivery network: the Nest and every hospital are nodes, fully connected by edges weighted with the straight-line distance between them. The scheduler reasons over this graph to plan multi-stop flights that respect range, payload, and fleet constraints.

**Step 1 — ZipScheduler (foundation)**: build the delivery graph from `hospitals.csv`, then implement a scheduler that prioritizes Emergency over Resupply, plans multi-stop routes that always start and end at the Nest, enforces the cumulative range limit, and tracks fleet availability so in-flight zips are reused when they return. A 20/80 split reserves capacity for Emergencies while letting Resupply deliveries borrow from the reserve when end-of-day deadlines are at risk.

**Step 2 — Going deeper**: three threads built on top of the Step 1 graph.

- **Smarter routing via richer edge weights** — the graph is the seam. Step 1 uses pure Euclidean distance; Step 2 layers additional considerations into the edge weight calculation so the same scheduler can produce more realistic routes without being rewritten.
- **Configurability** — the scheduling knobs (e.g. emergency wait threshold, resupply batching, edge-weight model) are exposed through the API config so behaviors can be tuned and compared.
- **Visualization** — a pre-execution preview in the frontend that renders the Nest, hospitals, and planned flight paths so the operator can see what the algorithm decided before zips launch.

**Beyond Step 2**: anything that would push complexity past a clean, demonstrable solution is captured below in the Prioritization section — advanced TSP solvers, terrain-aware edge weights, weather-aware dynamic edge weights, richer visualization, and expanded test coverage all live there.

### Scratchpad

- Location map
- dijkstra's algorithm weights for shortest path
- order delivery map from nest to each hospital
- additional delivery legs between each hospital
- delivery legs that are greater than the max range are identified and ignored
- vector store for relationships
- distance data
- terrain avoidance, static, periodic changes for cities
- Weather avoidance, dynamic, active monitoring from weather stations, Storm avoidance, wind direction optimization
