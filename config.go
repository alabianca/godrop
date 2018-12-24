package main

type config struct {
	Port          string
	IP            string
	ServiceName   string
	Host          string
	ServiceWeight uint16
	TTL           uint32
	Priority      uint16
}
