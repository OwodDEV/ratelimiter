package ratelimiter

import (
	"net/http"
	"net/url"
)

func (m *MasterManager) getOrCreateHandler(w http.ResponseWriter, r *http.Request) {
	tag := r.URL.Query().Get("tag")
	if tag == "" {
		http.Error(w, "tag is required", http.StatusBadRequest)
		return
	}

	_, err := m.GetOrCreate(tag)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (m *MasterManager) setRuleHandler(w http.ResponseWriter, r *http.Request) {
	tag := r.URL.Query().Get("tag")
	ruleStrEscaped := r.URL.Query().Get("rule")
	ruleStr, err := url.QueryUnescape(ruleStrEscaped)
	if err != nil {
		http.Error(w, "rule can not be decoded", http.StatusBadRequest)
		return
	}

	if tag == "" || ruleStr == "" {
		http.Error(w, "tag and rule are required", http.StatusBadRequest)
		return
	}

	rl, err := m.GetOrCreate(tag)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = rl.SetRules(ruleStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (m *MasterManager) allowHandler(w http.ResponseWriter, r *http.Request) {
	tag := r.URL.Query().Get("tag")
	if tag == "" {
		http.Error(w, "tag is required", http.StatusBadRequest)
		return
	}

	rl, err := m.GetOrCreate(tag)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	allowed, err := rl.Allow(tag)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if allowed {
		w.Write([]byte("true"))
	} else {
		w.Write([]byte("false"))
	}
}

func (m *MasterManager) incHandler(w http.ResponseWriter, r *http.Request) {
	tag := r.URL.Query().Get("tag")
	if tag == "" {
		http.Error(w, "tag is required", http.StatusBadRequest)
		return
	}

	rl, err := m.GetOrCreate(tag)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = rl.Inc(tag)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
