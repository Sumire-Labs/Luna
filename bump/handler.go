package bump

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/Sumire-Labs/Luna/database"
	"github.com/Sumire-Labs/Luna/embed"
)

const (
	DISBOARD_BOT_ID = "302050872383242240"
	BUMP_COOLDOWN   = 2 * time.Hour
)

type Handler struct {
	session *discordgo.Session
	db      *database.Service
}

func NewHandler(session *discordgo.Session, db *database.Service) *Handler {
	return &Handler{
		session: session,
		db:      db,
	}
}

// RegisterHandlers はbump関連のイベントハンドラーを登録します
func (h *Handler) RegisterHandlers() {
	h.session.AddHandler(h.onMessageCreate)
	h.session.AddHandler(h.onInteractionCreate)
}

// onMessageCreate はDISBOARDのbump成功メッセージを検知します
func (h *Handler) onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// BOT自身のメッセージは無視
	if m.Author.ID == s.State.User.ID {
		return
	}
	
	// DISBOARDからのメッセージかチェック
	if m.Author.ID != DISBOARD_BOT_ID {
		return
	}
	
	// Bump成功メッセージの検知（埋め込みメッセージをチェック）
	if len(m.Embeds) > 0 {
		embed := m.Embeds[0]
		// DISBOARDの成功メッセージパターン
		if embed.Image != nil && strings.Contains(embed.Description, "表示順をアップしたよ") ||
		   strings.Contains(embed.Description, "Bump done") ||
		   strings.Contains(embed.Description, "サーバーの表示順をアップ") {
			h.handleBumpSuccess(s, m.GuildID, m.ChannelID)
		}
	}
}

// handleBumpSuccess はbump成功時の処理を行います
func (h *Handler) handleBumpSuccess(s *discordgo.Session, guildID, channelID string) {
	// 設定を取得
	settings, err := h.db.GetGuildSettings(guildID)
	if err != nil || !settings.BumpEnabled {
		return
	}
	
	// Bump時刻を更新
	if err := h.db.UpdateBumpTime(guildID); err != nil {
		return
	}
	
	// 成功メッセージを送信
	successEmbed := embed.New().
		SetTitle("✅ Bump成功！").
		SetDescription("サーバーの表示順位が上がりました！").
		AddField("⏰ 次回Bump可能時刻", fmt.Sprintf("<t:%d:R>", time.Now().Add(BUMP_COOLDOWN).Unix()), true).
		AddField("🔔 リマインダー", "2時間後に通知します", true).
		SetColor(embed.M3Colors.Success).
		SetFooter("Luna Bump Tracker", "")
	
	// 通知チャンネルに送信
	notifyChannel := settings.BumpChannelID
	if notifyChannel == "" {
		notifyChannel = channelID
	}
	
	s.ChannelMessageSendEmbed(notifyChannel, successEmbed.Build())
	
	// 2時間後のリマインダーをスケジュール
	go h.scheduleBumpReminder(guildID, BUMP_COOLDOWN)
}

// scheduleBumpReminder は指定時間後にBumpリマインダーを送信します
func (h *Handler) scheduleBumpReminder(guildID string, duration time.Duration) {
	time.Sleep(duration)
	
	// 設定を再取得
	settings, err := h.db.GetGuildSettings(guildID)
	if err != nil || !settings.BumpEnabled {
		return
	}
	
	// 既にリマインダーが送信されている場合はスキップ
	if settings.BumpReminderSent {
		return
	}
	
	// リマインダーを送信
	reminderEmbed := embed.New().
		SetTitle("🔔 Bump可能になりました！").
		SetDescription("DISBOARDでサーバーをBumpできるようになりました！").
		AddField("📌 コマンド", "`/bump` を実行してください", false).
		SetColor(embed.M3Colors.Primary).
		SetFooter("Luna Bump Reminder", "").
		SetTimestamp()
	
	// メンション設定
	var content string
	if settings.BumpRoleID != "" {
		content = fmt.Sprintf("<@&%s>", settings.BumpRoleID)
	}
	
	// 通知チャンネルに送信
	if settings.BumpChannelID != "" {
		h.session.ChannelMessageSendComplex(settings.BumpChannelID, &discordgo.MessageSend{
			Content: content,
			Embed:   reminderEmbed.Build(),
		})
		
		// リマインダー送信済みフラグを更新
		h.db.MarkBumpReminderSent(guildID)
	}
}

