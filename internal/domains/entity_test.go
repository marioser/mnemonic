package domains

import (
	"strings"
	"testing"
)

func TestBuildDocument(t *testing.T) {
	e := Entity{
		Type:     "project",
		Title:    "SCADA Mamonal",
		ClientID: "cli-ecopetrol",
		Industry: "automatizacion-industrial",
		Content:  "Sistema SCADA redundante para planta Mamonal",
		Tags:     []string{"scada", "oil-gas"},
	}

	doc := e.BuildDocument()

	if !strings.Contains(doc, "Project: SCADA Mamonal") {
		t.Error("document should contain capitalized type and title")
	}
	if !strings.Contains(doc, "Cliente: cli-ecopetrol") {
		t.Error("document should contain client")
	}
	if !strings.Contains(doc, "Industria: automatizacion-industrial") {
		t.Error("document should contain industry")
	}
	if !strings.Contains(doc, "Sistema SCADA redundante") {
		t.Error("document should contain content")
	}
	if !strings.Contains(doc, "Tags: scada, oil-gas") {
		t.Error("document should contain tags")
	}
}

func TestBuildDocumentMinimal(t *testing.T) {
	e := Entity{
		Type:  "lesson",
		Title: "No cotizar comisionamiento menor a 2 semanas",
	}

	doc := e.BuildDocument()
	if doc != "Lesson: No cotizar comisionamiento menor a 2 semanas" {
		t.Errorf("unexpected document: %s", doc)
	}
}

func TestBuildSummary(t *testing.T) {
	e := Entity{
		Summary: "Resumen corto",
		Content: "Contenido largo que debería ignorarse cuando hay summary",
	}

	s := e.BuildSummary(300)
	if s != "Resumen corto" {
		t.Errorf("expected summary, got: %s", s)
	}
}

func TestBuildSummaryFallbackToContent(t *testing.T) {
	e := Entity{
		Content: "Contenido que se usa como fallback",
	}

	s := e.BuildSummary(300)
	if s != "Contenido que se usa como fallback" {
		t.Errorf("expected content as fallback, got: %s", s)
	}
}

func TestBuildSummaryTruncation(t *testing.T) {
	e := Entity{
		Summary: strings.Repeat("x", 500),
	}

	s := e.BuildSummary(300)
	if len(s) != 300 {
		t.Errorf("expected truncated to 300, got %d", len(s))
	}
}

func TestBuildSummaryFallbackToTitle(t *testing.T) {
	e := Entity{
		Title: "My Title",
	}

	s := e.BuildSummary(300)
	if s != "My Title" {
		t.Errorf("expected title fallback, got: %s", s)
	}
}

func TestToMetadata(t *testing.T) {
	e := Entity{
		Type:       "proposal",
		Domain:     "commercial",
		Title:      "Propuesta SCADA",
		Summary:    "Resumen de la propuesta",
		ClientID:   "cli-ecopetrol",
		Industry:   "automatizacion-industrial",
		Status:     "draft",
		Source:     "manual",
		RelatedIDs: []string{"proj-001", "opp-042"},
		Tags:       []string{"scada", "ecopetrol"},
	}

	meta, err := e.ToMetadata(300)
	if err != nil {
		t.Fatalf("ToMetadata failed: %v", err)
	}
	if meta == nil {
		t.Fatal("metadata should not be nil")
	}

	// Verify key fields
	typ, ok := meta.GetString("type")
	if !ok || typ != "proposal" {
		t.Errorf("expected type=proposal, got %s (ok: %v)", typ, ok)
	}

	domain, ok := meta.GetString("domain")
	if !ok || domain != "commercial" {
		t.Errorf("expected domain=commercial, got %s", domain)
	}

	clientID, ok := meta.GetString("client_id")
	if !ok || clientID != "cli-ecopetrol" {
		t.Errorf("expected client_id=cli-ecopetrol, got %s", clientID)
	}

	relIDs, ok := meta.GetString("related_ids")
	if !ok || relIDs != "proj-001,opp-042" {
		t.Errorf("expected related_ids=proj-001,opp-042, got %s", relIDs)
	}

	tags, ok := meta.GetString("tags")
	if !ok || tags != "scada,ecopetrol" {
		t.Errorf("expected tags=scada,ecopetrol, got %s", tags)
	}

	// created_at and updated_at should be set automatically
	createdAt, ok := meta.GetString("created_at")
	if !ok || createdAt == "" {
		t.Error("created_at should be set automatically")
	}
}

func TestCapitalize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"project", "Project"},
		{"", ""},
		{"A", "A"},
		{"already", "Already"},
	}

	for _, tt := range tests {
		got := capitalize(tt.input)
		if got != tt.expected {
			t.Errorf("capitalize(%s) = %s, want %s", tt.input, got, tt.expected)
		}
	}
}
