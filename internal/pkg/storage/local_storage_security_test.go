package storage

import (
	"context"
	"strings"
	"testing"
)

func TestLocalStoragePathValidation(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{name: "valid path", path: "test/file.txt", wantErr: false},
		{name: "path traversal attempt", path: "../secret.txt", wantErr: true},
		{name: "absolute path", path: "/etc/passwd", wantErr: true},
		{name: "nested traversal", path: "sub/../../secret.txt", wantErr: true},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, err := NewLocalStorage(&LocalStorageConfig{
				BasePath: t.TempDir(),
				BaseURL:  "http://example.com",
			})
			if err != nil {
				t.Fatalf("NewLocalStorage() error = %v", err)
			}

			_, err = store.Upload(ctx, tt.path, strings.NewReader("test"), "text/plain")
			if (err != nil) != tt.wantErr {
				t.Fatalf("Upload() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLocalStorageReadOperationsRejectUnsafePath(t *testing.T) {
	ctx := context.Background()
	storeIface, err := NewLocalStorage(&LocalStorageConfig{
		BasePath: t.TempDir(),
		BaseURL:  "http://example.com",
	})
	if err != nil {
		t.Fatalf("NewLocalStorage() error = %v", err)
	}
	store := storeIface.(*LocalStorage)

	unsafe := "../secret.txt"
	if _, err := store.Download(ctx, unsafe); err == nil {
		t.Fatal("Download() expected error for unsafe path")
	}
	if err := store.Delete(ctx, unsafe); err == nil {
		t.Fatal("Delete() expected error for unsafe path")
	}
	if _, err := store.Exists(ctx, unsafe); err == nil {
		t.Fatal("Exists() expected error for unsafe path")
	}
	if _, err := store.GetURL(ctx, unsafe, 0); err == nil {
		t.Fatal("GetURL() expected error for unsafe path")
	}
	if _, err := store.GetObjectInfo(ctx, unsafe); err == nil {
		t.Fatal("GetObjectInfo() expected error for unsafe path")
	}
}
