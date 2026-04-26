import { z } from "zod";

// Single source of truth for simulation config: validation schema + render
// descriptors. Mirrors core.SimulationConfig in backend/core/simulation.go —
// keep field names, types, and defaults aligned with the Go struct. Form
// state holds strings (HTML inputs are strings); the schema coerces and
// validates at the form boundary.

export const EDGE_WEIGHT_MODELS = ["euclidean"];

const positiveInt = z.coerce.number().int().min(1);
const nonNegativeInt = z.coerce.number().int().min(0);

export const ConfigSchema = z.object({
  numZips: positiveInt,
  maxPackagesPerZip: positiveInt,
  zipSpeedMps: positiveInt,
  zipMaxCumulativeRangeM: positiveInt,
  edgeWeightModel: z.enum(EDGE_WEIGHT_MODELS),
  emergencyWaitThresholdSec: nonNegativeInt,
});

export const configFields = [
  { key: "numZips", label: "Number of Zips", kind: "int", min: 1, step: 1 },
  { key: "maxPackagesPerZip", label: "Max packages per Zip", kind: "int", min: 1, step: 1 },
  { key: "zipSpeedMps", label: "Zip speed (m/s)", kind: "int", min: 1, step: 1 },
  { key: "zipMaxCumulativeRangeM", label: "Max range (m)", kind: "int", min: 1, step: 1 },
  { key: "edgeWeightModel", label: "Edge weight model", kind: "enum", options: EDGE_WEIGHT_MODELS },
  { key: "emergencyWaitThresholdSec", label: "Emergency wait threshold (s)", kind: "int", min: 0, step: 1 },
];

export const defaultConfig = {
  numZips: 10,
  maxPackagesPerZip: 3,
  zipSpeedMps: 30,
  zipMaxCumulativeRangeM: 160000,
  edgeWeightModel: EDGE_WEIGHT_MODELS[0],
  emergencyWaitThresholdSec: 0,
};

export function configToInputs(config) {
  return Object.fromEntries(
    configFields.map(({ key }) => [
      key,
      String(config?.[key] ?? defaultConfig[key]),
    ])
  );
}

// Returns { data, fieldErrors } — data is the parsed numeric config when
// valid, fieldErrors is a { [key]: message } map when invalid.
export function parseConfigInputs(inputs) {
  const result = ConfigSchema.safeParse(inputs);
  if (result.success) {
    return { data: result.data, fieldErrors: null };
  }
  const fieldErrors = {};
  for (const issue of result.error.issues) {
    const key = issue.path[0];
    if (key && !fieldErrors[key]) {
      fieldErrors[key] = issue.message;
    }
  }
  return { data: null, fieldErrors };
}
