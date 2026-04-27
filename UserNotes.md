# User Notes

## Overview

> Briefly describe what you changed and what you chose to focus on.

I have a history with the Traveling Salesman problem. I worked on mars rover path planning as my Senior Thesis in undergrad. Being in the logistics industry meant a lot of route optimization work. I very likely went overboard on this challenge, but I was having fun and couldn't help myself in places.

I set out to build what essentially is a network latency map but with hospitals and euclidean coord plane distances. This meant a Dikstra algorithm could easily find paths between nodes and consider routes that start and end at the nest.

If you want to dig in more, you can find my thoughts and raw notes in the [Ideation](#ideation) section below.

## How To Run

> List the exact commands needed to run your solution.

My solution has been setup as a stand along project at [https://github.com/vordimous/travelling_zip/tree/mikado/traveling-zip-plan]() where the feature branch `mikado/traveling-zip-plan` shows my changes and `main` is the code at the point it was in your challenge zip. I kept everything runnable using go and npm but make a dedicated make file for clarity. I added CI, because CI is just great.

You can find run instructions in the [Quickstart](https://github.com/vordimous/travelling_zip/tree/mikado/traveling-zip-plan#quick-start) of the readme.

Include:

- which backend language you chose: Go
- any required versions or environment assumptions. [Prerequisites](https://github.com/vordimous/travelling_zip/tree/mikado/traveling-zip-plan#prerequisitest)
- any setup steps beyond the defaults in the repo: No

## Design Decisions

> Describe the most important decisions you made while working on the exercise.

I decided a graph representation of the hospitals would be the easiest way to reason about how the Zips could be routed. I wanted to render the graph to make it more clear when certain parameters affected a Zips ability to reach hospitals. The existing lists of orders and fulfillment was enough to describe the success criteria when changing parameter, but the UI make it hard to see at a glance. I recovered some screen real estate to make it so you don't need to scroll as much.

I wanted to do a lot with the FE/BE interaction but settled on a simple Zod implementation. I typically perfer heavy API normalization with OpenAPI schemas but didn't want to go down that road here. Zod schemas have native support for OpenAPI so it would be a good first step in that direction. There are also lots of options for Live data VS REST but ultimately would be overkill given the dataset.

I had implemented portions of a 80/20 rule for Zip fleet capacity to keep some zips in reserve to only be used for Emergency flights. I removed it in the end given there wasn't any "live" emergency flights being added and the code was extra noise for this challenge.

Useful things to cover:

- product or UX choices
- architectural or code organization choices
- scheduling approach
- frontend/backend interaction
- tradeoffs between simplicity, correctness, extensibility, and time

## Prioritization

> What did you choose not to do?

I had implemented portions of a 80/20 rule for Zip fleet capacity to keep some zips in reserve to only be used for Emergency flights. I removed it in the end given there wasn't any "live" emergency flights being added and the code was extra noise for this challenge.

I also didn't add more repo specific stuff. Linters and extra things that I would add in a typical project were left out other than the core bits that would make adding them very easy in the future.

> If you had another 2-4 hours, what would you improve next?

The edge weight options allows for generating multiple different graphs to make informed decisions. The next thing that would make this challenge pretty interesting would be a 3rd dataset of timed obstructions like storms or Zip speed deltas due head and tail wind. That data could be timed and set to coord locations or wind compass direction. That would be enough to generate multiple other edge weight calculations and graphs. I could then see where Resupply flights could take on more risk to fly into wind or light storms, but Emergency flights would be more conservative.

## AI Usage

> Describe how you used AI tools, if at all.

I used the [Mikado skill](https://github.com/vordimous/mikado-skills) that I developed with Claude Code. The manner in which my plan was implemented was directly controlled by this skill. However Claude didn't come up with any of the ideas or solutions to the problem, it was only used in the implementation and validation phases. I was monitoring and tweaking implementations after each mikado leaf-node was committed. You can browse the commit history to see it played out.

Useful things to cover:

- what AI helped with
- what you validated yourself
- where you changed or rejected AI-generated suggestions

## Known Issues / Limitations

> List any known bugs, incomplete areas, or rough edges in your solution.

No explicit bug that I have found. I think the Flight builder algorithm works well for the size of data in this challenge, however it needs to be optimized for larger datasets if there were +1k hospitals or +100k orders. There is also a lot of room for UI polish and I have only one edge weight option making the selector feel like it is missing something.

## Feedback on the Test

> Share any feedback on the exercise itself.

Love it, probably too much. You do a lot of the same things I do with my take home assessment. I spent ~5hr on this so it feels like a good size.

Useful things to cover:

- what felt clear or unclear
- what parts of the starter code or prompt were most helpful
- what you would change to improve the exercise

## Ideation

The core of the solution is a graph that models the delivery network: the Nest and every hospital are nodes, fully connected by edges weighted with the straight-line distance between them. The scheduler reasons over this graph to plan multi-stop flights that respect range, payload, and fleet constraints.

**Step 1 — ZipScheduler (foundation)**: build the delivery graph from `hospitals.csv`, then implement a scheduler that prioritizes Emergency over Resupply, plans multi-stop routes that always start and end at the Nest, enforces the cumulative range limit, and tracks fleet availability so in-flight zips are reused when they return.

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
- multiple edges to identify all possible routes between two places
- in the real world euclidean coords would be replaces with full geo spacial route planning
- Dev cleanup
  - C2: build edges at the same time as the graph so each edge is pre calculated
  - C2: Round trip calc is wrong and should be using the routeDistance function, 2x route is an overestimation, this should be fixed or commented that it isn't optimal and errors on the side of simple vs accurate. would always be larger than reality so it will still have the same effect to document potential risk.
  - C1: Migrate the UI to something a bit better to look at. Title section should be smaller and be a normal header nav section pinned to the top. config settings should be updated in a modal. edit config and run simulator can be buttons in the header. There should be one summary seaching at the top of the page with each data table section's totals in it. the data table sections should be implemented using a generic component extracted from existing code. the new component should have a fixed height option.