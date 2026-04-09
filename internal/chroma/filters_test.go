package chroma

import (
	"testing"
)

func TestFilterBuilderEmpty(t *testing.T) {
	fb := NewFilter()
	if fb.HasFilters() {
		t.Error("empty builder should not have filters")
	}
	if fb.Build() != nil {
		t.Error("empty builder should return nil clause")
	}
}

func TestFilterBuilderSingle(t *testing.T) {
	fb := NewFilter().Type("project")
	if !fb.HasFilters() {
		t.Error("should have filters")
	}
	clause := fb.Build()
	if clause == nil {
		t.Error("single filter should return a clause")
	}
}

func TestFilterBuilderMultiple(t *testing.T) {
	fb := NewFilter().
		Type("proposal").
		Client("cli-ecopetrol").
		Status("active").
		Industry("automatizacion-industrial")

	if !fb.HasFilters() {
		t.Error("should have filters")
	}
	clause := fb.Build()
	if clause == nil {
		t.Error("multiple filters should return an AND clause")
	}
}

func TestFilterBuilderSkipsEmpty(t *testing.T) {
	fb := NewFilter().
		Type("").      // should be skipped
		Client("").    // should be skipped
		Status("active")

	clause := fb.Build()
	if clause == nil {
		t.Error("should have one filter (status)")
	}
}

func TestFilterBuilderAllEmpty(t *testing.T) {
	fb := NewFilter().
		Type("").
		Client("").
		Status("").
		Industry("")

	if fb.HasFilters() {
		t.Error("all empty values should result in no filters")
	}
	if fb.Build() != nil {
		t.Error("all empty should return nil")
	}
}

func TestFilterBuilderRelatedTo(t *testing.T) {
	fb := NewFilter().RelatedTo("proj-001")
	if !fb.HasFilters() {
		t.Error("should have filter for related_ids")
	}
}

func TestFilterBuilderEq(t *testing.T) {
	fb := NewFilter().Eq("custom_field", "custom_value")
	if !fb.HasFilters() {
		t.Error("should have filter for custom field")
	}
}

func TestFilterBuilderEqSkipsEmpty(t *testing.T) {
	fb := NewFilter().Eq("", "value").Eq("key", "")
	if fb.HasFilters() {
		t.Error("empty key or value should be skipped")
	}
}
