package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
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

func dumpGuilds(cfg *config, session *discordgo.Session, cutoff time.Time) error {
	log.Println("Starting dump process...")

	f, err := os.Create(filepath.Join(cfg.Target, "guilds.json"))
	if err != nil {
		return fmt.Errorf("failed to create guilds.json: %v", err)
	}
	defer f.Close()

	beforeID := ""
	chunk := 100

	for {
		userGuilds, err := session.UserGuilds(chunk, beforeID, "")
		if err != nil {
			return fmt.Errorf("failed to fetch guilds: %v", err)
		}

		for _, userGuild := range userGuilds {
			appendStruct(f, userGuild)

			err := dumpGuild(cfg, session, cutoff, userGuild)
			if err != nil {
				return fmt.Errorf("failed to dump guild: %v", err)
			}

			beforeID = userGuild.ID
		}

		if len(userGuilds) < chunk {
			break
		}
	}

	return nil
}

func dumpGuild(cfg *config, session *discordgo.Session, cutoff time.Time, guild *discordgo.UserGuild) error {
	log.Printf("Dumping %s (%s)...", guild.ID, guild.Name)

	err := os.MkdirAll(filepath.Join(cfg.Target, guild.ID), 0755)
	if err != nil {
		return fmt.Errorf("failed to create data target directory: %v", err)
	}

	f, err := os.Create(filepath.Join(cfg.Target, guild.ID, "channels.json"))
	if err != nil {
		return fmt.Errorf("failed to create channels.json: %v", err)
	}
	defer f.Close()

	channels, err := session.GuildChannels(guild.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch channels: %v", err)
	}

	for _, channel := range channels {
		appendStruct(f, channel)

		err = dumpChannel(cfg, session, cutoff, guild, channel)
		if err != nil {
			return fmt.Errorf("failed to dump channel: %v", err)
		}
	}

	return nil
}

func dumpChannel(cfg *config, session *discordgo.Session, cutoff time.Time, guild *discordgo.UserGuild, channel *discordgo.Channel) error {
	logfile := filepath.Join(cfg.Target, guild.ID, fmt.Sprintf("%s.json", channel.ID))
	beforeID := findOldestKnown(logfile)

	if beforeID != "" {
		log.Printf("  Dumping %s (%s) (resuming at %s)...", channel.ID, channel.Name, beforeID)
	} else {
		log.Printf("  Dumping %s (%s)...", channel.ID, channel.Name)
	}

	fp, err := os.OpenFile(logfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer fp.Close()

	chunk := 100

	for {
		log.Printf("    Fetching %d starting at %s...", chunk, beforeID)

		messages, err := session.ChannelMessages(channel.ID, chunk, beforeID, "", "")
		if err != nil {
			return err
		}

		oldest := time.Time{}

		for _, msg := range messages {
			appendStruct(fp, msg)
			beforeID = msg.ID
			oldest, _ = msg.Timestamp.Parse()
		}

		if oldest.Before(cutoff) {
			break
		}

		time.Sleep(1 * time.Second)
	}

	return nil
}

func logMessage(f *os.File, msg interface{}) {
	encoded, _ := json.Marshal(msg)
	f.Write(append(encoded, '\n'))
}

func findOldestKnown(filename string) string {
	f, err := os.Open(filename)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lastLine := ""
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lastLine = line
		}
	}

	type msg struct {
		ID string `json:"id"`
	}

	m := msg{}
	json.Unmarshal([]byte(lastLine), &m)

	return m.ID
}

func appendStruct(f *os.File, s interface{}) {
	encoded, _ := json.Marshal(s)
	f.Write(append(encoded, '\n'))
}