// CheckPendingReminders は起動時に保留中のリマインダーをチェックします
func (h *Handler) CheckPendingReminders() {
	guilds, err := h.db.GetBumpableGuilds()
	if err != nil {
		return
	}
	
	for _, guild := range guilds {
		if guild.BumpLastTime != nil && !guild.BumpReminderSent {
			// 2時間経過しているか確認
			if time.Since(*guild.BumpLastTime) >= BUMP_COOLDOWN {
				// 即座にリマインダーを送信
				go h.scheduleBumpReminder(guild.GuildID, 0)
			} else {
				// 残り時間を計算してスケジュール
				remaining := BUMP_COOLDOWN - time.Since(*guild.BumpLastTime)
				go h.scheduleBumpReminder(guild.GuildID, remaining)
			}
		}
	}
}

// onInteractionCreate はslash commandのbump設定を処理します
func (h *Handler) onInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}
	
	customID := i.MessageComponentData().CustomID
	
	// Bump設定モーダル
	if customID == "config_main_bump" {
		h.showBumpSettingsModal(s, i)
	} else if customID == "config_bump_submit" {
		h.handleBumpSettingsSubmit(s, i)
	}
}

// showBumpSettingsModal はbump設定モーダルを表示します
func (h *Handler) showBumpSettingsModal(s *discordgo.Session, i *discordgo.InteractionCreate) {
	modal := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: "modal_bump_settings",
			Title:    "🔔 Bump通知設定",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "bump_channel",
							Label:       "通知チャンネル",
							Style:       discordgo.TextInputShort,
							Placeholder: "チャンネルIDを入力",
							Required:    true,
							MaxLength:   20,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "bump_role",
							Label:       "通知ロール（オプション）",
							Style:       discordgo.TextInputShort,
							Placeholder: "ロールIDを入力（省略可）",
							Required:    false,
							MaxLength:   20,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "bump_enabled",
							Label:       "機能を有効化",
							Style:       discordgo.TextInputShort,
							Placeholder: "true または false",
							Value:       "true",
							Required:    true,
							MaxLength:   5,
						},
					},
				},
			},
		},
	}
	
	s.InteractionRespond(i.Interaction, modal)
}

// handleBumpSettingsSubmit はbump設定の保存を処理します
func (h *Handler) handleBumpSettingsSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ModalSubmitData()
	
	// 現在の設定を取得
	settings, err := h.db.GetGuildSettings(i.GuildID)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ 設定の取得に失敗しました",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}
	
	// モーダルからのデータを処理
	for _, comp := range data.Components {
		row := comp.(*discordgo.ActionsRow)
		for _, c := range row.Components {
			input := c.(*discordgo.TextInput)
			switch input.CustomID {
			case "bump_channel":
				settings.BumpChannelID = input.Value
			case "bump_role":
				if input.Value != "" {
					settings.BumpRoleID = input.Value
				}
			case "bump_enabled":
				settings.BumpEnabled = strings.ToLower(input.Value) == "true"
			}
		}
	}
	
	// 設定を保存
	if err := h.db.UpsertGuildSettings(settings); err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ 設定の保存に失敗しました",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}
	
	// 成功メッセージ
	resultEmbed := embed.New().
		SetTitle("✅ Bump通知設定を更新しました").
		SetDescription("設定が正常に保存されました").
		AddField("📢 通知チャンネル", fmt.Sprintf("<#%s>", settings.BumpChannelID), true).
		SetColor(embed.M3Colors.Success)
	
	if settings.BumpRoleID != "" {
		resultEmbed.AddField("🔔 通知ロール", fmt.Sprintf("<@&%s>", settings.BumpRoleID), true)
	}
	
	resultEmbed.AddField("📌 状態", func() string {
		if settings.BumpEnabled {
			return "✅ 有効"
		}
		return "❌ 無効"
	}(), true)
	
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{resultEmbed.Build()},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}