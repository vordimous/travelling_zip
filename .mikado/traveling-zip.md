# Mikado Goal: Implement The Traveling Zip plan (Step 1 + Step 2a + Step 2b + tests)

**Slug:** traveling-zip
**Started:** 2026-04-25
**Plan:** /Users/adanelz/.claude/plans/rustling-knitting-squid.md
**Build:** `make build`
> frontend: build-frontend
> backend: build-backend
**Test:** `make test`
> frontend: test-frontend
> backend: test-backend
**Commit strategy:** folded
**Base commit:** d3d128caa73ff02ee65d5e91eb964d05de464dc8

## Status

Prerequisites recorded from naive experiment; ready to start leaf loop.

## Mikado Graph

```mermaid
graph TD
  G((Goal: Traveling Zip — Step 1 + 2a + 2b + tests))
  G --> S1[Step 1: ZipScheduler over delivery graph]
  G --> S2a[Step 2a: pluggable edge-weight model + config knob]
  G --> S2b[Step 2b: pre-launch flight-path visualization]

  S1 --> G1[Build delivery graph type Nest+hospitals, Euclidean edges ✓]
  S1 --> SCHED[Scheduler logic on top of graph ✓]
  S1 --> T1I[Integration test: full orders.csv fulfilled, no constraint violations ✓]

  G1 --> T1G[Unit tests: graph nodes, edge weights, symmetry ✓]

  SCHED --> P1[Track fleet availability via zip return times ✓]
  SCHED --> P2[Order pending queue by Emergency-before-Resupply ✓]
  SCHED --> P3[Build a flight: collapse duplicate stops, cap at MaxPackages, enforce range ✓]
  SCHED --> P4[Multi-stop route ordering nearest-neighbor over graph ✓]
  SCHED --> P5[20/80 reserve policy with EoD-deadline override for Resupply ✓]
  P1 --> T1S1[Unit test: fleet capacity never exceeded ✓]
  P2 --> T1S2[Unit test: priority ordering observed ✓]
  P3 --> T1S3[Unit test: range gating, dedupe, MaxPackages cap ✓]
  P5 --> T1S4[Unit test: 20/80 reserve + deadline override ✓]

  S2a --> C1[Configurability]
  C1 --> C1a[SimulationConfig + ParseConfig: edgeWeightModel + scheduling knob ✓]
  C1 --> C1b[UI migration]
  C1b --> C1b1[Extract DataTable to own file with fixedHeight prop]
  C1b --> C1b2[Pinned header layout + sticky styles + App test heading update]
  C1b --> C1b3[Summary bar above FlightMap]
  C1b --> C1b4[Config modal opened from header button]
  C1b1 --> C1b2
  C1b2 --> C1b3
  C1b2 --> C1b4
  S2a --> C2[EdgeWeight strategy + cleanup]
  C2 --> C2a[EdgeWeight strategy interface; default=Euclidean ✓]
  C2 --> C2b[Pre-compute edges at graph construction ✓]
  C2 --> C2c[TODO comment on resupplyAtRisk round-trip overestimate ✓]
  S2a --> C3[One alternative edge-weight model selectable via config — deferred post-Step-2]
  S2a --> C4[One scheduling knob with observable behavior change ✓]
  C1a --> C2a
  C2a --> C3
  C1a --> C4
  S2a --> T2A[Unit tests: each strategy + knob changes behavior observably ✓]

  S2b --> V1[Snapshot already carries hospitals+flights; verify shape suffices ✓]
  S2b --> V2[SVG/Canvas FlightMap component: Nest, hospitals, flight legs ✓]
  S2b --> V3[Wire FlightMap into App.jsx with latest snapshot ✓]
  V1 --> V2
  V2 --> V3
  S2b --> T2B[Vitest smoke test: FlightMap renders given representative snapshot ✓]
```

## Prerequisites

### Step 1 — Foundation

- [x] **G1**: Introduce `Graph` type (Nest + hospital nodes; symmetric Euclidean edges) in a new file under `backend/core/`. Goal is a real type, not just a distance map.
- [x] **T1G**: Unit tests for graph (node set = Nest + 21 hospitals, edge weights match Euclidean, symmetric edges).

