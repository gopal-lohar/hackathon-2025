package main

import "github.com/gopal-lohar/hackathon-2025/internal/firewall"

func main() {
	firewall := firewall.NewFirewall()
	firewall.Run()
}
