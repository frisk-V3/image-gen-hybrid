package image

import (
    "errors"
    "image"
    "image/png"
    _ "image/jpeg"
    _ "image/png"
    "os"
    "path/filepath"
    "sync"

    "github.com/nfnt/resize"
)

type ImageMeta struct {
    Path      string `json:"path"`
    Width     int    `json:"width"`
    Height    int    `json:"height"`
    ThumbPath string `json:"thumb_path"`
}

func ListImages(dir string) ([]string, error) {
    var out []string
    err := filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
        if err != nil { return err }
        if info.IsDir() { return nil }
        ext := filepath.Ext(p)
        switch ext {
        case ".jpg", ".jpeg", ".png", ".webp":
            out = append(out, p)
        }
        return nil
    })
    return out, err
}

func ProcessBatch(paths []string, w, h uint) ([]ImageMeta, error) {
    if len(paths) == 0 { return nil, errors.New("empty batch") }
    out := make([]ImageMeta, len(paths))
    var wg sync.WaitGroup
    var mu sync.Mutex
    var firstErr error

    for i, p := range paths {
        wg.Add(1)
        go func(i int, p string) {
            defer wg.Done()
            f, err := os.Open(p)
            if err != nil {
                mu.Lock(); if firstErr==nil { firstErr = err }; mu.Unlock()
                return
            }
            defer f.Close()
            img, _, err := image.Decode(f)
            if err != nil {
                mu.Lock(); if firstErr==nil { firstErr = err }; mu.Unlock()
                return
            }
            resized := resize.Thumbnail(w, h, img, resize.Lanczos3)
            tmp := filepath.Join(os.TempDir(), filepath.Base(p)+".thumb.png")
            outf, err := os.Create(tmp)
            if err != nil {
                mu.Lock(); if firstErr==nil { firstErr = err }; mu.Unlock()
                return
            }
            defer outf.Close()
            if err := png.Encode(outf, resized); err != nil {
                mu.Lock(); if firstErr==nil { firstErr = err }; mu.Unlock()
                return
            }
            b := resized.Bounds()
            out[i] = ImageMeta{Path: p, Width: b.Dx(), Height: b.Dy(), ThumbPath: tmp}
        }(i, p)
    }
    wg.Wait()
    if firstErr != nil { return nil, firstErr }
    return out, nil
}