### Step 1 — Scheduler

- [x] **P1**: Track fleet availability — `zipReturnTimes` slice; `availableZips(currentTime)` reclaims returned zips. Naive proved this is required to avoid over-launching.
- [x] **P2**: Sort/partition pending orders so Emergency is considered before Resupply.
- [x] **P3**: Flight builder that (a) collapses duplicate hospitals into one stop with N packages, (b) caps stops by MaxPackagesPerZip total packages, (c) rejects routes exceeding `zipMaxCumulativeRangeM`.
- [x] **P4**: Multi-stop route ordering using nearest-neighbor traversal over the graph (improvement on FIFO).
- [x] **P5**: 20/80 fleet reserve — cap concurrent Resupply launches at 80% of fleet, but allow borrowing the reserve when a Resupply order risks missing its EoD deadline. EoD risk threshold: round-trip direct flight time > seconds remaining in day. Default policy is `ReserveSoft`. Also wires `LaunchFlights` end-to-end.
- [x] **T1S1**: Unit test for fleet capacity (availableZips reclaim, LaunchFlights respects fleet size).
- [x] **T1S2**: Unit test for priority ordering (pendingByPriority + LaunchFlights launches Emergency before earlier Resupply).
- [x] **T1S3**: Unit tests for buildFlight rules — range gating (Far rejected, Near+Mid kept), duplicate-stop collapse (3 orders, 1 stop), MaxPackages cap (third order deferred).
- [x] **T1S4**: Reserve policy unit tests — Hard blocks Resupply at the reserve boundary, None ignores it, Soft blocks early in the day but borrows the reserve when EoD round-trip exceeds remaining time. Plus a direct test of resupplyAtRisk thresholding.
- [x] **T1I**: Integration test running full `orders.csv` through the simulator; asserts 0 unfulfilled, no flight exceeds range, never more than `numZips` concurrent flights, Emergency mean delay < Resupply mean delay.

### Step 2a — Configurable routing

- [x] **C1a**: Extend `SimulationConfig` (and `ParseConfig`) with optional fields. Adds `EdgeWeightModel` (default `"euclidean"`, consumed by C3) and `EmergencyWaitThresholdSec` (default 0, consumed by C4). `ParseConfig` accepts both as optional with defaults; existing payloads continue to work unchanged.
- **C1b**: UI migration — split into ordered sub-leaves after a worktree experiment. Each sub-leaf keeps tests green on its own.
  - [ ] **C1b-1**: Extract `DataTable` to its own file with a `fixedHeight` prop. Drop-in, no visible UI change.
  - [ ] **C1b-2**: Replace the hero block with a pinned `.app-header` (title, status pill, "Edit Config" + "Run Simulation" buttons). Move `<main className="app-shell">` below the header. Add sticky positioning + baseline header styles. Update `App.test.jsx` heading regex to match the new header copy.
  - [ ] **C1b-3**: Summary bar between header and `FlightMap` showing totals (Hospitals / Orders / Flights / Unfulfilled).
  - [ ] **C1b-4**: Extracted `ConfigModal` component opened from the header button. Submit runs the simulator and closes the modal. Modal accessibility (focus trap, Escape-to-close) is deferred to post-Step-2; current scope is a backdrop + close button.
- [x] **C2a**: `EdgeWeight` strategy interface; default impl returns Euclidean (extracted from G1). `Graph.EdgeWeight` now delegates to a pluggable strategy; `NewGraph` keeps the default Euclidean behavior, and `NewGraphWithEdgeWeight` lets callers inject alternatives (used by C3).
- [x] **C2b**: Pre-compute edges at graph construction so `EdgeWeight` is an O(1) map lookup rather than a `sqrt` per call. The Euclidean strategy now builds an `edges[from][to]` matrix once.
- [x] **C2c**: Add a TODO comment on `resupplyAtRisk` documenting that `2 * EdgeWeight(Nest, X)` overestimates the actual round-trip time when the order flies as part of a multi-stop, so the at-risk threshold triggers reserve borrowing slightly earlier than strictly necessary — a conservative fail-safe. Code change deferred to post-Step-2.
- [~] **C3** (deferred): Alternative edge-weight model. Skipped for this challenge — not needed for Step 2 deliverables. The seam (C2a + EdgeWeightModel config field) is in place; `NewGraphWithEdgeWeight` carries an expansion-instruction comment that walks through how to add a model later. Revisit post-Step-2.
- [x] **C4**: `EmergencyWaitThresholdSec` is now read by `NewZipScheduler` and added as a second OR-trigger inside `resupplyAtRisk`. When set > 0, a Resupply order that has been queued for at least the threshold seconds is treated as at-risk and may borrow the reserve under `ReserveSoft`. T2A will pin the observable behavior change.
- [x] **T2A**: Three tests pin the C4 knob: threshold=0 (default) blocks Resupply mid-day under ReserveSoft; threshold=200 with 300 s wait borrows the reserve; threshold=500 with 300 s wait still blocks. C3 strategy tests are deferred along with C3.

