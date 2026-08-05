package database

import "testing"

// TestMockLookupByStringID covers the shape every route hits: a record created
// with a numeric id, then fetched with the id taken from a path parameter,
// which arrives as a string. Comparing the interface values directly never
// matched, so get/update/delete by id silently returned nothing.
func TestMockLookupByStringID(t *testing.T) {
	db := NewMockDatabase()
	table := db.Table("documents")

	created := table.Create(map[string]interface{}{"title": "ponytail"})
	if created["id"] == nil {
		t.Fatal("Create did not assign an id")
	}

	got := table.Get("1")
	if got == nil {
		t.Fatal("Get(\"1\") returned nil for the record created with id 1")
	}
	if record, ok := got.(map[string]interface{}); !ok || record["title"] != "ponytail" {
		t.Fatalf("Get returned the wrong record: %#v", got)
	}

	if updated := table.Update("1", map[string]interface{}{"title": "updated"}); updated == nil {
		t.Fatal("Update(\"1\") did not find the record")
	}
	if record, ok := table.Get(int64(1)).(map[string]interface{}); !ok || record["title"] != "updated" {
		t.Fatalf("Get(int64) did not see the update: %#v", table.Get(int64(1)))
	}

	if !table.Delete("1") {
		t.Fatal("Delete(\"1\") did not find the record")
	}
	if table.Get("1") != nil {
		t.Fatal("record still present after Delete")
	}
}

// TestSameIDRejectsNil guards the nil case: a record without an id must not
// match a nil lookup just because both format identically.
func TestSameIDRejectsNil(t *testing.T) {
	if sameID(nil, nil) {
		t.Error("sameID(nil, nil) must be false")
	}
	if sameID(int64(1), nil) {
		t.Error("sameID(1, nil) must be false")
	}
	if !sameID(int64(7), "7") {
		t.Error("sameID(int64(7), \"7\") must be true")
	}
}
