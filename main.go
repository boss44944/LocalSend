package main

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	qrcode "github.com/skip2/go-qrcode"
)

const maxHistory = 100

//go:embed static/*
var embedded embed.FS

// HistoryItem is one text, image, or file event.
type HistoryItem struct {
	Type string `json:"type"`
	Time time.Time `json:"time"`
	Content string `json:"content,omitempty"`
	Filename string `json:"filename,omitempty"`
	Size int64 `json:"size,omitempty"`
	URL string `json:"url,omitempty"`
	MIME string `json:"mime_type,omitempty"`
}

// App owns runtime state and HTTP handlers.
type App struct {
	port int
	dataDir, uploadDir, historyPath, lanIP, lanURL, password string
	auth bool
	maxUpload int64
	sessionKey []byte
	mu sync.RWMutex
	history []HistoryItem
	clients map[*websocket.Conn]struct{}
	writeMu sync.Mutex
	upgrader websocket.Upgrader
}

// NewApp initializes runtime folders, history, network address, and authentication.
func NewApp(port int, dataDir string, auth bool, password string, maxUpload int64) (*App, error) {
	abs, err := filepath.Abs(dataDir)
	if err != nil { return nil, err }
	uploadDir := filepath.Join(abs, "uploads")
	if err := os.MkdirAll(uploadDir, 0755); err != nil { return nil, err }
	ip, err := detectLANIP()
	if err != nil { log.Printf("warning: %v", err); ip = "127.0.0.1" }
	if auth && password == "" { password, err = randomPassword(); if err != nil { return nil, err } }
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil { return nil, err }
	a := &App{port: port, dataDir: abs, uploadDir: uploadDir, historyPath: filepath.Join(abs, "history.json"), lanIP: ip, lanURL: fmt.Sprintf("http://%s:%d", ip, port), auth: auth, password: password, maxUpload: maxUpload, sessionKey: key, clients: map[*websocket.Conn]struct{}{}}
	a.upgrader = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024, CheckOrigin: func(r *http.Request) bool { o := r.Header.Get("Origin"); return o == "" || o == "http://"+r.Host || o == "https://"+r.Host }}
	if err := a.loadHistory(); err != nil { log.Printf("warning: load history: %v", err) }
	return a, nil
}

// Handler returns the complete application router.
func (a *App) Handler() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("GET /", a.index)
	m.HandleFunc("POST /api/login", a.login)
	m.HandleFunc("GET /api/history", a.guard(a.getHistory))
	m.HandleFunc("POST /api/text", a.guard(a.postText))
	m.HandleFunc("POST /api/upload", a.guard(a.postUpload))
	m.HandleFunc("GET /api/qrcode", a.qrCode)
	m.HandleFunc("GET /ws", a.guard(a.webSocket))
	m.Handle("GET /uploads/", a.guard(http.StripPrefix("/uploads/", http.FileServer(http.Dir(a.uploadDir))).ServeHTTP))
	staticFS, _ := fs.Sub(embedded, "static")
	m.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	return security(m)
}

func (a *App) index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" { http.NotFound(w, r); return }
	b, err := embedded.ReadFile("static/index.html")
	if err != nil { http.Error(w, "index unavailable", 500); return }
	s := strings.ReplaceAll(string(b), "{{LAN_ADDRESS}}", strings.TrimPrefix(a.lanURL, "http://"))
	s = strings.ReplaceAll(s, "{{AUTH_REQUIRED}}", strconv.FormatBool(a.auth))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.WriteString(w, s)
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	if !a.auth { writeJSON(w, 200, map[string]bool{"ok": true}); return }
	var p struct{ Password string `json:"password"` }
	if err := decodeJSON(r, &p, 4096); err != nil { writeError(w, 400, err.Error()); return }
	if subtle.ConstantTimeCompare([]byte(p.Password), []byte(a.password)) != 1 { writeError(w, 401, "密码错误"); return }
	http.SetCookie(w, &http.Cookie{Name: "lan_clipboard_session", Value: a.token(), Path: "/", HttpOnly: true, SameSite: http.SameSiteStrictMode, MaxAge: 2592000})
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (a *App) postText(w http.ResponseWriter, r *http.Request) {
	var p struct{ Content string `json:"content"` }
	if err := decodeJSON(r, &p, 1<<20); err != nil { writeError(w, 400, err.Error()); return }
	p.Content = strings.TrimSpace(p.Content)
	if p.Content == "" { writeError(w, 400, "文本不能为空"); return }
	item := HistoryItem{Type: "text", Time: time.Now(), Content: p.Content}
	a.add(item)
	if err := copyClipboard(p.Content); err != nil { log.Printf("clipboard: %v", err) }
	log.Printf("%s\nTEXT\n%s", item.Time.Format("15:04:05"), item.Content)
	writeJSON(w, 201, item)
}

