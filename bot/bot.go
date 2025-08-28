package bot

import (
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/Sumire-Labs/Luna/config"
	"github.com/Sumire-Labs/Luna/database"
)

type Bot struct {
	session  *discordgo.Session
	config   *config.Config
	db       *database.Service
	startTime time.Time
}

func New(session *discordgo.Session, cfg *config.Config, db *database.Service) *Bot {
	return &Bot{
		session:  session,
		config:   cfg,
		db:       db,
		startTime: time.Now(),
	}
}

func (b *Bot) Start() error {
	b.session.AddHandler(b.onReady)
	b.session.AddHandler(b.onGuildCreate)
	b.session.AddHandler(b.onMessageCreate)

	if err := b.session.Open(); err != nil {
		return fmt.Errorf("failed to open Discord session: %w", err)
	}

	log.Println("Bot is now running. Press CTRL+C to exit.")
	return nil
}

func (b *Bot) Stop() error {
	return b.session.Close()
}

func (b *Bot) onReady(s *discordgo.Session, event *discordgo.Ready) {
	log.Printf("Logged in as: %v#%v", event.User.Username, event.User.Discriminator)
	
	status := b.config.Bot.StatusMessage
	if status == "" {
		status = fmt.Sprintf("Luna Bot | %d servers", len(event.Guilds))
	}

	s.UpdateStatusComplex(discordgo.UpdateStatusData{
		Activities: []*discordgo.Activity{
			{
				Name: status,
				Type: discordgo.ActivityType(b.config.Bot.ActivityType),
			},
		},
		Status: "online",
	})
}

func (b *Bot) onGuildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	if event.Guild.Unavailable {
		return
	}

	err := b.db.UpsertGuild(event.Guild.ID, event.Guild.Name, b.config.Bot.Prefix)
	if err != nil {
		log.Printf("Failed to upsert guild %s: %v", event.Guild.ID, err)
	}
}

func (b *Bot) onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID || m.Author.Bot {
		return
	}

	err := b.db.UpsertUser(
		m.Author.ID,
		m.Author.Username,
		m.Author.Discriminator,
		m.Author.Avatar,
		m.Author.Bot,
	)
	if err != nil {
		log.Printf("Failed to upsert user %s: %v", m.Author.ID, err)
	}
	
	// Count bracket pairs in message
	if m.GuildID != "" {
		halfWidthPairs, fullWidthPairs := b.countBracketPairs(m.Content)
		totalPairs := halfWidthPairs + fullWidthPairs
		
		// Update database if any bracket pairs found
		if totalPairs > 0 {
			err := b.db.UpdateBracketUsage(m.GuildID, m.Author.ID, halfWidthPairs, fullWidthPairs)
			if err != nil {
				log.Printf("Failed to update bracket usage: %v", err)
			}
		}
	}
}

// countBracketPairs counts complete bracket pairs in the message
func (b *Bot) countBracketPairs(content string) (halfWidth, fullWidth int) {
	// Count half-width bracket pairs ()
	halfWidth = b.countPairs(content, '(', ')')
	
	// Count full-width bracket pairs （）
	fullWidth = b.countPairs(content, '（', '）')
	
	return
}

// countPairs counts complete pairs of open and close characters
func (b *Bot) countPairs(content string, open, close rune) int {
	openCount := 0
	pairs := 0
	
	for _, r := range content {
		if r == open {
			openCount++
		} else if r == close && openCount > 0 {
			openCount--
			pairs++
		}
	}
	
	return pairs
}

func (b *Bot) GetUptime() time.Duration {
	return time.Since(b.startTime)
}

func (b *Bot) GetSession() *discordgo.Session {
	return b.session
}

func (b *Bot) GetConfig() *config.Config {
	return b.config
}

func (b *Bot) GetDatabase() *database.Service {
	return b.db
}