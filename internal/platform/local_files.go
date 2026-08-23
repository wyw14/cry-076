package platform

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type LocalFileStore struct{ Root string }

func (s LocalFileStore) safePath(name string) (string, error) {
	clean := filepath.Clean(strings.ReplaceAll(name, "\\", "/"))
	if clean == "." || filepath.IsAbs(clean) || strings.HasPrefix(clean, "..") || strings.Contains(clean, "../") {
		return "", fmt.Errorf("unsafe local path")
	}
	target := filepath.Join(s.Root, filepath.FromSlash(clean))
	relative, err := filepath.Rel(s.Root, target)
	if err != nil || strings.HasPrefix(relative, "..") {
		return "", fmt.Errorf("unsafe local path")
	}
	return target, nil
}
func (s LocalFileStore) Put(ctx context.Context, name string, body io.Reader) (string, string, int64, error) {
	target, err := s.safePath(name)
	if err != nil {
		return "", "", 0, err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o750); err != nil {
		return "", "", 0, err
	}
	tmp, err := os.CreateTemp(filepath.Dir(target), ".upload-*")
	if err != nil {
		return "", "", 0, err
	}
	defer os.Remove(tmp.Name())
	hasher := sha256.New()
	written, copyErr := io.Copy(io.MultiWriter(tmp, hasher), &contextReader{ctx: ctx, reader: body})
	closeErr := tmp.Close()
	if copyErr != nil {
		return "", "", 0, copyErr
	}
	if closeErr != nil {
		return "", "", 0, closeErr
	}
	if err := os.Rename(tmp.Name(), target); err != nil {
		return "", "", 0, err
	}
	return filepath.ToSlash(name), hex.EncodeToString(hasher.Sum(nil)), written, nil
}
func (s LocalFileStore) Open(_ context.Context, name string) (io.ReadCloser, error) {
	target, err := s.safePath(name)
	if err != nil {
		return nil, err
	}
	return os.Open(target)
}
func (s LocalFileStore) Delete(_ context.Context, name string) error {
	target, err := s.safePath(name)
	if err != nil {
		return err
	}
	if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r *contextReader) Read(buffer []byte) (int, error) {
	select {
	case <-r.ctx.Done():
		return 0, r.ctx.Err()
	default:
		return r.reader.Read(buffer)
	}
}
