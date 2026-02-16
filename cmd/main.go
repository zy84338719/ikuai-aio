package main

import (
	"log"

	"github.com/NERVEbing/ikuai-aio/config"
	"github.com/NERVEbing/ikuai-aio/exporter"
	"github.com/NERVEbing/ikuai-aio/job"
)

func main() {
	c := config.Load()

	go func() {
		if err := job.Run(c); err != nil {
			log.Println("Job error:", err)
		}
	}()

	if err := exporter.Run(c); err != nil {
		log.Fatalln(err)
	}
}