func (a *App) postUpload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, a.maxUpload+(1<<20))
	if err := r.ParseMultipartForm(a.maxUpload); err != nil { writeError(w, 413, "文件过大或上传格式无效"); return }
	f, h, err := r.FormFile("file")
	if err != nil { writeError(w, 400, "请选择文件"); return }
	defer f.Close()
	original := h.Filename
	stored := uniqueName(a.uploadDir, cleanName(original))
	path := filepath.Join(a.uploadDir, stored)
	out, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil { writeError(w, 500, "无法保存文件"); return }
	n, copyErr := io.Copy(out, io.LimitReader(f, a.maxUpload+1))
	closeErr := out.Close()
	if copyErr != nil || closeErr != nil || n > a.maxUpload { _ = os.Remove(path); writeError(w, 413, "文件保存失败或超过大小限制"); return }
	mt := h.Header.Get("Content-Type")
	if mt == "" || mt == "application/octet-stream" { mt = mime.TypeByExtension(filepath.Ext(stored)) }
	kind := "file"; if strings.HasPrefix(mt, "image/") { kind = "image" }
	item := HistoryItem{Type: kind, Time: time.Now(), Filename: original, Size: n, URL: "/uploads/"+stored, MIME: mt}
	a.add(item)
	log.Printf("%s\n%s\n%s", item.Time.Format("15:04:05"), strings.ToUpper(kind), original)
	writeJSON(w, 201, item)
}

func (a *App) getHistory(w http.ResponseWriter, _ *http.Request) { a.mu.RLock(); items := append([]HistoryItem(nil), a.history...); a.mu.RUnlock(); writeJSON(w, 200, items) }
func (a *App) qrCode(w http.ResponseWriter, _ *http.Request) { b, err := qrcode.Encode(a.lanURL, qrcode.Medium, 256); if err != nil { http.Error(w, "QR error", 500); return }; w.Header().Set("Content-Type", "image/png"); w.Header().Set("Cache-Control", "no-store"); _, _ = w.Write(b) }

func (a *App) webSocket(w http.ResponseWriter, r *http.Request) {
	c, err := a.upgrader.Upgrade(w, r, nil); if err != nil { return }
	a.mu.Lock(); a.clients[c] = struct{}{}; a.mu.Unlock()
	defer func(){ a.mu.Lock(); delete(a.clients, c); a.mu.Unlock(); _ = c.Close() }()
	c.SetReadLimit(1024)
	for { if _, _, err := c.ReadMessage(); err != nil { return } }
}

func (a *App) add(item HistoryItem) {
	a.mu.Lock(); a.history = append([]HistoryItem{item}, a.history...); if len(a.history) > maxHistory { a.history = a.history[:maxHistory] }; snapshot := append([]HistoryItem(nil), a.history...); clients := make([]*websocket.Conn, 0, len(a.clients)); for c := range a.clients { clients = append(clients, c) }; a.mu.Unlock()
	if err := atomicJSON(a.historyPath, snapshot); err != nil { log.Printf("save history: %v", err) }
	payload, _ := json.Marshal(item); a.writeMu.Lock(); defer a.writeMu.Unlock()
	for _, c := range clients { _ = c.SetWriteDeadline(time.Now().Add(5*time.Second)); if err := c.WriteMessage(websocket.TextMessage, payload); err != nil { a.mu.Lock(); delete(a.clients, c); a.mu.Unlock(); _ = c.Close() } }
}

func (a *App) loadHistory() error {
	b, err := os.ReadFile(a.historyPath)
	if errors.Is(err, os.ErrNotExist) { return atomicJSON(a.historyPath, []HistoryItem{}) }
	if err != nil { return err }
	if len(strings.TrimSpace(string(b))) == 0 { return nil }
	if err := json.Unmarshal(b, &a.history); err != nil { return err }
	if len(a.history) > maxHistory { a.history = a.history[:maxHistory] }
	return nil
}

func (a *App) guard(next http.HandlerFunc) http.HandlerFunc { return func(w http.ResponseWriter, r *http.Request) { if !a.auth || a.authenticated(r) { next(w, r); return }; writeError(w, 401, "需要访问密码") } }
func (a *App) authenticated(r *http.Request) bool { c, err := r.Cookie("lan_clipboard_session"); return err == nil && subtle.ConstantTimeCompare([]byte(c.Value), []byte(a.token())) == 1 }
func (a *App) token() string { h := sha256.New(); _, _ = h.Write(a.sessionKey); _, _ = h.Write([]byte(a.password)); return base64.RawURLEncoding.EncodeToString(h.Sum(nil)) }

func detectLANIP() (string, error) {
	ifs, err := net.Interfaces(); if err != nil { return "", err }
	var fallback string
	for _, in := range ifs { if in.Flags&net.FlagUp == 0 || in.Flags&net.FlagLoopback != 0 { continue }; addrs, _ := in.Addrs(); for _, addr := range addrs { ip, _, err := net.ParseCIDR(addr.String()); if err != nil || ip.To4() == nil || ip.IsLoopback() || ip.IsLinkLocalUnicast() { continue }; v := ip.To4(); s := v.String(); if v[0] == 10 || v[0] == 192 && v[1] == 168 || v[0] == 172 && v[1] >= 16 && v[1] <= 31 { return s, nil }; if fallback == "" { fallback = s } } }
	if fallback != "" { return fallback, nil }; return "", errors.New("no non-loopback IPv4 address found")
}

