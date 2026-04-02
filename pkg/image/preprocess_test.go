package image

import (
    "os"
    "path/filepath"
    "testing"
)

func TestProcessBatchEmpty(t *testing.T) {
    _, err := ProcessBatch([]string{}, 128, 128)
    if err == nil {
        t.Fatalf("expected error for empty batch")
    }
}

func TestListImagesAndProcess(t *testing.T) {
    dir := t.TempDir()
    path := filepath.Join(dir, "test.png")
    f, err := os.Create(path)
    if err != nil {
        t.Fatalf("create file: %v", err)
    }
    // write minimal PNG header to make decode fail gracefully in test
    f.Write([]byte{137, 80, 78, 71, 13, 10, 26, 10})
    f.Close()

    files, err := ListImages(dir)
    if err != nil {
        t.Fatalf("list images: %v", err)
    }
    if len(files) != 1 {
        t.Fatalf("expected 1 file, got %d", len(files))
    }
    _, err = ProcessBatch(files, 64, 64)
    if err == nil {
        t.Fatalf("expected decode error for invalid png")
    }
}
