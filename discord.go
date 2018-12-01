package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bwmarrin/discordgo"
)

func dumpGuilds(cfg *config, session *discordgo.Session, cutoff time.Time) error {
	logPrint("Starting dump process...\n")

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

			if cfg.isGuildIgnored(userGuild.ID) {
				continue
			}

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
	logPrint("Dumping %s (%s)...\n", guild.ID, guild.Name)

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

		if cfg.isChannelIgnored(channel.ID) {
			continue
		}

		err = dumpChannel(cfg, session, cutoff, guild, channel)
		if err != nil {
			return fmt.Errorf("failed to dump channel: %v", err)
		}
	}

	return nil
}

func dumpChannel(cfg *config, session *discordgo.Session, cutoff time.Time, guild *discordgo.UserGuild, channel *discordgo.Channel) error {
	logPrint("  Dumping %s (%s)...\n", channel.ID, channel.Name)

	logfile := filepath.Join(cfg.Target, guild.ID, fmt.Sprintf("%s.json", channel.ID))
	beforeID := findOldestKnown(logfile)

	fp, err := os.OpenFile(logfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer fp.Close()

	chunk := 100

	for {
		logPrint("    Fetching %d...", chunk)

		messages, err := session.ChannelMessages(channel.ID, chunk, beforeID, "", "")
		if err != nil {
			// do not fail if we simply have no access to a channel
			if restError, ok := err.(*discordgo.RESTError); ok && restError.Message != nil && restError.Message.Code == discordgo.ErrCodeMissingAccess {
				logEndLine(" error: no access to this channel.")
				return nil
			}

			logEndLine("")
			return err
		}

		if len(messages) == 0 {
			logEndLine(" no further messages.")
			break
		}

		oldest := time.Time{}

		for _, msg := range messages {
			appendStruct(fp, msg)
			beforeID = msg.ID
			oldest, _ = msg.Timestamp.Parse()
		}

		logEndLine(" reached %s", oldest.Format(time.RFC822))

		if oldest.Before(cutoff) {
			break
		}

		time.Sleep(1 * time.Second)
	}

	return nil
}
