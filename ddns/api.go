package ddns

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

var (
	defaultApiAddr = ":5843"
	defaultToken   = "SUNLZWxpbkBub3RyLnRlY2g="
)

type ApiConfig struct {
	Addr  string `json:"addr" toml:"addr"`
	Token string `json:"token" toml:"token"`
}

type Api struct {
	addr  string
	token string
	done  chan struct{}
	abort chan struct{}
	s     *Store
}

func NewApi(cfg *ApiConfig, s *Store) *Api {
	addr := cfg.Addr
	if addr == "" {
		addr = defaultApiAddr
	}

	token := cfg.Token
	if token == "" {
		token = defaultToken
	}

	return &Api{
		addr:  addr,
		token: token,
		s:     s,
		done:  make(chan struct{}),
		abort: make(chan struct{}),
	}
}

func (a *Api) Run() {
	http.HandleFunc("/api/v1/build_in_domain/add", a.addDomain)

	go func() {
		LogInfo("api listenning %s", a.addr)
		LogErr("%v", http.ListenAndServe(a.addr, nil))
		close(a.abort)
	}()

	select {
	case <-a.done:
		LogInfo("api module stop signal")
		return

	case <-a.abort:
		LogErr("api module abort")
		return
	}
}

func (a *Api) Stop() {
	close(a.done)
}

type AddDomainForm struct {
	Domain string `json:"domain"`
	IP     string `json:"ip"`
}

func (a *Api) addDomain(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("access-token")
	if token != a.token {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var form AddDomainForm
	err = json.Unmarshal(data, &form)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	if form.Domain == "" {
		w.Write([]byte("empty domain"))
		return
	}

	a.s.Set(form.Domain, form.IP)
	LogInfo("apply %s => %s", form.Domain, form.IP)
}
