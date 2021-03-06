package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/RTradeLtd/dpinner/config"
	ipfsapi "github.com/RTradeLtd/go-ipfs-api"
	"github.com/RTradeLtd/thc"
	"github.com/bwmarrin/discordgo"
)

var (
	shell *ipfsapi.Shell
	tc    *thc.V2
)

func main() {
	cfg, err := config.LoadConfig("./config.json")
	if err != nil {
		log.Fatal(err)
	}
	if cfg.Discord.Token == "" {
		cfg.Discord.Token = os.Getenv("DISCORD_TOKEN")
	}
	if cfg.Temporal.User == "" {
		cfg.Temporal.User = os.Getenv("TEMPORAL_USER")
	}
	if cfg.Temporal.Pass == "" {
		cfg.Temporal.Pass = os.Getenv("TEMPORAL_PASS")
	}
	if cfg.Temporal.WarpURL == "" {
		cfg.Temporal.WarpURL = "https://dev.api.ipfs.temporal.cloud"
	}
	tc = thc.NewV2(cfg.Temporal.User, cfg.Temporal.Pass, thc.DevURL)
	if err := tc.Login(); err != nil {
		fmt.Println("failed to login with temporal")
		log.Fatal(err)
	}
	jwt, err := tc.GetJWT()
	if err != nil {
		fmt.Println("failed to get jwt")
		log.Fatal(err)
	}
	shell = ipfsapi.NewDirectShell(cfg.Temporal.WarpURL)
	shell = shell.WithAuthorization(jwt)
	// we need to prepend Bot to allow discord
	// to assign permissions properly
	dg, err := discordgo.New("Bot " + cfg.Discord.Token)
	if err != nil {
		fmt.Println("failed to authenticate with discord")
		log.Fatal(err)
	}
	dg.AddHandler(messageCreate)
	if err := dg.Open(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("bot is now running")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// parse the message contents based off string fields
	args := strings.Fields(m.Content)
	if len(args) == 0 {
		return
	}
	// ensure the first field is a valid invocation of dpinner
	if args[0] != "!dpinner" {
		return
	}
	// If the message is "ping" reply with "Pong!"
	if args[1] == "ping" {
		if _, err := s.ChannelMessageSend(m.ChannelID, "Pong!"); err != nil {
			fmt.Println(err)
		}
		return
	}
	// If the message is "pong" reply with "Ping!"
	if args[1] == "pong" {
		if _, err := s.ChannelMessageSend(m.ChannelID, "Ping!"); err != nil {
			fmt.Println(err)
		}
		return
	}
	if args[1] == "upload" {
		processUpload(s, m.Attachments, m.ChannelID)
	}
	if args[1] == "pin" {
		for i := 2; i < len(args); i++ {
			if err := shell.Pin(args[i]); err != nil {
				s.ChannelMessageSend(m.ChannelID, "failed to process pin request(s)")
				return
			}
		}
		s.ChannelMessageSend(m.ChannelID, "successfully processed pin request(s)")
	}

	if args[1] == "help" {
		msg := &discordgo.MessageEmbed{
			Author: &discordgo.MessageEmbedAuthor{
				Name: "dpinner help menu",
			},
			Description: `available commands are:

			upload: upload an attachment to ipfs

			pin: pin hash(es) with Temporal

			index: index hash(es) with Lens

			search: search Lens for ipfs hashes matching your query

			`,
			Footer: &discordgo.MessageEmbedFooter{
				Text: "dpinner makes use of Lens, Temporal, and Warp. Try Temporal at https://temporal.cloud",
			},
			Fields: []*discordgo.MessageEmbedField{
				{
					Name: "upload",
					Value: `
					when uploading an attachment, use a comment of **!dpinner upload**
					this will upload all attachments associated with the comment to ipfs
					`,
				},
				{
					Name: "pin",
					Value: `
					the command is **!dpinner upload** and takes a list of hashes
					example: **!dpinner upload hash1 hash2 hash3**`,
				},
				{
					Name: "index",
					Value: `
					the command is **!dpinner index** and takes a list of hashes
					example: **!dpinner index hash1 hash2 hash3**`,
				},
				{
					Name: "search",
					Value: `
					the command is **!dpinner search** and takes a search query
					example: **!dpinner search blockchain**`,
				},
			},
		}
		s.ChannelMessageSendEmbed(m.ChannelID, msg)
	}
	if args[1] == "index" {
		for i := 2; i < len(args)-1; i++ {
			if _, err := tc.IndexHash(args[i], args[len(args)-1] == "reindex=true"); err != nil {
				s.ChannelMessageSend(m.ChannelID, "failed to process lens index request(s)")
				return
			}
		}
		s.ChannelMessageSend(m.ChannelID, "successfully indexed hashe(s)")
	}

	if args[1] == "search" {
		processLensSearch(s, m.ChannelID, args)
	}
}

func processLensSearch(s *discordgo.Session, channelID string, args []string) {
	searchArgs := args[2:]
	query := strings.Join(searchArgs, " ")
	resp, err := tc.SearchLens(query)
	if err != nil {
		s.ChannelMessageSend(channelID, "failed to submit search query to lens")
		return
	}
	var (
		results []string
		count   int
	)
	if len(resp.Response.Results) > 10 {
		foundResults := make(map[string]bool)
		for {
			if count == 10 {
				break
			}
			hash := resp.Response.Results[rand.Intn(len(resp.Response.Results))].Doc.Hash
			if foundResults[hash] {
				continue
			}
			results = append(results, hash)
			foundResults[hash] = true
			count++
		}
	} else {
		for _, v := range resp.Response.Results {
			results = append(results, v.Doc.Hash)
		}
	}
	output := strings.Join(results, " ")
	msg := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name: "Lens Search Results",
		},
		Description: output,
	}
	s.ChannelMessageSendEmbed(channelID, msg)
}
func processUpload(s *discordgo.Session, attachments []*discordgo.MessageAttachment, channelID string) {
	for _, v := range attachments {
		fmt.Printf("fetching object %+v\n", v)
		resp, err := http.Get(v.URL)
		if err != nil {
			s.ChannelMessageSend(channelID, "failed to process attachments")
			return
		}
		defer resp.Body.Close()
		hash, err := shell.Add(resp.Body)
		if err != nil {
			s.ChannelMessageSend(channelID, "failed to add attachments to ipfs")
			return
		}
		s.ChannelMessageSend(channelID, fmt.Sprintf("the hash of your file is %s", hash))
	}
}
