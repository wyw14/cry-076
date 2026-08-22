package platform

import (
	"context"
	"strings"
	"testing"
)

func TestLocalFileStoreRejectsTraversalAndHashesContent(t *testing.T) {
	store := LocalFileStore{Root: t.TempDir()}
	if _, _, _, err := store.Put(context.Background(), "../secret", strings.NewReader("x")); err == nil {
		t.Fatal("expected traversal rejection")
	}
	path, hash, size, err := store.Put(context.Background(), "exports/result.json", strings.NewReader("hello"))
	if err != nil {
		t.Fatal(err)
	}
	if path != "exports/result.json" || hash != "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824" || size != 5 {
		t.Fatalf("path=%s hash=%s size=%d", path, hash, size)
	}
}
