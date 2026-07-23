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

func (a *App) Handler() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("GET /", a.index)
	m.HandleFunc("POST /api/login", a.login)
	m.HandleFunc("GET /api/config", a.getConfig)
	m.HandleFunc("POST /api/config", a.updateConfig)
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
	if r.URL.Path != "/" { http.NotFound(w,r); return }
	b, err := embedded.ReadFile("static/index.html")
	if err != nil { http.Error(w,"index unavailable",500); return }
	s := strings.ReplaceAll(string(b), "{{LAN_ADDRESS}}", strings.TrimPrefix(a.lanURL,"http://"))
	s = strings.ReplaceAll(s, "{{AUTH_REQUIRED}}", strconv.FormatBool(a.auth))
	w.Header().Set("Content-Type","text/html; charset=utf-8")
	_, _ = io.WriteString(w,s)
}

func (a *App) qrCode(w http.ResponseWriter,_ *http.Request){b,e:=qrcode.Encode(a.lanURL,qrcode.Medium,256);if e!=nil{http.Error(w,"QR error",500);return};w.Header().Set("Content-Type","image/png");_,_=w.Write(b)}

func copyClipboard(text string) error { var c *exec.Cmd; switch runtime.GOOS {case "darwin": c=exec.Command("pbcopy");case "windows":c=exec.Command("cmd","/c","clip");default:return errors.New("unsupported")};c.Stdin=strings.NewReader(text);return c.Run()}

func (a *App) token() string {h:=sha256.New();h.Write(a.sessionKey);h.Write([]byte(a.password));return base64.RawURLEncoding.EncodeToString(h.Sum(nil))}
func (a *App) authenticated(r *http.Request)bool{c,e:=r.Cookie("lan_clipboard_session");return e==nil&&subtle.ConstantTimeCompare([]byte(c.Value),[]byte(a.token()))==1}
func (a *App) guard(next http.HandlerFunc)http.HandlerFunc{return func(w http.ResponseWriter,r *http.Request){if !a.auth||a.authenticated(r){next(w,r);return};writeError(w,401,"需要访问密码")}}

func writeJSON(w http.ResponseWriter,status int,v any){w.Header().Set("Content-Type","application/json");w.WriteHeader(status);json.NewEncoder(w).Encode(v)}
func writeError(w http.ResponseWriter,status int,msg string){writeJSON(w,status,map[string]string{"error":msg})}

func security(next http.Handler)http.Handler{return next}
func main(){log.SetFlags(0);cfg:=loadConfig();_ = cfg;fmt.Println("LAN Clipboard")}
