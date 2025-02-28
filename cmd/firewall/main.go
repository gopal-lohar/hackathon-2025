package main

import (
	"github.com/rudransh-shrivastava/context-aware-firewall/internal/firewall"
)

func main() {
	firewall := firewall.NewFirewall()
	firewall.Run()
}