func copyClipboard(text string) error {
	var c *exec.Cmd
	switch runtime.GOOS { case "darwin": c = exec.Command("pbcopy"); case "windows": c = exec.Command("cmd", "/c", "clip"); case "linux": if _, err := exec.LookPath("xclip"); err == nil { c = exec.Command("xclip", "-selection", "clipboard") } else if _, err := exec.LookPath("wl-copy"); err == nil { c = exec.Command("wl-copy") } else { return errors.New("install xclip or wl-clipboard") }; default: return fmt.Errorf("unsupported OS: %s", runtime.GOOS) }
	c.Stdin = strings.NewReader(text); out, err := c.CombinedOutput(); if err != nil { return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out))) }; return nil
}

func openBrowser(url string) error { var c *exec.Cmd; switch runtime.GOOS { case "darwin": c = exec.Command("open", url); case "windows": c = exec.Command("cmd", "/c", "start", "", url); default: c = exec.Command("xdg-open", url) }; return c.Start() }
func randomPassword() (string, error) { b := make([]byte, 6); r := make([]byte, 6); if _, err := rand.Read(r); err != nil { return "", err }; for i := range b { b[i] = '0'+r[i]%10 }; return string(b), nil }
func cleanName(name string) string { name = filepath.Base(strings.ReplaceAll(name, "\\", "/")); var b strings.Builder; for _, r := range strings.TrimSpace(name) { if r < 32 || strings.ContainsRune(`<>:"/\\|?*`, r) { b.WriteByte('_') } else { b.WriteRune(r) } }; name = strings.Trim(b.String(), ". "); if name == "" { name = "upload" }; return name }
func uniqueName(dir, name string) string { ext, base := filepath.Ext(name), strings.TrimSuffix(name, filepath.Ext(name)); for i := 0; ; i++ { candidate := name; if i > 0 { candidate = fmt.Sprintf("%s-%d%s", base, i, ext) }; if _, err := os.Stat(filepath.Join(dir, candidate)); errors.Is(err, os.ErrNotExist) { return candidate } } }
func decodeJSON(r *http.Request, dst any, limit int64) error { d := json.NewDecoder(io.LimitReader(r.Body, limit)); d.DisallowUnknownFields(); if err := d.Decode(dst); err != nil { return fmt.Errorf("请求格式无效: %w", err) }; return nil }
func writeJSON(w http.ResponseWriter, status int, v any) { w.Header().Set("Content-Type", "application/json; charset=utf-8"); w.WriteHeader(status); if err := json.NewEncoder(w).Encode(v); err != nil { log.Printf("response: %v", err) } }
func writeError(w http.ResponseWriter, status int, msg string) { writeJSON(w, status, map[string]string{"error": msg}) }
func atomicJSON(path string, v any) error { b, err := json.MarshalIndent(v, "", "  "); if err != nil { return err }; tmp := path+".tmp"; if err := os.WriteFile(tmp, append(b, '\n'), 0644); err != nil { return err }; if err := os.Rename(tmp, path); err == nil { return nil }; _ = os.Remove(path); return os.Rename(tmp, path) }
func security(next http.Handler) http.Handler { return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Header().Set("X-Content-Type-Options", "nosniff"); w.Header().Set("X-Frame-Options", "DENY"); w.Header().Set("Referrer-Policy", "no-referrer"); w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data: blob:; style-src 'self'; script-src 'self'; connect-src 'self' ws: wss:; object-src 'none'; frame-ancestors 'none'"); next.ServeHTTP(w, r) }) }

func main() {
	log.SetFlags(0)
	port := flag.Int("port", 8000, "HTTP listen port")
	dataDir := flag.String("data-dir", ".", "uploads and history directory")
	auth := flag.Bool("auth", true, "require access password")
	password := flag.String("password", "", "fixed password; generated when empty")
	open := flag.Bool("open", true, "open browser on startup")
	maxMB := flag.Int64("max-upload", 512, "maximum upload size in MiB")
	flag.Parse()
	if *port < 1 || *port > 65535 || *maxMB < 1 { log.Fatal("invalid command-line options") }
	a, err := NewApp(*port, *dataDir, *auth, *password, *maxMB<<20); if err != nil { log.Fatalf("startup failed: %v", err) }
	fmt.Printf("LAN Clipboard\n\nListening:\n\n%s\n", a.lanURL)
	if a.auth { fmt.Printf("\nAccess password: %s\n\n", a.password) }
	if qr, err := qrcode.New(a.lanURL, qrcode.Medium); err == nil { fmt.Print(qr.ToSmallString(false)) }
	if *open { go func(){ time.Sleep(300*time.Millisecond); if err := openBrowser(fmt.Sprintf("http://localhost:%d", *port)); err != nil { log.Printf("open browser: %v", err) } }() }
	s := &http.Server{Addr: ":"+strconv.Itoa(*port), Handler: a.Handler(), ReadHeaderTimeout: 10*time.Second, ReadTimeout: 30*time.Second, WriteTimeout: 10*time.Minute, IdleTimeout: 2*time.Minute}
	if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) { log.Fatal(err) }
}
