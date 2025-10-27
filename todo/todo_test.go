package todo

import "testing"

// Status.Validate cases
// TestTodo_StatusValidate ensures that Status.Validate correctly
// accepts valid statuses and rejects invalid ones.
// It covers all defined statuses and an invalid example.
func TestTodo_StatusValidate_valid_not_started(t *testing.T) {
	tc := struct {
		name    string
		status  Status
		wantErr bool
	}{"valid_not_started", StatusNotStarted, false}
	err := tc.status.Validate()
	if (err != nil) != tc.wantErr {
		t.Fatalf("Validate() err=%v wantErr=%v", err, tc.wantErr)
	}
}

func TestTodo_StatusValidate_valid_started(t *testing.T) {
	tc := struct {
		name    string
		status  Status
		wantErr bool
	}{"valid_started", StatusStarted, false}
	err := tc.status.Validate()
	if (err != nil) != tc.wantErr {
		t.Fatalf("Validate() err=%v wantErr=%v", err, tc.wantErr)
	}
}

func TestTodo_StatusValidate_valid_completed(t *testing.T) {
	tc := struct {
		name    string
		status  Status
		wantErr bool
	}{"valid_completed", StatusCompleted, false}
	err := tc.status.Validate()
	if (err != nil) != tc.wantErr {
		t.Fatalf("Validate() err=%v wantErr=%v", err, tc.wantErr)
	}
}

func TestTodo_StatusValidate_valid_mixed_case(t *testing.T) {
	tc := struct {
		name    string
		status  Status
		wantErr bool
	}{"valid_mixed_case", Status("StArTeD"), false}
	err := tc.status.Validate()
	if (err != nil) != tc.wantErr {
		t.Fatalf("Validate() err=%v wantErr=%v", err, tc.wantErr)
	}
}

func TestTodo_StatusValidate_invalid_value(t *testing.T) {
	tc := struct {
		name    string
		status  Status
		wantErr bool
	}{"invalid_value", Status("finished"), true}
	err := tc.status.Validate()
	if (err != nil) != tc.wantErr {
		t.Fatalf("Validate() err=%v wantErr=%v", err, tc.wantErr)
	}
}

// TestTodo_Add verifies that adding items works correctly,
// including validation of description and status.
// It checks both successful additions and expected errors.
// The test covers various scenarios including valid additions and error cases.
func TestTodo_Add_ok_not_started(t *testing.T) {
	tc := struct {
		name    string
		desc    string
		status  Status
		wantErr bool
		wantLen int
	}{"ok_not_started", "Buy milk", StatusNotStarted, false, 1}
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
}

func TestTodo_Add_ok_started(t *testing.T) {
	tc := struct {
		name    string
		desc    string
		status  Status
		wantErr bool
		wantLen int
	}{"ok_started", "Ship parcel", StatusStarted, false, 1}
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
}

func TestTodo_Add_empty_desc_error(t *testing.T) {
	tc := struct {
		name    string
		desc    string
		status  Status
		wantErr bool
		wantLen int
	}{"empty_desc_error", "   ", StatusStarted, true, 0}
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
}

func TestTodo_Add_invalid_status_error(t *testing.T) {
	tc := struct {
		name    string
		desc    string
		status  Status
		wantErr bool
		wantLen int
	}{"invalid_status_error", "Learn Go", Status("nope"), true, 0}
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
}

// TestTodo_UpdateStatus verifies that updating the status of to-do items works correctly,
// including handling of invalid IDs and statuses.
// It checks both successful updates and expected errors.
// The test covers various scenarios including valid updates and error cases.
func TestTodo_UpdateStatus_ok_update(t *testing.T) {
	list := []Item{{ID: 1, Description: "A", Status: StatusNotStarted}, {ID: 2, Description: "B", Status: StatusNotStarted}}
	tc := struct {
		name    string
		id      int
		status  Status
		wantErr bool
	}{"ok_update", 1, StatusCompleted, false}
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
}

func TestTodo_UpdateStatus_missing_id(t *testing.T) {
	list := []Item{{ID: 1, Description: "A", Status: StatusNotStarted}, {ID: 2, Description: "B", Status: StatusNotStarted}}
	tc := struct {
		name    string
		id      int
		status  Status
		wantErr bool
	}{"missing_id", 99, StatusStarted, true}
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
}

func TestTodo_UpdateStatus_invalid_status(t *testing.T) {
	list := []Item{{ID: 1, Description: "A", Status: StatusNotStarted}, {ID: 2, Description: "B", Status: StatusNotStarted}}
	tc := struct {
		name    string
		id      int
		status  Status
		wantErr bool
	}{"invalid_status", 2, Status("no"), true}
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
}

// TestTodo_UpdateDescription verifies that updating the description of to-do items works correctly,
// including handling of invalid IDs and empty descriptions.
// It checks both successful updates and expected errors.
// The test covers various scenarios including valid updates and error cases.
func TestTodo_UpdateDescription_ok_update(t *testing.T) {
	list := []Item{{ID: 10, Description: "Old", Status: StatusNotStarted}}
	tc := struct {
		name    string
		id      int
		newDesc string
		wantErr bool
	}{"ok_update", 10, "New text", false}
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
}

func TestTodo_UpdateDescription_empty_desc(t *testing.T) {
	list := []Item{{ID: 10, Description: "Old", Status: StatusNotStarted}}
	tc := struct {
		name    string
		id      int
		newDesc string
		wantErr bool
	}{"empty_desc", 10, "   ", true}
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
}

func TestTodo_UpdateDescription_missing_id(t *testing.T) {
	list := []Item{{ID: 10, Description: "Old", Status: StatusNotStarted}}
	tc := struct {
		name    string
		id      int
		newDesc string
		wantErr bool
	}{"missing_id", 99, "X", true}
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
}

// TestTodo_Delete verifies that deleting to-do items works correctly,
// including handling of invalid IDs.
// It checks both successful deletions and expected errors.
// The test covers various scenarios including valid deletions and error cases.
func TestTodo_Delete_ok_remove_middle(t *testing.T) {
	list := []Item{{ID: 1, Description: "A", Status: StatusNotStarted}, {ID: 2, Description: "B", Status: StatusStarted}, {ID: 3, Description: "C", Status: StatusCompleted}}
	tc := struct {
		name    string
		id      int
		wantLen int
		wantErr bool
	}{"ok_remove_middle", 2, 2, false}
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
}

func TestTodo_Delete_missing_id_error(t *testing.T) {
	list := []Item{{ID: 1, Description: "A", Status: StatusNotStarted}, {ID: 2, Description: "B", Status: StatusStarted}, {ID: 3, Description: "C", Status: StatusCompleted}}
	tc := struct {
		name    string
		id      int
		wantLen int
		wantErr bool
	}{"missing_id_error", 42, 3, true}
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
}
