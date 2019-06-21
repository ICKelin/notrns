package ddns

import (
	"fmt"
	"net"
	"strings"

	"github.com/miekg/dns"
)

var (
	defaultDDNSAddr = ":53"
)

type DDNSConfig struct {
	Addr string `json:"ddns_server" toml:"addr"`
}

type DDNS struct {
	s     *Store
	addr  string
	conn  *net.UDPConn
	done  chan struct{}
	abort chan struct{}
}

func NewDDNS(cfg *DDNSConfig, s *Store) (*DDNS, error) {
	addr := cfg.Addr
	if addr == "" {
		addr = defaultDDNSAddr
	}

	udpAddr, err := net.ResolveUDPAddr("udp", cfg.Addr)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, err
	}

	ddns := &DDNS{
		s:    s,
		addr: addr,
		conn: conn,
		done: make(chan struct{}),
	}
	return ddns, nil
}

func (d *DDNS) Run() {
	defer d.conn.Close()

	LogInfo("ddns server listening on %s", d.addr)

	go func() {
		for {
			bytes := make([]byte, 512)
			nr, addr, err := d.conn.ReadFromUDP(bytes)
			if err != nil {
				LogErr("read from udp fail: %v", err)
				break
			}

			go d.onPacket(bytes[:nr], addr)
		}

	}()

	select {
	case <-d.done:
		LogInfo("ddns module done signal")
		return

	case <-d.abort:
		LogInfo("ddns module abort")
		return
	}
}

func (d *DDNS) Stop() {
	close(d.done)
}

func (d *DDNS) onPacket(pkt []byte, addr *net.UDPAddr) {
	var req = dns.Msg{}
	err := req.Unpack(pkt)
	if err != nil {
		LogErr("handle packet fail: %v", err)
		return
	}

	question, err := d.getQuestion(&req)
	if err != nil {
		LogErr("get question fail: %v", err)
		return
	}

	answer, err := d.s.Get(question)
	if err != nil {
		LogErr("get answer for", question, "fail: ", err)
		return
	}

	ansip, ok := answer.(string)
	if !ok {
		LogErr("invalid answer %v", answer)
		return
	}

	LogInfo("question: %s answer %s", question, ansip)
	err = d.response(&req, ansip, addr)
	if err != nil {
		LogErr("response domain:", question, " ans ", answer, " for client fail: ", err)
	}
}

func (d *DDNS) response(req *dns.Msg, ans string, addr *net.UDPAddr) error {
	nip := net.ParseIP(ans)
	question := req.Question[0]
	ra := &dns.A{
		Hdr: dns.RR_Header{
			Name:     question.Name,
			Rrtype:   dns.TypeA,
			Class:    question.Qclass,
			Ttl:      500,
			Rdlength: 4,
		},
		A: nip.To4(),
	}

	rsp := req.Copy()
	rsp.Answer = []dns.RR{ra}
	rsp.Response = true
	rsp.Rcode = dns.RcodeSuccess
	rsp.RecursionAvailable = true

	bytes, err := rsp.Pack()
	if err != nil {
		return err
	}

	_, err = d.conn.WriteToUDP(bytes, addr)
	return err
}

func (d *DDNS) getQuestion(msg *dns.Msg) (string, error) {
	if len(msg.Question) != 1 {
		return "", fmt.Errorf("only support one question")
	}

	return strings.TrimSuffix(msg.Question[0].Name, "."), nil
}
