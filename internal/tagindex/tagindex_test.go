package tagindex_test

import (
	"sort"
	"testing"

	"github.com/example/driftwatch/internal/tagindex"
)

func sorted(s []string) []string {
	sort.Strings(s)
	return s
}

func TestNew_StartsEmpty(t *testing.T) {
	idx := tagindex.New()
	if got := idx.PathsForTag("anything"); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
	if got := idx.TagsForPath("/etc/hosts"); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestTag_AssociatesPathWithTags(t *testing.T) {
	idx := tagindex.New()
	if err := idx.Tag("/etc/nginx/nginx.conf", "nginx", "critical"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	paths := sorted(idx.PathsForTag("nginx"))
	if len(paths) != 1 || paths[0] != "/etc/nginx/nginx.conf" {
		t.Fatalf("unexpected paths: %v", paths)
	}

	tags := sorted(idx.TagsForPath("/etc/nginx/nginx.conf"))
	if len(tags) != 2 || tags[0] != "critical" || tags[1] != "nginx" {
		t.Fatalf("unexpected tags: %v", tags)
	}
}

func TestTag_MultiplePathsSameTag(t *testing.T) {
	idx := tagindex.New()
	_ = idx.Tag("/etc/nginx/nginx.conf", "web")
	_ = idx.Tag("/etc/apache2/apache2.conf", "web")

	paths := sorted(idx.PathsForTag("web"))
	if len(paths) != 2 {
		t.Fatalf("expected 2 paths, got %v", paths)
	}
}

func TestTag_DuplicateTagIsIdempotent(t *testing.T) {
	idx := tagindex.New()
	_ = idx.Tag("/etc/hosts", "base")
	_ = idx.Tag("/etc/hosts", "base")

	paths := idx.PathsForTag("base")
	if len(paths) != 1 {
		t.Fatalf("expected 1 path, got %d", len(paths))
	}
}

func TestTag_EmptyPathReturnsError(t *testing.T) {
	idx := tagindex.New()
	if err := idx.Tag("", "sometag"); err == nil {
		t.Fatal("expected error for empty path, got nil")
	}
}

func TestTag_EmptyTagReturnsError(t *testing.T) {
	idx := tagindex.New()
	if err := idx.Tag("/etc/hosts", ""); err == nil {
		t.Fatal("expected error for empty tag, got nil")
	}
}

func TestUntag_RemovesAllAssociations(t *testing.T) {
	idx := tagindex.New()
	_ = idx.Tag("/etc/hosts", "base", "critical")
	idx.Untag("/etc/hosts")

	if got := idx.PathsForTag("base"); got != nil {
		t.Fatalf("expected nil after untag, got %v", got)
	}
	if got := idx.TagsForPath("/etc/hosts"); got != nil {
		t.Fatalf("expected nil after untag, got %v", got)
	}
}

func TestUntag_EmptyTagMapCleanedUp(t *testing.T) {
	idx := tagindex.New()
	_ = idx.Tag("/etc/hosts", "solo")
	idx.Untag("/etc/hosts")
	// Adding the same tag to a new path must still work correctly.
	_ = idx.Tag("/etc/resolv.conf", "solo")
	paths := idx.PathsForTag("solo")
	if len(paths) != 1 || paths[0] != "/etc/resolv.conf" {
		t.Fatalf("unexpected paths after re-tag: %v", paths)
	}
}
