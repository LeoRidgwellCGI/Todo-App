package todo

import "testing"

func TestStatusValidate(t *testing.T) {
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

func TestAdd(t *testing.T) {
	cases := []struct {
		name, desc string
		status     Status
		wantErr    bool
		wantLen    int
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
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(list) != tc.wantLen {
				t.Fatalf("len=%d want %d", len(list), tc.wantLen)
			}
			if it.ID == 0 || it.Description == "" {
				t.Fatalf("bad item: %+v", it)
			}
		})
	}
}

func TestUpdateStatus(t *testing.T) {
	list := []Item{{ID: 1, Description: "A", Status: StatusNotStarted}, {ID: 2, Description: "B", Status: StatusNotStarted}}
	cases := []struct {
		name    string
		id      int
		s       Status
		wantErr bool
	}{
		{"ok_update", 1, StatusCompleted, false},
		{"missing_id", 99, StatusStarted, true},
		{"invalid_status", 2, Status("no"), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := UpdateStatus(list, tc.id, tc.s)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.id == 1 && out[0].Status != tc.s {
				t.Fatalf("not updated: %+v", out[0])
			}
		})
	}
}

func TestUpdateDescription(t *testing.T) {
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
				t.Fatalf("not updated: %+v", out[0])
			}
		})
	}
}

func TestDelete(t *testing.T) {
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
