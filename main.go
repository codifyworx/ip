package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/oschwald/geoip2-golang"
)

type config struct {
	addr          string
	basePath      string
	trustedCIDRs  []*net.IPNet
	geoCityPath   string
	geoASNPath    string
	readTimeout   time.Duration
	writeTimeout  time.Duration
	idleTimeout   time.Duration
}

type app struct {
	cfg    config
	cityDB *geoip2.Reader
	asnDB  *geoip2.Reader
}

type ipResponse struct {
	IP           string            `json:"ip"`
	Version      string            `json:"version"`
	ASN          *uint             `json:"asn,omitempty"`
	ISP          string            `json:"isp,omitempty"`
	Organization string            `json:"organization,omitempty"`
	Country      string            `json:"country,omitempty"`
	Region       string            `json:"region,omitempty"`
	City         string            `json:"city,omitempty"`
	Timezone     string            `json:"timezone,omitempty"`
	Latitude     *float64          `json:"latitude,omitempty"`
	Longitude    *float64          `json:"longitude,omitempty"`
	ReverseDNS   string            `json:"reverse_dns,omitempty"`
	Headers      map[string]string `json:"headers,omitempty"`
	Method       string            `json:"method,omitempty"`
	Path         string            `json:"path,omitempty"`
}

var pageTemplate = template.Must(template.New("index").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>codifyworx IP</title>
  <style>
    :root { color-scheme: light dark; font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; }
    body { margin: 0; padding: 2rem; line-height: 1.45; }
    main { max-width: 760px; margin: 0 auto; }
    .ip { font-size: clamp(2rem, 8vw, 5rem); font-weight: 800; letter-spacing: -0.08em; overflow-wrap: anywhere; }
    .grid { display: grid; grid-template-columns: 10rem 1fr; gap: .5rem 1rem; margin: 2rem 0; }
    code { background: color-mix(in srgb, CanvasText 8%, transparent); padding: .1rem .35rem; border-radius: .25rem; }
    a { color: inherit; }
    @media (max-width: 640px) { body { padding: 1rem; } .grid { grid-template-columns: 1fr; } }
  </style>
</head>
<body>
<main>
  <div>Your public IP is</div>
  <div class="ip">{{.IP}}</div>
  <div class="grid">
    <div>Version</div><div>{{.Version}}</div>
    <div>ASN</div><div>{{if .ASN}}{{.ASN}}{{else}}unavailable{{end}}</div>
    <div>ISP</div><div>{{or .ISP "unavailable"}}</div>
    <div>Location</div><div>{{or .City "unknown"}}{{if .Region}}, {{.Region}}{{end}}{{if .Country}}, {{.Country}}{{end}}</div>
    <div>User agent</div><div>{{index .Headers "user-agent"}}</div>
  </div>
  <p>CLI: <code>curl {{.BaseURL}}ip</code> or <code>curl {{.BaseURL}}json</code></p>
  <p>Endpoints: <a href="{{.BaseURL}}ip">ip</a>, <a href="{{.BaseURL}}json">json</a>, <a href="{{.BaseURL}}headers">headers</a>, <a href="{{.BaseURL}}all">all</a>, <a href="{{.BaseURL}}all.json">all.json</a>.</p>
</main>
</body>
</html>`))

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	a, err := newApp(cfg)
	if err != nil {
		log.Fatalf("init: %v", err)
	}
	defer a.close()

	mux := http.NewServeMux()
	mux.HandleFunc("/", a.route)

	server := &http.Server{
		Addr:         cfg.addr,
		Handler:      requestLogger(mux),
		ReadTimeout:  cfg.readTimeout,
		WriteTimeout: cfg.writeTimeout,
		IdleTimeout:  cfg.idleTimeout,
	}

	log.Printf("listening on %s base_path=%q", cfg.addr, cfg.basePath)
	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func loadConfig() (config, error) {
	cfg := config{
		addr:         env("ADDR", ":8080"),
		basePath:     normalizeBasePath(env("BASE_PATH", "/ip")),
		geoCityPath:  os.Getenv("GEOIP_CITY_DB"),
		geoASNPath:   os.Getenv("GEOIP_ASN_DB"),
		readTimeout:  5 * time.Second,
		writeTimeout: 10 * time.Second,
		idleTimeout:  60 * time.Second,
	}

	cidrs := env("TRUSTED_PROXY_CIDRS", "127.0.0.1/32,::1/128,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16")
	for _, raw := range strings.Split(cidrs, ",") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		_, network, err := net.ParseCIDR(raw)
		if err != nil {
			return cfg, fmt.Errorf("invalid TRUSTED_PROXY_CIDRS entry %q: %w", raw, err)
		}
		cfg.trustedCIDRs = append(cfg.trustedCIDRs, network)
	}

	return cfg, nil
}

func newApp(cfg config) (*app, error) {
	a := &app{cfg: cfg}
	var err error

	if cfg.geoCityPath != "" {
		a.cityDB, err = geoip2.Open(cfg.geoCityPath)
		if err != nil {
			return nil, fmt.Errorf("open GEOIP_CITY_DB: %w", err)
		}
	}
	if cfg.geoASNPath != "" {
		a.asnDB, err = geoip2.Open(cfg.geoASNPath)
		if err != nil {
			return nil, fmt.Errorf("open GEOIP_ASN_DB: %w", err)
		}
	}

	return a, nil
}

func (a *app) close() {
	if a.cityDB != nil {
		_ = a.cityDB.Close()
	}
	if a.asnDB != nil {
		_ = a.asnDB.Close()
	}
}

func (a *app) route(w http.ResponseWriter, r *http.Request) {
	path := a.stripBasePath(r.URL.Path)

	switch path {
	case "/", "":
		a.handleIndex(w, r)
	case "/ip":
		a.writeText(w, a.clientIP(r).String()+"\n")
	case "/ua":
		a.writeText(w, r.UserAgent()+"\n")
	case "/lang":
		a.writeText(w, r.Header.Get("Accept-Language")+"\n")
	case "/encoding":
		a.writeText(w, r.Header.Get("Accept-Encoding")+"\n")
	case "/mime":
		a.writeText(w, r.Header.Get("Accept")+"\n")
	case "/charset":
		a.writeText(w, r.Header.Get("Accept-Charset")+"\n")
	case "/forwarded":
		a.writeText(w, r.Header.Get("X-Forwarded-For")+"\n")
	case "/headers":
		a.handleHeaders(w, r)
	case "/json", "/geo", "/all.json":
		a.writeJSON(w, a.describe(r, path == "/headers" || path == "/all.json"))
	case "/all":
		a.handleAllText(w, r)
	case "/healthz":
		a.writeText(w, "ok\n")
	default:
		http.NotFound(w, r)
	}
}

func (a *app) stripBasePath(path string) string {
	if a.cfg.basePath == "" || a.cfg.basePath == "/" {
		return path
	}
	if path == a.cfg.basePath {
		return "/"
	}
	return strings.TrimPrefix(path, a.cfg.basePath)
}

func (a *app) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := struct {
		ipResponse
		BaseURL string
	}{
		ipResponse: a.describe(r, true),
		BaseURL:    a.baseURL(),
	}
	if err := pageTemplate.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *app) handleHeaders(w http.ResponseWriter, r *http.Request) {
	a.writeJSON(w, map[string]any{
		"ip":      a.clientIP(r).String(),
		"headers": headersToMap(r),
	})
}

func (a *app) handleAllText(w http.ResponseWriter, r *http.Request) {
	info := a.describe(r, true)
	lines := []string{
		"ip: " + info.IP,
		"version: " + info.Version,
		"asn: " + optionalUint(info.ASN),
		"isp: " + info.ISP,
		"organization: " + info.Organization,
		"country: " + info.Country,
		"region: " + info.Region,
		"city: " + info.City,
		"timezone: " + info.Timezone,
		"reverse_dns: " + info.ReverseDNS,
		"user_agent: " + r.UserAgent(),
		"method: " + r.Method,
		"language: " + r.Header.Get("Accept-Language"),
		"referer: " + r.Referer(),
		"encoding: " + r.Header.Get("Accept-Encoding"),
		"mime: " + r.Header.Get("Accept"),
		"charset: " + r.Header.Get("Accept-Charset"),
		"forwarded: " + r.Header.Get("X-Forwarded-For"),
	}
	a.writeText(w, strings.Join(lines, "\n")+"\n")
}

func (a *app) describe(r *http.Request, includeHeaders bool) ipResponse {
	ip := a.clientIP(r)
	resp := ipResponse{
		IP:      ip.String(),
		Version: ipVersion(ip),
		Method:  r.Method,
		Path:    r.URL.Path,
	}
	if includeHeaders {
		resp.Headers = headersToMap(r)
	}

	if names, err := net.LookupAddr(ip.String()); err == nil && len(names) > 0 {
		resp.ReverseDNS = strings.TrimSuffix(names[0], ".")
	}

	if a.asnDB != nil {
		if record, err := a.asnDB.ASN(ip); err == nil {
			asn := uint(record.AutonomousSystemNumber)
			resp.ASN = &asn
			resp.Organization = record.AutonomousSystemOrganization
			resp.ISP = record.AutonomousSystemOrganization
		}
	}

	if a.cityDB != nil {
		if record, err := a.cityDB.City(ip); err == nil {
			resp.Country = record.Country.IsoCode
			resp.City = record.City.Names["en"]
			resp.Timezone = record.Location.TimeZone
			if len(record.Subdivisions) > 0 {
				resp.Region = record.Subdivisions[0].Names["en"]
			}
			if record.Location.Latitude != 0 || record.Location.Longitude != 0 {
				lat := record.Location.Latitude
				lon := record.Location.Longitude
				resp.Latitude = &lat
				resp.Longitude = &lon
			}
		}
	}

	return resp
}

func (a *app) clientIP(r *http.Request) net.IP {
	remoteIP := parseHostIP(r.RemoteAddr)
	if !a.isTrustedProxy(remoteIP) {
		return remoteIP
	}

	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		parts := strings.Split(forwardedFor, ",")
		for i := len(parts) - 1; i >= 0; i-- {
			ip := net.ParseIP(strings.TrimSpace(parts[i]))
			if ip != nil && !a.isTrustedProxy(ip) {
				return ip
			}
		}
	}

	for _, header := range []string{"X-Real-IP", "CF-Connecting-IP", "True-Client-IP"} {
		ip := net.ParseIP(strings.TrimSpace(r.Header.Get(header)))
		if ip != nil {
			return ip
		}
	}

	return remoteIP
}

func (a *app) isTrustedProxy(ip net.IP) bool {
	if ip == nil {
		return false
	}
	for _, cidr := range a.cfg.trustedCIDRs {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}

func (a *app) baseURL() string {
	if a.cfg.basePath == "" || a.cfg.basePath == "/" {
		return "/"
	}
	return a.cfg.basePath + "/"
}

func (a *app) writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *app) writeText(w http.ResponseWriter, value string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(value))
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}

func parseHostIP(remoteAddr string) net.IP {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	return net.ParseIP(host)
}

func ipVersion(ip net.IP) string {
	if ip == nil {
		return "unknown"
	}
	if ip.To4() != nil {
		return "IPv4"
	}
	return "IPv6"
}

func headersToMap(r *http.Request) map[string]string {
	headers := make(map[string]string, len(r.Header)+1)
	for name, values := range r.Header {
		headers[strings.ToLower(name)] = strings.Join(values, ", ")
	}
	headers["user-agent"] = r.UserAgent()
	return headers
}

func optionalUint(value *uint) string {
	if value == nil {
		return ""
	}
	return strconv.FormatUint(uint64(*value), 10)
}

func env(name string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func normalizeBasePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" || path == "/" {
		return ""
	}
	path = "/" + strings.Trim(path, "/")
	return path
}
