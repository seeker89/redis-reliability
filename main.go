package main

import "github.com/seeker89/redis-resiliency-toolkit/cmd"

// these will be injected during build
var (
	Version, Build string
)

func main() {
	cmd.Execute(Version, Build)
}
