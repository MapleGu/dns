package main

import (
	"encoding/json"
	"github.com/MapleGu/dns/dns"
	"github.com/pkg/errors"
	"golang.org/x/net/dns/dnsmessage"
	"net/http"
)

var (
	errCheckAuthFail = errors.New("auth fail")
)

// HTTPServer http server interface
type HTTPServer interface {
	Create() http.HandlerFunc
	Read() http.HandlerFunc
	Update() http.HandlerFunc
	Delete() http.HandlerFunc
}

// HTTPService implement HTTPServer
type HTTPService struct {
	Book  dns.Store
	token string
}

// POSTRequest post request data
type POSTRequest struct {
	Domain string   `json:"domain"`
	IP     []string `json:"ip"`
}

// Create http handler to create DNS
func (s *HTTPService) Create(w http.ResponseWriter, r *http.Request) {
	if !s.checkAuth(r) {
		http.Error(w, errCheckAuthFail.Error(), http.StatusUnauthorized)
		return
	}
	var req POSTRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resource, err := toResource(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// s.Dn.save(ntString(resource.Header.Name, resource.Header.Type), resource, nil)
	w.WriteHeader(http.StatusCreated)
}

func (s *HTTPService) checkAuth(r *http.Request) bool {
	if r.Header.Get("Authorize") == s.token {
		return true
	}
	return false
}

func toResource(req POSTRequest) (dnsmessage.Resource, error) {
	rName, err := dnsmessage.NewName(req.Domain)
	none := dnsmessage.Resource{}
	if err != nil {
		return none, err
	}

	var rType dnsmessage.Type
	var rBody dnsmessage.ResourceBody

	switch req {
	case condition:

	}
}
