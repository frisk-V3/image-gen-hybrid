package main

import (
    "bufio"
    "bytes"
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "mime/multipart"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "time"

    "github.com/yourrepo/image-gen-gui/pkg/image"
)

type GenRequest struct {
    Op    string        `json:"op"`
    Batch []image.ImageMeta `json:"batch"`
}

type GenResponse struct {
    Status  string   `json:"status"`
    Outputs []string `json:"outputs"`
}

func main() {
    addr := flag.String("addr", "127.0.0.1:8080", "server address")
    static := flag.String("static", "./webui", "static webui dir")
    flag.Parse()

    http.Handle("/", http.FileServer(http.Dir(*static)))
    http.HandleFunc("/upload", uploadHandler)
    http.HandleFunc("/events", sseHandler)
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("ok"))
    })

    log.Printf("starting server at %s", *addr)
    log.Fatal(http.ListenAndServe(*addr, nil))
}

// simple in-memory event broadcaster for SSE
var clients = make(map[chan string]bool)

func sseHandler(w http.ResponseWriter, r *http.Request) {
    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "streaming unsupported", http.StatusInternalServerError)
        return
    }
    ch := make(chan string)
    clients[ch] = true
    defer func() {
        delete(clients, ch)
        close(ch)
    }()
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    for {
        select {
        case msg := <-ch:
            fmt.Fprintf(w, "data: %s\n\n", msg)
            flusher.Flush()
        case <-r.Context().Done():
            return
        }
    }
}

func broadcast(msg string) {
    for ch := range clients {
        select {
        case ch <- msg:
        default:
        }
    }
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "method", http.StatusMethodNotAllowed)
        return
    }
    // parse multipart
    err := r.ParseMultipartForm(100 << 20) // 100MB
    if err != nil {
        http.Error(w, "parse form: "+err.Error(), http.StatusBadRequest)
        return
    }
    files := r.MultipartForm.File["files"]
    if len(files) == 0 {
        http.Error(w, "no files", http.StatusBadRequest)
        return
    }
    tmpDir, err := os.MkdirTemp("", "img-upload-*")
    if err != nil {
        http.Error(w, "tmpdir: "+err.Error(), http.StatusInternalServerError)
        return
    }
    defer os.RemoveAll(tmpDir)

    var paths []string
    for _, fh := range files {
        src, err := fh.Open()
        if err != nil {
            http.Error(w, "open file: "+err.Error(), http.StatusInternalServerError)
            return
        }
        dstPath := filepath.Join(tmpDir, filepath.Base(fh.Filename))
        dst, err := os.Create(dstPath)
        if err != nil {
            src.Close()
            http.Error(w, "create tmp: "+err.Error(), http.StatusInternalServerError)
            return
        }
        _, err = ioCopy(dst, src)
        src.Close()
        dst.Close()
        if err != nil {
            http.Error(w, "save tmp: "+err.Error(), http.StatusInternalServerError)
            return
        }
        paths = append(paths, dstPath)
    }

    // preprocess in parallel
    batchMeta, err := image.ProcessBatch(paths, 512, 512)
    if err != nil {
        http.Error(w, "process batch: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // stream progress to clients
    go func(meta []image.ImageMeta) {
        b, _ := json.Marshal(map[string]interface{}{"event": "preprocess_done", "count": len(meta)})
        broadcast(string(b))
        // call python generator subprocess
        req := GenRequest{Op: "generate", Batch: meta}
        reqb, _ := json.Marshal(req)
        // assume python/generator/app.py is executable with python3
        cmd := exec.Command("python3", "python/generator/app.py")
        stdin, _ := cmd.StdinPipe()
        stdout, _ := cmd.StdoutPipe()
        if err := cmd.Start(); err != nil {
            broadcast(string(mustMarshal(map[string]interface{}{"event":"error","msg":err.Error()})))
            return
        }
        // send request
        stdin.Write(reqb)
        stdin.Write([]byte("\n"))
        stdin.Close()
        // read response (single line JSON)
        scanner := bufio.NewScanner(stdout)
        for scanner.Scan() {
            line := scanner.Text()
            // forward to clients
            broadcast(string(mustMarshal(map[string]interface{}{"event":"generate_output","payload": json.RawMessage(line)})))
        }
        cmd.Wait()
        broadcast(string(mustMarshal(map[string]interface{}{"event":"done"})))
    }(batchMeta)

    // immediate response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{"status": "accepted", "count": len(batchMeta)})
}

// small helpers
func mustMarshal(v interface{}) []byte {
    b, _ := json.Marshal(v)
    return b
}

func ioCopy(dst *os.File, src multipart.File) (int64, error) {
    buf := make([]byte, 32*1024)
    var total int64
    for {
        n, err := src.Read(buf)
        if n > 0 {
            w, ew := dst.Write(buf[:n])
            total += int64(w)
            if ew != nil {
                return total, ew
            }
        }
        if err != nil {
            if err == os.ErrClosed || err.Error() == "EOF" {
                return total, nil
            }
            return total, err
        }
    }
}
