//go:build !lambda

package main

import (
	"flag"
	"log"
	"strings"
	"time"
)

func main() {
	flag.Parse()

	for _, filePath := range flag.Args() {
		output := strings.TrimSuffix(filePath, ".pdf") + ".png"
		log.Println("--", filePath)
		start := time.Now()
		err := renderPage(filePath, output)
		if err != nil {
			log.Println(filePath, err)
			continue
		}
		elap := time.Since(start)
		log.Println("==", filePath, elap.String())
	}
}
