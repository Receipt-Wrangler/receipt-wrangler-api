package commands

import (
	"encoding/json"
	"testing"
)

// The desktop serializes the filter straight out of NGXS state, so the wire
// keys are the lowercase ones on SystemTaskPagedRequestFilter. A capitalized
// tag would still unmarshal (Go is case-insensitive) but would serialize a key
// the client cannot read -- the hazard documented for the receipt filter.
func TestGetSystemTaskCommandUnmarshalsFilter(t *testing.T) {
	body := `{
		"page": 1,
		"pageSize": 50,
		"orderBy": "started_at",
		"sortDirection": "desc",
		"filter": {
			"type": { "operation": "CONTAINS", "value": ["QUICK_SCAN"] },
			"ranBy": { "operation": "CONTAINS", "value": [-1, 4] },
			"startedAt": { "operation": "BETWEEN", "value": ["2026-03-10T00:00:00Z", "2026-03-12T00:00:00Z"] },
			"endedAt": { "operation": "LESS_THAN", "value": "2026-03-12T00:00:00Z" }
		}
	}`

	command := GetSystemTaskCommand{}
	if err := json.Unmarshal([]byte(body), &command); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if command.Filter.Type.Operation != CONTAINS {
		t.Errorf("expected type operation CONTAINS, got %v", command.Filter.Type.Operation)
	}

	types, ok := command.Filter.Type.Value.([]interface{})
	if !ok || len(types) != 1 || types[0] != "QUICK_SCAN" {
		t.Errorf("expected the type value to survive as a one-element slice, got %v", command.Filter.Type.Value)
	}

	// Ids decode as float64, which is what the repository's numeric coercion
	// and the "System" sentinel comparison have to handle.
	ranBy, ok := command.Filter.RanBy.Value.([]interface{})
	if !ok || len(ranBy) != 2 || ranBy[0] != float64(-1) || ranBy[1] != float64(4) {
		t.Errorf("expected the ran by ids to decode as float64, got %v", command.Filter.RanBy.Value)
	}

	startedAt, ok := command.Filter.StartedAt.Value.([]interface{})
	if !ok || len(startedAt) != 2 {
		t.Fatalf("expected a two-bound startedAt range, got %v", command.Filter.StartedAt.Value)
	}

	if command.Filter.EndedAt.Operation != LESS_THAN || command.Filter.EndedAt.Value != "2026-03-12T00:00:00Z" {
		t.Errorf("expected the endedAt scalar to survive, got %v %v",
			command.Filter.EndedAt.Operation, command.Filter.EndedAt.Value)
	}
}

// A body with no filter key must leave a zero-value filter, which the query
// builder treats as "no predicates" -- the two embedded task tables never send
// one.
func TestGetSystemTaskCommandWithoutFilterIsZeroValue(t *testing.T) {
	command := GetSystemTaskCommand{}
	if err := json.Unmarshal([]byte(`{"page": 1, "pageSize": 50}`), &command); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if command.Filter.Type.Value != nil || command.Filter.RanBy.Value != nil ||
		command.Filter.StartedAt.Value != nil || command.Filter.EndedAt.Value != nil {
		t.Errorf("expected every filter field to be nil, got %+v", command.Filter)
	}
}

func TestSystemTaskPagedRequestFilterMarshalsLowercaseKeys(t *testing.T) {
	bytes, err := json.Marshal(SystemTaskPagedRequestFilter{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	marshalled := map[string]interface{}{}
	if err := json.Unmarshal(bytes, &marshalled); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	for _, key := range []string{"type", "ranBy", "startedAt", "endedAt"} {
		if _, ok := marshalled[key]; !ok {
			t.Errorf("expected key %q, got %v", key, marshalled)
		}
	}
}
