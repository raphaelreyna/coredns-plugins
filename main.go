package main

import (
	"fmt"

	"github.com/miekg/dns"
)

func main() {
	conn, err := dns.Dial("udp", "1.1.1.1:53")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	reqMsg := new(dns.Msg)
	reqMsg.SetQuestion("example.com.", dns.TypeA)

	err = conn.WriteMsg(reqMsg)
	if err != nil {
		panic(err)
	}

	respMsg, err := conn.ReadMsg()
	if err != nil {
		panic(err)
	}

	fmt.Printf("ttl: %d\n", respMsg.Answer[0].Header().Ttl)

	fmt.Println(respMsg)
}
