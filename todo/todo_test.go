package todo

import "testing"

// TestStatusValidate exercises the Status.Validate method across
// valid and invalid inputs, including case-insensitive values.
func TestStatusValidate(t *testing.T) {
	tests := []struct {
		name    string
		status  Status
		wantErr bool
	}{
		{"valid_not_started", StatusNotStarted, false},
		{"valid_started", StatusStarted, false},
		{"valid_completed", StatusCompleted, false},
		{"valid_mixed_case", Status("StArTeD"), false}, // case-insensitive
		{"invalid_value", Status("finished"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.status.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() err=%v wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

// TestAdd validates business rules for Add, including:
// - successful creation with correct fields
// - rejection of empty descriptions
func TestAdd(t *testing.T) {
	var list []Item

	// OK add
	it, err := Add(&list, "Buy milk", StatusNotStarted)
	if err != nil {
		t.Fatalf("Add() unexpected error: %v", err)
	}
	if it.ID == 0 || it.Description != "Buy milk" || it.Status != StatusNotStarted {
		t.Fatalf("Add() returned unexpected item: %+v", it)
	}
	if len(list) != 1 {
		t.Fatalf("Add() list length = %d, want 1", len(list))
	}

	// Empty description -> error
	if _, err := Add(&list, "   ", StatusStarted); err == nil {
		t.Fatalf("Add() expected error for empty description")
	}
}

// TestUpdateDescription ensures:
// - description updates correctly by ID
// - missing IDs and empty new descriptions error out
func TestUpdateDescription(t *testing.T) {
	list := []Item{
		{ID: 1, Description: "A", Status: StatusNotStarted},
		{ID: 2, Description: "B", Status: StatusStarted},
	}
	updated, err := UpdateDescription(list, 2, "Bravo")
	if err != nil {
		t.Fatalf("UpdateDescription() unexpected error: %v", err)
	}
	if updated[1].Description != "Bravo" {
		t.Fatalf("UpdateDescription() description=%q want=%q", updated[1].Description, "Bravo")
	}

	if _, err := UpdateDescription(list, 999, "X"); err == nil {
		t.Fatalf("UpdateDescription() expected error for missing id")
	}
	if _, err := UpdateDescription(list, 1, "   "); err == nil {
		t.Fatalf("UpdateDescription() expected error for empty new description")
	}
}

// TestDelete verifies item removal by ID and the error path for missing IDs.
func TestDelete(t *testing.T) {
	list := []Item{
		{ID: 1, Description: "A", Status: StatusNotStarted},
		{ID: 2, Description: "B", Status: StatusStarted},
		{ID: 3, Description: "C", Status: StatusCompleted},
	}
	out, err := Delete(list, 2)
	if err != nil {
		t.Fatalf("Delete() unexpected error: %v", err)
	}
	if len(out) != 2 || out[0].ID != 1 || out[1].ID != 3 {
		t.Fatalf("Delete() result unexpected: %+v", out)
	}
	if _, err := Delete(list, 42); err == nil {
		t.Fatalf("Delete() expected error for missing id")
	}
}
