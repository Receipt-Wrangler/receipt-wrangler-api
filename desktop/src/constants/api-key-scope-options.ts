import { ApiKeyScope } from "../open-api";

export const API_KEY_SCOPE_LABELS: Record<ApiKeyScope, string> = {
  r: "Read",
  w: "Write",
  rw: "Read/Write",
};

export const API_KEY_SCOPE_OPTIONS: { value: ApiKeyScope; label: string }[] = [
  ApiKeyScope.R,
  ApiKeyScope.W,
  ApiKeyScope.Rw,
].map((value) => ({ value, label: API_KEY_SCOPE_LABELS[value] }));
