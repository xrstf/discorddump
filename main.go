package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
)

var cutoffFlag = flag.String("cutoff", "", "stop when reaching this timestamp")

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		log.Fatal("No configuration file given.")
	}

	config, err := loadConfig(flag.Args()[0])
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	cutoff := time.Time{}
	if cutoffFlag != nil && *cutoffFlag != "" {
		parsed, err := time.Parse("2006-01-02", *cutoffFlag)
		if err != nil {
			log.Fatalf("Invalid -cutoff value: %v", err)
		}

		cutoff = parsed

		log.Printf("Stopping data collection at %s.", cutoff.Format("Mon, 02 Jan 2006"))
	}

	err = os.MkdirAll(config.Target, 0755)
	if err != nil {
		log.Fatalf("Failed to create data target directory: %v", err)
	}

	sesseion, err := discordgo.New(config.Username, config.Password)
	if err != nil {
		log.Fatalf("Failed to open Discord session: %v", err)
	}

	err = dumpGuilds(config, sesseion, cutoff)
	if err != nil {
		log.Fatalf("Failed to dump data: %v", err)
	}
}
