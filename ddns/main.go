package ddns

import (
	"flag"
)

var (
	defaultPath = "./etc/ddns.conf"
)

func Main() {
	cnf := flag.String("c", "", "config file")
	flag.Parse()

	if *cnf == "" {
		*cnf = defaultPath
	}

	conf, err := ParseConfig(*cnf)
	if err != nil {
		LogErr("parse config fail: %v", err)
		return
	}

	store, err := NewStore(conf.StoreConfig)
	if err != nil {
		LogErr("new store fail: %v", err)
		return
	}
	defer store.Close()

	api := NewApi(conf.ApiConfig, store)

	ddns, err := NewDDNS(conf.DDNSConfig, store)
	if err != nil {
		LogErr("new ddns fail: %v", err)
		return
	}

	go api.Run()
	defer api.Stop()

	ddns.Run()
	defer ddns.Stop()
}
