package main

import (
	"flag"
	"os"
	"time"

	"github.com/bwmarrin/discordgo"
)

var cutoffFlag = flag.String("cutoff", "", "stop when reaching this timestamp")

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		logFatal("No configuration file given.\n")
	}

	config, err := loadConfig(flag.Args()[0])
	if err != nil {
		logFatal("Failed to load config: %v\n", err)
	}

	cutoff := time.Time{}
	if cutoffFlag != nil && *cutoffFlag != "" {
		parsed, err := time.Parse("2006-01-02", *cutoffFlag)
		if err != nil {
			logFatal("Invalid -cutoff value: %v\n", err)
		}

		cutoff = parsed

		logPrint("Stopping data collection at %s.\n", cutoff.Format("Mon, 02 Jan 2006"))
	}

	err = os.MkdirAll(config.Target, 0755)
	if err != nil {
		logFatal("Failed to create data target directory: %v\n", err)
	}

	session, err := discordgo.New(config.Username, config.Password)
	if err != nil {
		logFatal("Failed to open Discord session: %v\n", err)
	}

	err = dumpGuilds(config, session, cutoff)
	if err != nil {
		logFatal("Failed to dump data: %v\n", err)
	}
}
