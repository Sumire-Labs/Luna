package commands

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/Sumire-Labs/Luna/embed"
)

type AvatarCommand struct{}

func NewAvatarCommand() *AvatarCommand {
	return &AvatarCommand{}
}

func (c *AvatarCommand) Name() string {
	return "avatar"
}

func (c *AvatarCommand) Description() string {
	return "ユーザーのアバターとバナーを表示します"
}

func (c *AvatarCommand) Usage() string {
	return "/avatar [user]"
}

func (c *AvatarCommand) Category() string {
	return "ユーティリティ"
}

func (c *AvatarCommand) Aliases() []string {
	return []string{"av", "pfp"}
}

func (c *AvatarCommand) Permission() int64 {
	return 0
}

func (c *AvatarCommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionUser,
			Name:        "user",
			Description: "アバターを表示するユーザー",
			Required:    false,
		},
		{
			Type:        discordgo.ApplicationCommandOptionBoolean,
			Name:        "show_banner",
			Description: "ユーザーのバナーも表示する",
			Required:    false,
		},
	}
}

func (c *AvatarCommand) Execute(ctx *Context) error {
	var targetUser *discordgo.User
	
	if userArg, ok := ctx.GetArg("user"); ok {
		if user, ok := userArg.(*discordgo.User); ok {
			targetUser = user
		}
	}
	
	if targetUser == nil {
		targetUser = ctx.GetUser()
	}

	showBanner := ctx.GetBoolArg("show_banner")

	if err := ctx.DeferReply(false); err != nil {
		return err
	}

	member, err := ctx.Session.GuildMember(ctx.GetGuild(), targetUser.ID)
	if err != nil {
		member = &discordgo.Member{User: targetUser}
	}

	avatarURL := targetUser.AvatarURL("4096")
	if avatarURL == "" {
		// Calculate default avatar URL based on discriminator or user ID
		var discriminatorInt int
		if targetUser.Discriminator != "" && targetUser.Discriminator != "0" {
			// Legacy username system
			discriminatorInt = 0
			if d := targetUser.Discriminator; len(d) > 0 {
				discriminatorInt = int(d[len(d)-1] - '0')
			}
		} else {
			// New username system - use user ID
			userID := targetUser.ID
			if len(userID) >= 2 {
				discriminatorInt = int((userID[len(userID)-2] - '0') + (userID[len(userID)-1] - '0'))
			}
			discriminatorInt = discriminatorInt % 5
		}
		avatarURL = fmt.Sprintf("https://cdn.discordapp.com/embed/avatars/%d.png", discriminatorInt)
	}

	embedBuilder := embed.New().
		SetTitle(fmt.Sprintf("👤 %s のプロフィール", targetUser.Username)).
		SetColor(c.getUserColor(member))

	embedBuilder.SetThumbnail(avatarURL)

	formats := c.getAvatarFormats(avatarURL)
	embedBuilder.AddField("🖼️ アバター形式", formats, false)

	sizes := c.getAvatarSizes(avatarURL)
	embedBuilder.AddField("📐 利用可能なサイズ", sizes, false)

	if showBanner {
		bannerURL := c.getUserBannerURL(targetUser)
		if bannerURL != "" {
			embedBuilder.SetImage(bannerURL)
			embedBuilder.AddField("🎨 バナー", "[フルサイズで表示]("+bannerURL+")", false)
		} else {
			embedBuilder.AddField("🎨 バナー", "カスタムバナーは設定されていません", false)
		}
	}

	userInfo := c.getUserInfo(targetUser, member)
	embedBuilder.AddField("ℹ️ ユーザー情報", userInfo, false)

	embedBuilder.SetFooter(
		fmt.Sprintf("ID: %s", targetUser.ID),
		"",
	)

	return ctx.EditReplyEmbed(embedBuilder.Build())
}

func (c *AvatarCommand) getUserColor(member *discordgo.Member) int {
	if member != nil && len(member.Roles) > 0 {
		return embed.M3Colors.Primary
	}
	return embed.M3Colors.Surface
}

func (c *AvatarCommand) getAvatarFormats(baseURL string) string {
	if baseURL == "" {
		return "デフォルトアバター"
	}

	urlParts := strings.Split(baseURL, ".")
	if len(urlParts) < 2 {
		return "不明な形式"
	}

	baseURLWithoutExt := strings.Join(urlParts[:len(urlParts)-1], ".")
	
	formats := []string{
		fmt.Sprintf("[PNG](%s.png)", baseURLWithoutExt),
		fmt.Sprintf("[JPG](%s.jpg)", baseURLWithoutExt),
		fmt.Sprintf("[WebP](%s.webp)", baseURLWithoutExt),
	}

	if strings.Contains(baseURL, "a_") {
		formats = append(formats, fmt.Sprintf("[GIF](%s.gif)", baseURLWithoutExt))
	}

	return strings.Join(formats, " • ")
}

func (c *AvatarCommand) getAvatarSizes(baseURL string) string {
	if baseURL == "" {
		return "N/A"
	}

	sizes := []string{"128", "256", "512", "1024", "2048", "4096"}
	sizeLinks := make([]string, len(sizes))

	for i, size := range sizes {
		sizeURL := strings.Replace(baseURL, "4096", size, 1)
		sizeLinks[i] = fmt.Sprintf("[%s](%s)", size, sizeURL)
	}

	return strings.Join(sizeLinks, " • ")
}

func (c *AvatarCommand) getUserBannerURL(user *discordgo.User) string {
	if user.Banner == "" {
		return ""
	}

	extension := ".png"
	if strings.HasPrefix(user.Banner, "a_") {
		extension = ".gif"
	}

	return fmt.Sprintf("https://cdn.discordapp.com/banners/%s/%s%s?size=4096", 
		user.ID, user.Banner, extension)
}

func (c *AvatarCommand) getUserInfo(user *discordgo.User, member *discordgo.Member) string {
	info := []string{
		fmt.Sprintf("**ユーザー名:** %s", user.Username),
		fmt.Sprintf("**表示名:** %s", user.GlobalName),
	}

	if user.Discriminator != "" && user.Discriminator != "0" {
		info = append(info, fmt.Sprintf("**ディスクリミネータ:** #%s", user.Discriminator))
	}

	if user.Bot {
		info = append(info, "**タイプ:** 🤖 Bot")
	} else {
		info = append(info, "**タイプ:** 👤 ユーザー")
	}

	if member != nil && member.Nick != "" {
		info = append(info, fmt.Sprintf("**ニックネーム:** %s", member.Nick))
	}

	return strings.Join(info, "\n")
}