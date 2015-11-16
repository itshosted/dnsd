package main

import (
	"fmt"
	"flag"
	"github.com/miekg/dns"
	"os"
	"bufio"
	"io"
	"bytes"
	"path/filepath"
)

var listd string

type Handle struct {
}
func (h *Handle) ServeDNS(w dns.ResponseWriter, req *dns.Msg) {
	fmt.Printf("Req %+v\n", req)

	// Blacklisted?

	// Forward
	c := new(dns.Client)
	res, rtt, err := c.Exchange(req, "8.8.8.8:53");
	if err != nil {
		fmt.Printf("Lookup fail %s", err.Error())
		m := new(dns.Msg)
		for _, r := range req.Extra {
			if r.Header().Rrtype == dns.TypeOPT {
				m.SetEdns0(4096, r.(*dns.OPT).Do())
			}
		}
		m.SetRcode(req, dns.RcodeRefused)
		w.WriteMsg(m)
		return
	}

	fmt.Printf("%s: request took %s\n", w.RemoteAddr(), rtt)
	w.WriteMsg(res)
}

func visit(path string, f os.FileInfo, err error) error {
	if path == listd {
		// ignore root
		return nil
	}
	cmps := [][]byte{
		[]byte("127.0.0.1"),
		[]byte("0.0.0.0"),
	}
	seps := [][]byte{
		[]byte(" "),
		[]byte("	"),
	}
	cmt := []byte("#")

	fmt.Printf("Parse %s\n", path)
	fd, e := os.Open(path)
	if e != nil {
		return e
	}
	r := bufio.NewReader(fd)

	for {
		line, e := r.ReadBytes('\n')
		line = bytes.TrimSpace(line)
		if e == io.EOF {
			return nil
		}
		if e != nil {
			return e
		}
		// Strip empty lines + comments
		if len(line) > 0 && line[0] != '#' {
			// Strip lines not beginning with cmp
			ok := false
			for _, cmp := range cmps {
				if bytes.Compare(line[0:len(cmp)], cmp) == 0 {
					ok = true
					break
				}
			}
			if !ok {
				fmt.Printf("WARN: Skip %s\n", line)
				continue
			}

			// 127.0.0.1 domain
			idx := -1
			for _, sep := range seps {
				idx = bytes.Index(line, sep)
				if idx != -1 {
					break
				}
			}
			if idx == -1 {
				fmt.Printf("WARN: ParseErr %s\n", line)
				continue
			}

			domain := line[idx+1:]
			hidx := bytes.Index(domain, cmt)
			if hidx != -1 {
				// Strip comment
				domain = domain[:hidx]
			}
			// TODO: Save
			//fmt.Printf("%s\n",domain)
		}
	}
	return nil
}

func main() {
	var (
		addr string
	)
	flag.StringVar(&addr, "l", "[::]:53", "listen on (both tcp and udp)")
	flag.StringVar(&listd, "d", "./list.d", "Dir with blacklisted domain files")
	handler := &Handle{}

	if e := filepath.Walk(listd, visit); e != nil {
		panic(e)
	}

	//for _, addr := range strings.Split(listen, ",") {
		fmt.Printf("DNS Listen %s\n", addr)
		go func() {
			if err := dns.ListenAndServe(addr, "udp", handler); err != nil {
				panic(err)
			}
		}()

		if err := dns.ListenAndServe(addr, "tcp", handler); err != nil {
			panic(err)
		}
	//}
}