### Step 2b — Visualization

- [x] **V1**: Confirmed by live probe of `POST /api/simulation`: snapshot keys are `[implementation, config, hospitals, orders, flights, unfulfilledOrders]`. `hospitals[i] = {name, northM, eastM}` and `flights[i] = {launchTime, hospitalNames, orderIds}`. The Nest is implicit at `(0,0)` and is the start + end of every route. Sufficient for V2's SVG/canvas rendering.
- [x] **V2**: `FlightMap` React component — SVG with auto-fit `viewBox` from hospital extents (north axis flipped to match SVG orientation), hospitals + Nest as labelled circles, flight legs drawn as low-opacity polylines (Nest → stops → Nest). Stroke and label sizes scale with the dataset span. Styles + responsive height added to `styles.css`.
- [x] **V3**: `FlightMap` mounted in `App.jsx` between Simulation Controls and the data tables, so the operator sees the planned routes immediately after a simulation run.
- [x] **T2B**: Four Vitest tests — null snapshot returns nothing, header renders with flight count, one polyline per flight + one circle per hospital plus the Nest are present, and an empty `flights[]` produces zero polylines while still rendering the heading.

## Notes and learnings

- Backend is Go 1.26, single package `backend/core` with scaffolded `ZipScheduler`. `Runner.Simulate` already wires queue + per-minute launch.
- Snapshot pipeline (`BuildSimulationSnapshot`) is in place and consumed by `cmd/api` and the React frontend; only the scheduler internals + frontend visualization are missing.
- Frontend has Vitest set up (commit 7f6e5d1).
- CI runs Go tests on PRs (`.github/workflows/ci.yml`).
- Constants (NumZips=10, MaxPackages=3, Speed=30 m/s, Range=160 km) live alongside `SimulationConfig` in `core/simulation.go`.
- **Naive observation**: pure FIFO greedy fulfills 300 orders but emergencies launch dozens of minutes after queueing. Priority + reserve are real requirements, not nice-to-haves.
- **Naive observation**: multi-order-same-hospital trips emit duplicate consecutive stops; flight builder must dedupe stops while still tracking N packages per stop.
- **Naive observation**: time conversion is `dist / speedMps` (m / (m/s) = s) — int truncation is fine for second-resolution simulation but worth noting in the scheduler comment.
- All interaction with code goes through `make` targets (per user direction); no direct `go build`/`go test` calls.
- **P4 experiment**: across all 1330 three-stop combinations of the 21 hospitals, NN gives avg route 188km vs FIFO 203km (and optimal 185km). NN matches optimal in 57% of combos and shrinks the over-160km set from 1021 to 944. NN sits inside `buildFlight`'s range gate so flights that fit only when reordered are not rejected.
- **P5 experiment**: against `orders.csv` with 10 zips, `ReserveHard` cuts emergency mean delay from 531s (no reserve) to 254s — a 52% win — at the cost of resupply mean delay rising from 2399s to 3056s. `ReserveSoft` is identical to Hard at 10 zips because the dataset's last order is at 19:56 and direct flight times don't outrun the seconds-remaining threshold until very late. Stress-tested at 4 zips: SOFT recovers 3 of the 7 extra unfulfilled orders that HARD leaves behind by letting at-risk Resupply borrow the reserve. Threshold tuning (a stale-resupply timeout) is a natural fit for the C4 configurable knob in Step 2a.
