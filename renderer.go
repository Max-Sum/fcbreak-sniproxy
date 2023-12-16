package main

import (
	"encoding/base64"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Max-Sum/fcbreak"
)

type Renderer struct {
	infopath string
	tmplpath string
	auth     bool
	username string
	password string
}

type RendererVars struct {
	Request     *http.Request
	Info        *fcbreak.ServiceInfo
	Username    string
	Password    string
	RemoteHost  string
	RemotePort  string
	ExposedHost string
	ExposedPort string
	ProxyHost   string
	ProxyPort   string
}

func NewRenderer(infopath, tmplpath string, auth bool, username, password string) *Renderer {
	p := &Renderer{
		infopath: infopath,
		tmplpath: tmplpath,
		auth:     auth,
		username: username,
		password: password,
	}
	return p
}

func (r *Renderer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ok, user, pass := r.Auth(req, "Authorization")
	if !ok {
		rw.Header().Set("WWW-Authenticate", `Basic realm="Restricted API", charset="UTF-8"`)
		http.Error(rw, "Unauthorized", http.StatusUnauthorized)
		return
	}
	vars := &RendererVars{
		Request:  req,
		Username: user,
		Password: pass,
	}
	info, err := readServiceInfo(r.infopath)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	vars.Info = info
	if info.RemoteAddr != "" {
		host, portStr, err := net.SplitHostPort(info.RemoteAddr)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		vars.RemoteHost = host
		vars.RemotePort = portStr
	}
	if info.ExposedAddr != "" {
		host, portStr, err := net.SplitHostPort(info.ExposedAddr)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		vars.ExposedHost = host
		vars.ExposedPort = portStr
	}
	if info.ProxyAddr != "" {
		host, portStr, err := net.SplitHostPort(info.ProxyAddr)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		vars.ProxyHost = host
		vars.ProxyPort = portStr
	}

	r.Handler(rw, vars)
}

func (r *Renderer) Handler(rw http.ResponseWriter, vars *RendererVars) {
	upath := vars.Request.URL.Path
	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
		vars.Request.URL.Path = upath
	}
	name := filepath.Join(r.tmplpath, path.Clean(upath))
	if info, err := os.Stat(name); os.IsNotExist(err) {
		http.NotFound(rw, vars.Request)
		return
	} else if info.IsDir() {
		http.Error(rw, "", http.StatusForbidden)
		return
	}
	tmpl, err := template.ParseFiles(name)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(rw, vars)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (r *Renderer) Auth(req *http.Request, header string) (bool, string, string) {
	if !r.auth {
		return true, "", ""
	}
	u, p, ok := basicAuth(req, header)
	if !ok {
		log.Printf("Authenication Failed: Failed to get auth info.\n")
		return false, u, p
	}
	if (r.username != "" && u != r.username) || (r.password != "" && p != r.password) {
		log.Printf("Authenication Failed: %s:%s not matched.\n", u, p)
		return false, u, p
	}
	return true, u, p
}

func basicAuth(req *http.Request, header string) (username, password string, ok bool) {
	auth := req.Header.Get(header)
	if auth == "" {
		return "", "", false
	}
	return parseBasicAuth(auth)
}

// parseBasicAuth parses an HTTP Basic Authentication string.
// "Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==" returns ("Aladdin", "open sesame", true).
func parseBasicAuth(auth string) (username, password string, ok bool) {
	const prefix = "Basic "
	if len(auth) < len(prefix) {
		return "", "", false
	}
	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return "", "", false
	}
	cs := string(c)
	username, password, ok = strings.Cut(cs, ":")
	if !ok {
		return "", "", false
	}
	return username, password, true
}
