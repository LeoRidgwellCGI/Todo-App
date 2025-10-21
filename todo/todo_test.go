package todo

import "testing"

// Status.Validate cases
// TestTodo_StatusValidate ensures that Status.Validate correctly
// accepts valid statuses and rejects invalid ones.
// It covers all defined statuses and an invalid example.
func TestTodo_StatusValidate(t *testing.T) {
	cases := []struct {
		name    string
		status  Status
		wantErr bool
	}{
		{"valid_not_started", StatusNotStarted, false},
		{"valid_started", StatusStarted, false},
		{"valid_completed", StatusCompleted, false},
		{"valid_mixed_case", Status("StArTeD"), false},
		{"invalid_value", Status("finished"), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.status.Validate()
			if (err != nil) != tc.wantErr {
				t.Fatalf("Validate() err=%v wantErr=%v", err, tc.wantErr)
			}
		})
	}
}

// TestTodo_Add verifies that adding items works correctly,
// including validation of description and status.
// It checks both successful additions and expected errors.
// The test covers various scenarios including valid additions and error cases.
func TestTodo_Add(t *testing.T) {
	cases := []struct {
		name    string
		desc    string
		status  Status
		wantErr bool
		wantLen int
	}{
		{"ok_not_started", "Buy milk", StatusNotStarted, false, 1},
		{"ok_started", "Ship parcel", StatusStarted, false, 1},
		{"empty_desc_error", "   ", StatusStarted, true, 0},
		{"invalid_status_error", "Learn Go", Status("nope"), true, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var list []Item
			list, it, err := Add(list, tc.desc, tc.status)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("Add() expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Add() unexpected error: %v", err)
			}
			if len(list) != tc.wantLen {
				t.Fatalf("Add() list length = %d, want %d", len(list), tc.wantLen)
			}
			if it.ID == 0 || it.Description == "" {
				t.Fatalf("Add() returned incomplete item: %+v", it)
			}
		})
	}
}

// TestTodo_UpdateStatus verifies that updating the status of to-do items works correctly,
// including handling of invalid IDs and statuses.
// It checks both successful updates and expected errors.
// The test covers various scenarios including valid updates and error cases.
func TestTodo_UpdateStatus(t *testing.T) {
	list := []Item{{ID: 1, Description: "A", Status: StatusNotStarted}, {ID: 2, Description: "B", Status: StatusNotStarted}}
	cases := []struct {
		name    string
		id      int
		status  Status
		wantErr bool
	}{
		{"ok_update", 1, StatusCompleted, false},
		{"missing_id", 99, StatusStarted, true},
		{"invalid_status", 2, Status("no"), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := UpdateStatus(list, tc.id, tc.status)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.id == 1 && out[0].Status != tc.status {
				t.Fatalf("status not updated: %+v", out[0])
			}
		})
	}
}

// TestTodo_UpdateDescription verifies that updating the description of to-do items works correctly,
// including handling of invalid IDs and empty descriptions.
// It checks both successful updates and expected errors.
// The test covers various scenarios including valid updates and error cases.
func TestTodo_UpdateDescription(t *testing.T) {
	list := []Item{{ID: 10, Description: "Old", Status: StatusNotStarted}}
	cases := []struct {
		name    string
		id      int
		newDesc string
		wantErr bool
	}{
		{"ok_update", 10, "New text", false},
		{"empty_desc", 10, "   ", true},
		{"missing_id", 99, "X", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := UpdateDescription(list, tc.id, tc.newDesc)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out[0].Description != tc.newDesc {
				t.Fatalf("description not updated: %+v", out[0])
			}
		})
	}
}

// TestTodo_Delete verifies that deleting to-do items works correctly,
// including handling of invalid IDs.
// It checks both successful deletions and expected errors.
// The test covers various scenarios including valid deletions and error cases.
func TestTodo_Delete(t *testing.T) {
	list := []Item{{ID: 1, Description: "A", Status: StatusNotStarted}, {ID: 2, Description: "B", Status: StatusStarted}, {ID: 3, Description: "C", Status: StatusCompleted}}
	cases := []struct {
		name    string
		id      int
		wantLen int
		wantErr bool
	}{
		{"ok_remove_middle", 2, 2, false},
		{"missing_id_error", 42, 3, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := Delete(list, tc.id)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(out) != tc.wantLen {
				t.Fatalf("len=%d want %d", len(out), tc.wantLen)
			}
		})
	}
}
