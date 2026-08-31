package catalog

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestBuildTreeNestsChildrenAndNeverEmitsNull(t *testing.T) {
	root := uuid.New()
	child := uuid.New()
	cats := []Category{
		{ID: root, Slug: "development", Name: "Development"},
		{ID: child, ParentID: &root, Slug: "backend", Name: "Backend"},
		{ID: uuid.New(), Slug: "design", Name: "Design"},
	}

	tree := buildTree(cats)
	if len(tree) != 2 {
		t.Fatalf("want 2 roots, got %d", len(tree))
	}

	var dev *categoryNode
	for i := range tree {
		if tree[i].Slug == "development" {
			dev = &tree[i]
		}
	}
	if dev == nil || len(dev.Children) != 1 || dev.Children[0].Slug != "backend" {
		t.Fatalf("development should have one child 'backend', got %+v", dev)
	}

	// Leaf nodes must serialise children as [] not null.
	b, _ := json.Marshal(tree)
	if string(b) == "" || containsNull(b) {
		t.Errorf("tree JSON must not contain null children: %s", b)
	}
}

func containsNull(b []byte) bool {
	return json.Valid(b) && (indexOf(string(b), `"children":null`) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
