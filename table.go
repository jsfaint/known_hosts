package main

import (
	"fmt"

	"github.com/cheynewallace/tabby"
)

func Dump(hosts []string) {
	t := tabby.New()

	t.AddHeader("Name", "IP", "Type")

	for _, v := range hosts {
		if v == "" {
			continue
		}
		h, err := NewHost(v)
		if err != nil {
			fmt.Println(err)
			continue
		}

		t.AddLine(h.Name, h.IP, h.KeyType)
	}

	t.Print()
}
