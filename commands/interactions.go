package commands

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Sumire-Labs/Luna/config"
	"github.com/Sumire-Labs/Luna/database"
	"github.com/Sumire-Labs/Luna/embed"
	"github.com/bwmarrin/discordgo"
)

type InteractionHandler struct {
	session *discordgo.Session
	config  *config.Config
	db      *database.Service
}

func NewInteractionHandler(session *discordgo.Session, cfg *config.Config, db *database.Service) *InteractionHandler {
	return &InteractionHandler{
		session: session,
		config:  cfg,
		db:      db,
	}
}

func (h *InteractionHandler) HandleComponentInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}

	customID := i.MessageComponentData().CustomID
	log.Printf("HandleComponentInteraction called with customID: %s", customID)

	switch {
	// メインメニュー
	case customID == "config_main_tickets":
		h.handleTicketSetupStart(s, i)
	case customID == "config_main_moderation":
		h.handleModerationSetup(s, i)
	case customID == "config_main_welcome":
		h.handleWelcomeSetup(s, i)
	case customID == "config_main_logging":
		h.handleLoggingSetup(s, i)
	case customID == "config_main_bump":
		h.handleBumpConfig(s, i)
	case customID == "config_main_view":
		h.handleViewAllSettings(s, i)
	case customID == "config_main_reset":
		h.handleResetMenu(s, i)

	// チケット設定
	case customID == "ticket_setup_start":
		h.handleTicketSetupStart(s, i)
	case customID == "ticket_setup_panel":
		h.handleTicketPanelSetup(s, i)
	case customID == "ticket_setup_done":
		h.handleTicketSetupDone(s, i)
	case customID == "ticket_create":
		h.handleTicketCreate(s, i)
	case customID == "setup_cancel":
		h.handleSetupCancel(s, i)

	// 埋め込みビルダー - メインメニュー
	case customID == "embed_main_custom":
		h.handleEmbedCustomCreate(s, i)
	case customID == "embed_main_template":
		h.handleEmbedTemplateMenu(s, i)
	case customID == "embed_main_edit":
		h.handleEmbedEditRequest(s, i)
	case customID == "embed_main_help":
		h.handleEmbedHelp(s, i)
	case customID == "embed_main_colors":
		h.handleEmbedColorGuide(s, i)

	// 埋め込みビルダー - テンプレート
	case strings.HasPrefix(customID, "embed_template_"):
		h.handleEmbedTemplateSelect(s, i)
	case strings.HasPrefix(customID, "template_edit_"):
		h.handleTemplateEdit(s, i)
	case customID == "template_delete":
		h.handleTemplateDelete(s, i)

	// リセット確認
	case strings.HasPrefix(customID, "config_reset_confirm_"):
		feature := strings.TrimPrefix(customID, "config_reset_confirm_")
		h.handleResetConfirm(s, i, feature)
	case customID == "config_reset_cancel":
		h.handleResetCancel(s, i)

	// その他
	case strings.HasPrefix(customID, "ticket_setup_"):
		h.handleTicketSetupStep(s, i, customID)
	case strings.HasPrefix(customID, "ticket_close_"):
		h.handleTicketClose(s, i, customID)
	case strings.HasPrefix(customID, "ticket_transcript_"):
		h.handleTicketTranscript(s, i, customID)
	case strings.HasPrefix(customID, "ticket_close_confirm_"):
		h.handleTicketCloseConfirm(s, i, customID)
	case customID == "ticket_close_cancel":
		h.handleTicketCloseCancel(s, i)
	default:
		log.Printf("Unhandled customID: %s", customID)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("❌ 未対応のインタラクション: %s", customID),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}
}

func (h *InteractionHandler) handleModerationSetup(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "🚧 モデレーション設定は近日公開予定です！",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (h *InteractionHandler) handleWelcomeSetup(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "🚧 ウェルカムシステム設定は近日公開予定です！",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (h *InteractionHandler) handleLoggingSetup(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.Printf("handleLoggingSetup called for guild: %s, user: %s", i.GuildID, i.Member.User.ID)
	
	modal := discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: "logging_setup_modal",
			Title:    "📝 ログシステム設定",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "log_channel",
							Label:       "ログチャンネルID",
							Style:       discordgo.TextInputShort,
							Placeholder: "ログを送信するチャンネルのID",
							Required:    true,
							MaxLength:   20,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "log_description",
							Label:       "設定内容（自動設定・編集不要）",
							Style:       discordgo.TextInputParagraph,
							Placeholder: "すべてのログイベントが自動で有効になります",
							Required:    false,
							MaxLength:   200,
							Value:       "✅ メッセージ編集/削除 ✅ メンバー参加/退出 ✅ チャンネル/ロールイベント ✅ ボイス/モデレーション ✅ サーバーイベント ✅ ニックネーム変更",
						},
					},
				},
			},
		},
	}

	err := s.InteractionRespond(i.Interaction, &modal)
	if err != nil {
		log.Printf("Failed to respond to logging setup interaction: %v", err)
		// フォールバックレスポンス
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("❌ ログ設定モーダルの表示に失敗しました: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}
}

func (h *InteractionHandler) handleViewAllSettings(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guildID := i.GuildID
	settings, err := h.db.GetGuildSettings(guildID)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ 設定の読み込みに失敗しました！",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	embedBuilder := embed.New().
		SetTitle("📋 現在の設定状況").
		SetColor(embed.M3Colors.Info)

	// チケットシステム
	ticketStatus := "❌ 未設定"
	if settings.TicketEnabled {
		ticketStatus = "✅ 設定済み"
	}
	embedBuilder.AddField("🎫 チケットシステム", ticketStatus, true)

	// ログシステム
	logStatus := "❌ 未設定"
	if settings.LoggingEnabled {
		logStatus = "✅ 設定済み"
	}
	embedBuilder.AddField("📝 ログシステム", logStatus, true)

	// その他の機能
	embedBuilder.AddField("🛡️ モデレーション", "❌ 未設定", true)
	embedBuilder.AddField("👋 ウェルカム", "❌ 未設定", true)

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embedBuilder.Build()},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}

func (h *InteractionHandler) handleResetMenu(s *discordgo.Session, i *discordgo.InteractionCreate) {
	embedBuilder := embed.Warning(
		"⚠️ 設定リセット",
		"リセットする機能を選択してください\n\n**この操作は取り消せません！**",
	)

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Style:    discordgo.DangerButton,
					Label:    "🎫 チケット",
					CustomID: "config_reset_confirm_tickets",
				},
				discordgo.Button{
					Style:    discordgo.DangerButton,
					Label:    "🛡️ モデレーション",
					CustomID: "config_reset_confirm_moderation",
				},
				discordgo.Button{
					Style:    discordgo.DangerButton,
					Label:    "👋 ウェルカム",
					CustomID: "config_reset_confirm_welcome",
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Style:    discordgo.DangerButton,
					Label:    "📝 ログ",
					CustomID: "config_reset_confirm_logging",
				},
				discordgo.Button{
					Style:    discordgo.DangerButton,
					Label:    "🗑️ 全設定",
					CustomID: "config_reset_confirm_all",
				},
				discordgo.Button{
					Style:    discordgo.SecondaryButton,
					Label:    "❌ キャンセル",
					CustomID: "config_reset_cancel",
				},
			},
		},
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embedBuilder},
			Components: components,
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	})
}

func (h *InteractionHandler) handleTicketSetupStart(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Create modal for ticket setup
	modal := discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: "ticket_setup_modal",
			Title:    "🎫 チケットシステム設定",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "ticket_category",
							Label:       "チケットカテゴリID",
							Style:       discordgo.TextInputShort,
							Placeholder: "チケットチャンネルを作成するカテゴリのID",
							Required:    true,
							MaxLength:   20,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "support_role",
							Label:       "サポートロールID",
							Style:       discordgo.TextInputShort,
							Placeholder: "サポートスタッフのロールID（全チケット閲覧可能）",
							Required:    true,
							MaxLength:   20,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "admin_role",
							Label:       "管理者ロールID（任意）",
							Style:       discordgo.TextInputShort,
							Placeholder: "チケット管理者のロールID",
							Required:    false,
							MaxLength:   20,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "log_channel",
							Label:       "ログチャンネルID（任意）",
							Style:       discordgo.TextInputShort,
							Placeholder: "チケットイベントを記録するチャンネルID",
							Required:    false,
							MaxLength:   20,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "auto_close_hours",
							Label:       "自動クローズ時間（デフォルト: 24）",
							Style:       discordgo.TextInputShort,
							Placeholder: "非アクティブチケットの自動クローズまでの時間（0で無効）",
							Required:    false,
							MaxLength:   3,
						},
					},
				},
			},
		},
	}

	s.InteractionRespond(i.Interaction, &modal)
}

func (h *InteractionHandler) HandleModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionModalSubmit {
		return
	}

	data := i.ModalSubmitData()

	switch {
	case data.CustomID == "ticket_setup_modal":
		h.handleTicketSetupModal(s, i)
	case data.CustomID == "ticket_create_modal":
		h.handleTicketCreateModal(s, i)
	case data.CustomID == "logging_setup_modal":
		h.handleLoggingSetupModal(s, i)
	case data.CustomID == "embed_create_modal":
		h.handleEmbedCreateModal(s, i)
	case strings.HasPrefix(data.CustomID, "embed_edit_modal_"):
		h.handleEmbedEditModal(s, i)
	case strings.HasPrefix(data.CustomID, "template_edit_modal_"):
		h.handleTemplateEditModal(s, i)
	case data.CustomID == "embed_edit_request_modal":
		h.handleEmbedEditRequestModal(s, i)
	case data.CustomID == "modal_bump_settings":
		h.handleBumpSettingsSubmit(s, i)
	}
}

func (h *InteractionHandler) handleTicketSetupModal(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ModalSubmitData()
	guildID := i.GuildID

	// Extract form data
	var categoryID, supportRoleID, adminRoleID, logChannelID string
	var autoCloseHours = 24

	for _, component := range data.Components {
		for _, comp := range component.(*discordgo.ActionsRow).Components {
			textInput := comp.(*discordgo.TextInput)
			value := textInput.Value

			switch textInput.CustomID {
			case "ticket_category":
				categoryID = value
			case "support_role":
				supportRoleID = value
			case "admin_role":
				adminRoleID = value
			case "log_channel":
				logChannelID = value
			case "auto_close_hours":
				if value != "" {
					fmt.Sscanf(value, "%d", &autoCloseHours)
				}
			}
		}
	}

	// Validate required fields
	if categoryID == "" || supportRoleID == "" {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ カテゴリIDとサポートロールIDは必須です！",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// Validate IDs exist
	if err := h.validateTicketSetup(guildID, categoryID, supportRoleID, adminRoleID, logChannelID); err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("❌ 検証に失敗しました: %s", err.Error()),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// ギルドが存在しない場合は先に登録
	guild, err := s.Guild(guildID)
	if err == nil {
		_ = h.db.UpsertGuild(guildID, guild.Name, "/")
	}

	// Get current settings
	settings, err := h.db.GetGuildSettings(guildID)
	if err != nil {
		// エラーでもデフォルト設定を作成
		settings = &database.GuildSettings{
			GuildID:              guildID,
			TicketAutoCloseHours: 24,
			TicketMaxPerUser:     3,
			LogMessageEdits:      true,
			LogMessageDeletes:    true,
			LogMemberJoins:       true,
			LogMemberLeaves:     true,
			LogChannelEvents:     false,
			LogRoleEvents:        false,
			LogVoiceEvents:       false,
			LogModerationEvents:  false,
			LogServerEvents:      false,
			LogNicknameChanges:   false,
		}
	}

	// Update ticket settings
	settings.TicketEnabled = true
	settings.TicketCategoryID = categoryID
	settings.TicketSupportRoleID = supportRoleID
	settings.TicketAdminRoleID = adminRoleID
	settings.TicketLogChannelID = logChannelID
	settings.TicketAutoCloseHours = autoCloseHours

	// Debug log
	log.Printf("Saving ticket settings for guild %s: %+v", guildID, settings)

	// Save settings
	if err := h.db.UpsertGuildSettings(settings); err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("❌ 設定の保存に失敗しました！エラー: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// Create success embed
	embedBuilder := embed.New().
		SetTitle("✅ チケットシステム設定完了！").
		SetDescription("チケットシステムが正常に設定されました。").
		SetColor(embed.M3Colors.Success)

	// Add configuration details
	embedBuilder.AddField("📁 カテゴリ", fmt.Sprintf("<#%s>", categoryID), true)
	embedBuilder.AddField("🛡️ サポートロール", fmt.Sprintf("<@&%s>", supportRoleID), true)

	if adminRoleID != "" {
		embedBuilder.AddField("👑 管理者ロール", fmt.Sprintf("<@&%s>", adminRoleID), true)
	}

	if logChannelID != "" {
		embedBuilder.AddField("📝 ログチャンネル", fmt.Sprintf("<#%s>", logChannelID), true)
	}

	embedBuilder.AddField("⏰ 自動クローズ", fmt.Sprintf("%d時間", autoCloseHours), true)

	// チケットパネル設置ボタンを追加
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "🎫 チケットパネルを設置",
					Style:    discordgo.PrimaryButton,
					CustomID: "ticket_setup_panel",
				},
				discordgo.Button{
					Label:    "✅ 完了",
					Style:    discordgo.SecondaryButton,
					CustomID: "ticket_setup_done",
				},
			},
		},
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embedBuilder.Build()},
			Components: components,
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	})
}

func (h *InteractionHandler) handleLoggingSetupModal(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ModalSubmitData()
	guildID := i.GuildID

	var logChannelID string

	for _, component := range data.Components {
		for _, comp := range component.(*discordgo.ActionsRow).Components {
			textInput := comp.(*discordgo.TextInput)
			value := textInput.Value

			switch textInput.CustomID {
			case "log_channel":
				logChannelID = value
			// log_description は無視（参考用なので）
			}
		}
	}

	// 必須フィールドの検証
	if logChannelID == "" {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ ログチャンネルIDは必須です！",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// チャンネルの存在確認
	channel, err := s.Channel(logChannelID)
	if err != nil || channel.Type != discordgo.ChannelTypeGuildText {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ 指定されたチャンネルが見つからないか、テキストチャンネルではありません！",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// ギルドが存在しない場合は先に登録
	guild, err := s.Guild(guildID)
	if err == nil {
		_ = h.db.UpsertGuild(guildID, guild.Name, "/")
	}

	// 現在の設定を取得
	settings, err := h.db.GetGuildSettings(guildID)
	if err != nil {
		// エラーでもデフォルト設定を作成
		settings = &database.GuildSettings{
			GuildID:              guildID,
			LogMessageEdits:      true,
			LogMessageDeletes:    true,
			LogMemberJoins:       true,
			LogMemberLeaves:      true,
			LogChannelEvents:     false,
			LogRoleEvents:        false,
			LogVoiceEvents:       false,
			LogModerationEvents:  false,
			LogServerEvents:      false,
			LogNicknameChanges:   false,
		}
	}

	// ログ設定を更新
	settings.LoggingEnabled = true
	settings.LogChannelID = logChannelID

	// すべてのログイベントを自動で有効にする
	settings.LogMessageEdits = true
	settings.LogMessageDeletes = true
	settings.LogMemberJoins = true
	settings.LogMemberLeaves = true
	settings.LogChannelEvents = true
	settings.LogRoleEvents = true
	settings.LogVoiceEvents = true
	settings.LogModerationEvents = true
	settings.LogServerEvents = true
	settings.LogNicknameChanges = true

	// 設定を保存
	if err := h.db.UpsertGuildSettings(settings); err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ 設定の保存に失敗しました！",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// 成功メッセージを作成
	embedBuilder := embed.New().
		SetTitle("✅ ログシステム設定完了！").
		SetDescription("すべてのログイベントが自動で有効になりました。").
		SetColor(embed.M3Colors.Success).
		AddField("📍 ログチャンネル", fmt.Sprintf("<#%s>", logChannelID), false).
		AddField("📋 有効なイベント", strings.Join([]string{
			"✅ メッセージ編集/削除",
			"✅ メンバー参加/退出", 
			"✅ チャンネル/ロールイベント",
			"✅ ボイス/モデレーションイベント",
			"✅ サーバーイベント/ニックネーム変更",
		}, "\n"), false).
		AddField("💡 使用方法", strings.Join([]string{
			"• 指定したログチャンネルで自動記録開始",
			"• `/config` でリセットや再設定が可能",
			"• すべてのサーバーアクティビティを網羅",
		}, "\n"), false)

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embedBuilder.Build()},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}

func (h *InteractionHandler) validateTicketSetup(guildID, categoryID, supportRoleID, adminRoleID, logChannelID string) error {
	// Validate category
	if categoryID != "" {
		channels, err := h.session.GuildChannels(guildID)
		if err != nil {
			return fmt.Errorf("チャンネルの取得に失敗しました")
		}

		categoryExists := false
		for _, channel := range channels {
			if channel.ID == categoryID && channel.Type == discordgo.ChannelTypeGuildCategory {
				categoryExists = true
				break
			}
		}

		if !categoryExists {
			return fmt.Errorf("カテゴリが見つからないか、カテゴリチャンネルではありません")
		}
	}

	// Validate support role
	if supportRoleID != "" {
		roles, err := h.session.GuildRoles(guildID)
		if err != nil {
			return fmt.Errorf("ロールの取得に失敗しました")
		}

		roleExists := false
		for _, role := range roles {
			if role.ID == supportRoleID {
				roleExists = true
				break
			}
		}

		if !roleExists {
			return fmt.Errorf("サポートロールが見つかりません")
		}
	}

	// Validate admin role (optional)
	if adminRoleID != "" {
		roles, err := h.session.GuildRoles(guildID)
		if err != nil {
			return fmt.Errorf("ロールの取得に失敗しました")
		}

		roleExists := false
		for _, role := range roles {
			if role.ID == adminRoleID {
				roleExists = true
				break
			}
		}

		if !roleExists {
			return fmt.Errorf("管理者ロールが見つかりません")
		}
	}

	// Validate log channel (optional)
	if logChannelID != "" {
		channels, err := h.session.GuildChannels(guildID)
		if err != nil {
			return fmt.Errorf("チャンネルの取得に失敗しました")
		}

		channelExists := false
		for _, channel := range channels {
			if channel.ID == logChannelID && channel.Type == discordgo.ChannelTypeGuildText {
				channelExists = true
				break
			}
		}

		if !channelExists {
			return fmt.Errorf("ログチャンネルが見つからないか、テキストチャンネルではありません")
		}
	}

	return nil
}

func (h *InteractionHandler) handleSetupCancel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    "❌ セットアップをキャンセルしました。",
			Embeds:     []*discordgo.MessageEmbed{},
			Components: []discordgo.MessageComponent{},
		},
	})
}

func (h *InteractionHandler) handleResetConfirm(s *discordgo.Session, i *discordgo.InteractionCreate, feature string) {
	guildID := i.GuildID

	// Reset the feature
	if err := h.db.ResetGuildSettings(guildID, feature); err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Content:    fmt.Sprintf("❌ %s設定のリセットに失敗しました！", h.getFeatureName(feature)),
				Embeds:     []*discordgo.MessageEmbed{},
				Components: []discordgo.MessageComponent{},
			},
		})
		return
	}

	// Create success message
	featureName := h.getFeatureName(feature)
	embedBuilder := embed.Success(
		"✅ 設定リセット完了",
		fmt.Sprintf("**%s**の設定が正常にリセットされました。", featureName),
	)

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embedBuilder},
			Components: []discordgo.MessageComponent{},
		},
	})
}

func (h *InteractionHandler) handleResetCancel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    "❌ リセットをキャンセルしました。",
			Embeds:     []*discordgo.MessageEmbed{},
			Components: []discordgo.MessageComponent{},
		},
	})
}

func (h *InteractionHandler) handleTicketSetupStep(s *discordgo.Session, i *discordgo.InteractionCreate, customID string) {
	// Handle different ticket setup steps
	// This can be expanded for multi-step setup processes
}

func (h *InteractionHandler) getFeatureName(feature string) string {
	names := map[string]string{
		"tickets":    "🎫 Ticket System",
		"moderation": "🛡️ Moderation",
		"welcome":    "👋 Welcome System",
		"logging":    "📝 Logging",
		"all":        "🔄 All Settings",
	}

	if name, ok := names[feature]; ok {
		return name
	}
	return feature
}

// 埋め込みビルダー関連のハンドラー

func (h *InteractionHandler) handleEmbedCustomCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	modal := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: "embed_create_modal",
			Title:    "📝 埋め込み作成",
			Components: []discordgo.MessageComponent{
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.TextInput{
							CustomID:    "embed_title",
							Label:       "タイトル",
							Style:       discordgo.TextInputShort,
							Placeholder: "埋め込みのタイトルを入力...",
							Required:    false,
							MaxLength:   256,
						},
					},
				},
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.TextInput{
							CustomID:    "embed_description",
							Label:       "説明",
							Style:       discordgo.TextInputParagraph,
							Placeholder: "埋め込みの説明を入力...",
							Required:    false,
							MaxLength:   4000,
						},
					},
				},
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.TextInput{
							CustomID:    "embed_color",
							Label:       "カラー (16進数 例: #6750A4 または 0x6750A4)",
							Style:       discordgo.TextInputShort,
							Placeholder: "#6750A4",
							Required:    false,
						},
					},
				},
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.TextInput{
							CustomID:    "embed_image",
							Label:       "画像URL (オプション)",
							Style:       discordgo.TextInputShort,
							Placeholder: "https://example.com/image.png",
							Required:    false,
						},
					},
				},
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.TextInput{
							CustomID:    "embed_footer",
							Label:       "フッター (オプション)",
							Style:       discordgo.TextInputShort,
							Placeholder: "フッターテキスト",
							Required:    false,
							MaxLength:   2048,
						},
					},
				},
			},
		},
	}

	s.InteractionRespond(i.Interaction, modal)
}

func (h *InteractionHandler) handleEmbedTemplateMenu(s *discordgo.Session, i *discordgo.InteractionCreate) {
	embedBuilder := embed.New().
		SetTitle("📋 テンプレート選択").
		SetDescription("使用するテンプレートを選択してください").
		SetColor(embed.M3Colors.Secondary).
		AddField("📢 お知らせ", "重要な告知用テンプレート", true).
		AddField("📋 ルール", "サーバールール用テンプレート", true).
		AddField("❓ FAQ", "よくある質問用テンプレート", true).
		AddField("🎉 イベント", "イベント告知用テンプレート", true).
		AddField("⚠️ 警告", "重要な警告用テンプレート", true).
		AddBlankField(true).
		SetFooter("テンプレートを選択してください", "")

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Style:    discordgo.PrimaryButton,
					Label:    "📢 お知らせ",
					CustomID: "embed_template_announcement",
				},
				discordgo.Button{
					Style:    discordgo.SecondaryButton,
					Label:    "📋 ルール",
					CustomID: "embed_template_rules",
				},
				discordgo.Button{
					Style:    discordgo.SecondaryButton,
					Label:    "❓ FAQ",
					CustomID: "embed_template_faq",
				},
			},
		},
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Style:    discordgo.SuccessButton,
					Label:    "🎉 イベント",
					CustomID: "embed_template_event",
				},
				discordgo.Button{
					Style:    discordgo.DangerButton,
					Label:    "⚠️ 警告",
					CustomID: "embed_template_warning",
				},
			},
		},
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embedBuilder.Build()},
			Components: components,
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	})
}

func (h *InteractionHandler) handleEmbedEditRequest(s *discordgo.Session, i *discordgo.InteractionCreate) {
	modal := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: "embed_edit_request_modal",
			Title:    "✏️ 埋め込み編集",
			Components: []discordgo.MessageComponent{
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.TextInput{
							CustomID:    "message_id",
							Label:       "編集するメッセージID",
							Style:       discordgo.TextInputShort,
							Placeholder: "123456789012345678",
							Required:    true,
							MinLength:   17,
							MaxLength:   20,
						},
					},
				},
			},
		},
	}

	s.InteractionRespond(i.Interaction, modal)
}

func (h *InteractionHandler) handleEmbedHelp(s *discordgo.Session, i *discordgo.InteractionCreate) {
	helpEmbed := embed.New().
		SetTitle("📚 埋め込みビルダー ヘルプ").
		SetDescription("埋め込みビルダーの使用方法を説明します").
		SetColor(embed.M3Colors.Info).
		AddField("🎨 カスタム作成", "自由にタイトル、説明、カラーなどを設定できます", false).
		AddField("📋 テンプレート", "事前定義されたデザインから選択できます", false).
		AddField("✏️ 編集機能", "既存の埋め込みメッセージを編集できます", false).
		AddField("🎨 カラーコード", "#6750A4 または 0x6750A4 の形式で指定", false).
		AddField("🖼️ 画像URL", "https:// で始まる画像URLを指定可能", false).
		AddField("⚠️ 制限事項", "タイトル: 256文字、説明: 4000文字、フッター: 2048文字", false).
		SetFooter("困ったときはサポートチャンネルへ", "")

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{helpEmbed.Build()},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}

func (h *InteractionHandler) handleEmbedColorGuide(s *discordgo.Session, i *discordgo.InteractionCreate) {
	colorEmbed := embed.New().
		SetTitle("🎨 カラーガイド").
		SetDescription("利用可能なカラーコードの例です").
		SetColor(embed.M3Colors.Primary).
		AddField("🟣 Primary", "#6750A4 (0x6750A4)", true).
		AddField("🟤 Secondary", "#625B71 (0x625B71)", true).
		AddField("🟫 Tertiary", "#7D5260 (0x7D5260)", true).
		AddField("🔴 Error", "#BA1A1A (0xBA1A1A)", true).
		AddField("🟢 Success", "#4CAF50 (0x4CAF50)", true).
		AddField("🟠 Warning", "#FF9800 (0xFF9800)", true).
		AddField("🔵 Info", "#2196F3 (0x2196F3)", true).
		AddField("⚫ Surface", "#1C1B1F (0x1C1B1F)", true).
		AddField("⭐ カスタム例", "#FF69B4, #00CED1, #FFD700 など", false).
		SetFooter("# または 0x プレフィックス付きで入力", "")

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{colorEmbed.Build()},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}

func (h *InteractionHandler) handleEmbedTemplateSelect(s *discordgo.Session, i *discordgo.InteractionCreate) {
	templateType := strings.TrimPrefix(i.MessageComponentData().CustomID, "embed_template_")

	var embedBuilder *embed.Builder

	switch templateType {
	case "announcement":
		embedBuilder = embed.New().
			SetTitle("📢 重要なお知らせ").
			SetDescription("ここにお知らせ内容を記入してください。").
			SetColor(embed.M3Colors.Info).
			AddField("📅 日時", "YYYY/MM/DD HH:MM", true).
			AddField("👤 投稿者", "管理者", true).
			AddField("🔗 詳細", "詳細情報がある場合はここに", false)

	case "rules":
		embedBuilder = embed.New().
			SetTitle("📋 サーバールール").
			SetDescription("このサーバーを快適に利用するためのルールです。").
			SetColor(embed.M3Colors.Primary).
			AddField("1️⃣ 基本的なマナー", "他の参加者を尊重し、礼儀正しく行動してください。", false).
			AddField("2️⃣ スパム禁止", "不要なメッセージの連投は禁止です。", false).
			AddField("3️⃣ 適切なチャンネル使用", "各チャンネルの目的に沿った投稿をしてください。", false).
			SetFooter("ルール違反には警告・キック・BANの対象となります", "")

	case "faq":
		embedBuilder = embed.New().
			SetTitle("❓ よくある質問").
			SetDescription("頻繁にお問い合わせいただく質問をまとめました。").
			SetColor(embed.M3Colors.Info).
			AddField("Q1: ○○はどうすればいいですか？", "A1: ○○の方法について説明...", false).
			AddField("Q2: ○○ができません", "A2: ○○の対処法について...", false).
			AddField("Q3: その他の質問", "A3: サポートチャンネルでお気軽にお尋ねください", false)

	case "event":
		embedBuilder = embed.New().
			SetTitle("🎉 イベント開催のお知らせ").
			SetDescription("楽しいイベントを開催します！ぜひご参加ください。").
			SetColor(embed.M3Colors.Success).
			AddField("📅 開催日時", "YYYY/MM/DD HH:MM〜", true).
			AddField("📍 場所", "○○チャンネル", true).
			AddField("🎯 参加条件", "特になし（どなたでも参加可能）", false).
			AddField("🏆 景品", "参加者全員にプレゼント！", false).
			SetFooter("参加表明は下のボタンをクリック", "")

	case "warning":
		embedBuilder = embed.New().
			SetTitle("⚠️ 重要な警告").
			SetDescription("緊急かつ重要な情報です。必ずお読みください。").
			SetColor(embed.M3Colors.Warning).
			AddField("🚨 警告内容", "具体的な警告内容をここに記載", false).
			AddField("📋 対処方法", "推奨される対処方法について", false).
			AddField("📞 お問い合わせ", "不明な点があれば管理者までご連絡ください", false).
			SetFooter("この警告を確認したら反応してください", "")

	default:
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ 不明なテンプレートタイプです",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// テンプレート埋め込みを送信
	response := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embedBuilder.Build()},
			Components: []discordgo.MessageComponent{
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.Button{
							Style:    discordgo.SecondaryButton,
							Label:    "✏️ 編集",
							CustomID: fmt.Sprintf("template_edit_%s", templateType),
						},
						&discordgo.Button{
							Style:    discordgo.DangerButton,
							Label:    "🗑️ 削除",
							CustomID: "template_delete",
						},
					},
				},
			},
		},
	}

	s.InteractionRespond(i.Interaction, response)
}

func (h *InteractionHandler) handleEmbedCreateModal(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ModalSubmitData()

	var title, description, colorStr, imageURL, footer string

	for _, component := range data.Components {
		for _, comp := range component.(*discordgo.ActionsRow).Components {
			textInput := comp.(*discordgo.TextInput)
			value := textInput.Value

			switch textInput.CustomID {
			case "embed_title":
				title = value
			case "embed_description":
				description = value
			case "embed_color":
				colorStr = value
			case "embed_image":
				imageURL = value
			case "embed_footer":
				footer = value
			}
		}
	}

	// 埋め込みを構築
	embedBuilder := embed.New()

	if title != "" {
		embedBuilder.SetTitle(title)
	}

	if description != "" {
		embedBuilder.SetDescription(description)
	}

	// カラーを解析
	if colorStr != "" {
		if color, err := parseColor(colorStr); err == nil {
			embedBuilder.SetColor(color)
		}
	}

	if imageURL != "" {
		embedBuilder.SetImage(imageURL)
	}

	if footer != "" {
		embedBuilder.SetFooter(footer, "")
	}

	// 埋め込みを送信
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embedBuilder.Build()},
			Components: []discordgo.MessageComponent{
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.Button{
							Style:    discordgo.SecondaryButton,
							Label:    "✏️ 編集",
							CustomID: "embed_edit_request",
						},
						&discordgo.Button{
							Style:    discordgo.DangerButton,
							Label:    "🗑️ 削除",
							CustomID: "embed_delete",
						},
					},
				},
			},
		},
	})
}

func (h *InteractionHandler) handleEmbedEditModal(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ModalSubmitData()
	messageID := strings.TrimPrefix(data.CustomID, "embed_edit_modal_")

	var title, description, colorStr, imageURL, footer string

	for _, component := range data.Components {
		for _, comp := range component.(*discordgo.ActionsRow).Components {
			textInput := comp.(*discordgo.TextInput)
			value := textInput.Value

			switch textInput.CustomID {
			case "embed_title":
				title = value
			case "embed_description":
				description = value
			case "embed_color":
				colorStr = value
			case "embed_image":
				imageURL = value
			case "embed_footer":
				footer = value
			}
		}
	}

	// 埋め込みを構築
	embedBuilder := embed.New()

	if title != "" {
		embedBuilder.SetTitle(title)
	}

	if description != "" {
		embedBuilder.SetDescription(description)
	}

	// カラーを解析
	if colorStr != "" {
		if color, err := parseColor(colorStr); err == nil {
			embedBuilder.SetColor(color)
		}
	}

	if imageURL != "" {
		embedBuilder.SetImage(imageURL)
	}

	if footer != "" {
		embedBuilder.SetFooter(footer, "")
	}

	// メッセージを編集
	_, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Channel: i.ChannelID,
		ID:      messageID,
		Embeds:  &[]*discordgo.MessageEmbed{embedBuilder.Build()},
	})

	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ 埋め込みの編集に失敗しました",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "✅ 埋め込みを正常に編集しました！",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (h *InteractionHandler) handleTemplateEdit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	templateType := strings.TrimPrefix(i.MessageComponentData().CustomID, "template_edit_")

	// 現在のメッセージから埋め込み情報を取得
	if len(i.Message.Embeds) == 0 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ 編集可能な埋め込みが見つかりません",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	currentEmbed := i.Message.Embeds[0]

	modal := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: fmt.Sprintf("template_edit_modal_%s_%s", templateType, i.Message.ID),
			Title:    "✏️ テンプレート編集",
			Components: []discordgo.MessageComponent{
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.TextInput{
							CustomID:    "embed_title",
							Label:       "タイトル",
							Style:       discordgo.TextInputShort,
							Placeholder: "埋め込みのタイトルを入力...",
							Required:    false,
							MaxLength:   256,
							Value:       currentEmbed.Title,
						},
					},
				},
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.TextInput{
							CustomID:    "embed_description",
							Label:       "説明",
							Style:       discordgo.TextInputParagraph,
							Placeholder: "埋め込みの説明を入力...",
							Required:    false,
							MaxLength:   4000,
							Value:       currentEmbed.Description,
						},
					},
				},
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.TextInput{
							CustomID:    "embed_color",
							Label:       "カラー (16進数)",
							Style:       discordgo.TextInputShort,
							Placeholder: "#6750A4",
							Required:    false,
							Value:       fmt.Sprintf("#%06X", currentEmbed.Color),
						},
					},
				},
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.TextInput{
							CustomID:    "embed_footer",
							Label:       "フッター",
							Style:       discordgo.TextInputShort,
							Placeholder: "フッターテキスト",
							Required:    false,
							MaxLength:   2048,
							Value:       getFooterText(currentEmbed),
						},
					},
				},
			},
		},
	}

	s.InteractionRespond(i.Interaction, modal)
}

func (h *InteractionHandler) handleTemplateEditModal(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ModalSubmitData()
	parts := strings.Split(data.CustomID, "_")
	if len(parts) < 5 {
		return
	}
	messageID := parts[len(parts)-1]

	var title, description, colorStr, footer string

	for _, component := range data.Components {
		for _, comp := range component.(*discordgo.ActionsRow).Components {
			textInput := comp.(*discordgo.TextInput)
			value := textInput.Value

			switch textInput.CustomID {
			case "embed_title":
				title = value
			case "embed_description":
				description = value
			case "embed_color":
				colorStr = value
			case "embed_footer":
				footer = value
			}
		}
	}

	// 元の埋め込みを取得
	originalMessage, err := s.ChannelMessage(i.ChannelID, messageID)
	if err != nil || len(originalMessage.Embeds) == 0 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ 元のメッセージが見つかりません",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// 埋め込みを構築（元のフィールドは保持）
	embedBuilder := embed.New()

	if title != "" {
		embedBuilder.SetTitle(title)
	}

	if description != "" {
		embedBuilder.SetDescription(description)
	}

	// カラーを解析
	if colorStr != "" {
		if color, err := parseColor(colorStr); err == nil {
			embedBuilder.SetColor(color)
		}
	}

	// 元のフィールドを復元
	for _, field := range originalMessage.Embeds[0].Fields {
		embedBuilder.AddField(field.Name, field.Value, field.Inline)
	}

	if footer != "" {
		embedBuilder.SetFooter(footer, "")
	}

	// メッセージを編集
	_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Channel: i.ChannelID,
		ID:      messageID,
		Embeds:  &[]*discordgo.MessageEmbed{embedBuilder.Build()},
	})

	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ テンプレートの編集に失敗しました",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "✅ テンプレートを正常に編集しました！",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (h *InteractionHandler) handleTemplateDelete(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.ChannelMessageDelete(i.ChannelID, i.Message.ID)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ メッセージの削除に失敗しました",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "🗑️ 埋め込みを削除しました",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (h *InteractionHandler) handleEmbedEditRequestModal(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ModalSubmitData()

	var messageID string
	for _, component := range data.Components {
		for _, comp := range component.(*discordgo.ActionsRow).Components {
			textInput := comp.(*discordgo.TextInput)
			if textInput.CustomID == "message_id" {
				messageID = textInput.Value
				break
			}
		}
	}

	// メッセージを取得して編集可能かチェック
	message, err := s.ChannelMessage(i.ChannelID, messageID)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ 指定されたメッセージが見つかりません",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	if message.Author.ID != s.State.User.ID {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ このボットが作成したメッセージのみ編集できます",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	if len(message.Embeds) == 0 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ 指定されたメッセージには埋め込みがありません",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	currentEmbed := message.Embeds[0]

	modal := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: fmt.Sprintf("embed_edit_modal_%s", messageID),
			Title:    "✏️ 埋め込み編集",
			Components: []discordgo.MessageComponent{
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.TextInput{
							CustomID:    "embed_title",
							Label:       "タイトル",
							Style:       discordgo.TextInputShort,
							Placeholder: "埋め込みのタイトルを入力...",
							Required:    false,
							MaxLength:   256,
							Value:       currentEmbed.Title,
						},
					},
				},
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.TextInput{
							CustomID:    "embed_description",
							Label:       "説明",
							Style:       discordgo.TextInputParagraph,
							Placeholder: "埋め込みの説明を入力...",
							Required:    false,
							MaxLength:   4000,
							Value:       currentEmbed.Description,
						},
					},
				},
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.TextInput{
							CustomID:    "embed_color",
							Label:       "カラー (16進数 例: #6750A4 または 0x6750A4)",
							Style:       discordgo.TextInputShort,
							Placeholder: "#6750A4",
							Required:    false,
							Value:       fmt.Sprintf("#%06X", currentEmbed.Color),
						},
					},
				},
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.TextInput{
							CustomID:    "embed_image",
							Label:       "画像URL (オプション)",
							Style:       discordgo.TextInputShort,
							Placeholder: "https://example.com/image.png",
							Required:    false,
							Value:       getImageURL(currentEmbed),
						},
					},
				},
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.TextInput{
							CustomID:    "embed_footer",
							Label:       "フッター (オプション)",
							Style:       discordgo.TextInputShort,
							Placeholder: "フッターテキスト",
							Required:    false,
							MaxLength:   2048,
							Value:       getFooterText(currentEmbed),
						},
					},
				},
			},
		},
	}

	s.InteractionRespond(i.Interaction, modal)
}

// ヘルパー関数
func parseColor(colorStr string) (int, error) {

	colorStr = strings.TrimSpace(colorStr)

	// # で始まる場合は除去
	if strings.HasPrefix(colorStr, "#") {
		colorStr = colorStr[1:]
	}

	// 0x で始まる場合は除去
	if strings.HasPrefix(strings.ToLower(colorStr), "0x") {
		colorStr = colorStr[2:]
	}

	// 16進数として解析
	color, err := strconv.ParseInt(colorStr, 16, 32)
	if err != nil {
		return embed.M3Colors.Primary, err
	}

	return int(color), nil
}

// getFooterText はembedからフッターテキストを取得します
func getFooterText(embed *discordgo.MessageEmbed) string {
	if embed.Footer != nil {
		return embed.Footer.Text
	}
	return ""
}

// getImageURL はembedから画像URLを取得します
func getImageURL(embed *discordgo.MessageEmbed) string {
	if embed.Image != nil {
		return embed.Image.URL
	}
	return ""
}

// handleBumpConfig はBump通知設定を処理します
func (h *InteractionHandler) handleBumpConfig(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: "modal_bump_settings",
			Title:    "🔔 Bump通知設定",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "bump_channel",
							Label:       "通知チャンネルID",
							Style:       discordgo.TextInputShort,
							Placeholder: "例: 1234567890123456789",
							Required:    true,
							MaxLength:   20,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "bump_role",
							Label:       "通知ロールID（任意）",
							Style:       discordgo.TextInputShort,
							Placeholder: "例: 1234567890123456789（空欄可）",
							Required:    false,
							MaxLength:   20,
						},
					},
				},
			},
		},
	})
}

// handleBumpSettingsSubmit はBump設定のモーダル送信を処理します
func (h *InteractionHandler) handleBumpSettingsSubmit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ModalSubmitData()
	guildID := i.GuildID
	
	// ギルドが存在しない場合は先に登録
	guild, err := s.Guild(guildID)
	if err == nil {
		_ = h.db.UpsertGuild(guildID, guild.Name, "/")
	}

	// 現在の設定を取得
	settings, err := h.db.GetGuildSettings(guildID)
	if err != nil {
		// エラーでもデフォルト設定を作成
		settings = &database.GuildSettings{
			GuildID:              guildID,
			TicketAutoCloseHours: 24,
			TicketMaxPerUser:     3,
			LogMessageEdits:      true,
			LogMessageDeletes:    true,
			LogMemberJoins:       true,
			LogMemberLeaves:      true,
			LogChannelEvents:     false,
			LogRoleEvents:        false,
			LogVoiceEvents:       false,
			LogModerationEvents:  false,
			LogServerEvents:      false,
			LogNicknameChanges:   false,
		}
	}
	
	// フォームデータを処理
	for _, row := range data.Components {
		if r, ok := row.(*discordgo.ActionsRow); ok {
			for _, comp := range r.Components {
				if input, ok := comp.(*discordgo.TextInput); ok {
					switch input.CustomID {
					case "bump_channel":
						settings.BumpChannelID = input.Value
					case "bump_role":
						settings.BumpRoleID = input.Value
					}
				}
			}
		}
	}
	
	// Bump機能を有効化
	settings.BumpEnabled = true
	
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
	embedBuilder := embed.New().
		SetTitle("✅ Bump通知設定完了").
		SetDescription("DISBOARD Bump通知機能を設定しました！").
		SetColor(embed.M3Colors.Success).
		AddField("📢 通知チャンネル", fmt.Sprintf("<#%s>", settings.BumpChannelID), true)
	
	if settings.BumpRoleID != "" {
		embedBuilder.AddField("🔔 通知ロール", fmt.Sprintf("<@&%s>", settings.BumpRoleID), true)
	}
	
	embedBuilder.AddField("📌 使い方", 
		"DISBOARDで `/bump` を実行すると、2時間後に自動で通知が送信されます", false)
	
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embedBuilder.Build()},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})
}

func (h *InteractionHandler) handleTicketPanelSetup(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guildID := i.GuildID
	channelID := i.ChannelID

	// チケット設定を確認
	settings, err := h.db.GetGuildSettings(guildID)
	if err != nil || !settings.TicketEnabled {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ チケットシステムが設定されていません！先に設定を完了してください。",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// チケット作成パネルを作成
	panelEmbed := embed.New().
		SetTitle("🎫 サポートチケット").
		SetDescription("サポートが必要な場合は、下のボタンをクリックしてチケットを作成してください。").
		SetColor(embed.M3Colors.Primary).
		AddField("📋 利用方法", strings.Join([]string{
			"1. 「📧 チケット作成」ボタンをクリック",
			"2. カテゴリを選択してチケットを作成",
			"3. 専用チャンネルでサポートを受ける",
		}, "\n"), false).
		AddField("⚠️ 注意事項", strings.Join([]string{
			"• 同時に作成できるチケットは3つまでです",
			fmt.Sprintf("• %d時間非アクティブの場合、自動でクローズされます", settings.TicketAutoCloseHours),
			"• 不適切な使用は禁止されています",
		}, "\n"), false)

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "📧 チケット作成",
					Style:    discordgo.PrimaryButton,
					CustomID: "ticket_create",
				},
			},
		},
	}

	// パネルを現在のチャンネルに投稿
	_, err = s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{panelEmbed.Build()},
		Components: components,
	})

	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("❌ チケットパネルの作成に失敗しました: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// 成功メッセージ
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "✅ チケットパネルを作成しました！ユーザーはボタンをクリックしてチケットを作成できます。",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (h *InteractionHandler) handleTicketSetupDone(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    "✅ チケットシステム設定が完了しました！",
			Embeds:     []*discordgo.MessageEmbed{},
			Components: []discordgo.MessageComponent{},
		},
	})
}

func (h *InteractionHandler) handleTicketCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	guildID := i.GuildID

	// チケット設定を確認
	settings, err := h.db.GetGuildSettings(guildID)
	if err != nil || !settings.TicketEnabled {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ チケットシステムが設定されていません！",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// モーダルでチケットの詳細を入力
	modal := discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: "ticket_create_modal",
			Title:    "🎫 チケット作成",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "ticket_subject",
							Label:       "件名",
							Style:       discordgo.TextInputShort,
							Placeholder: "問題の概要を簡潔に入力してください",
							Required:    true,
							MaxLength:   100,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "ticket_description",
							Label:       "詳細説明",
							Style:       discordgo.TextInputParagraph,
							Placeholder: "問題の詳細、発生状況、求める解決策などを詳しく説明してください",
							Required:    true,
							MaxLength:   1000,
						},
					},
				},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID:    "ticket_priority",
							Label:       "優先度 (low/medium/high/urgent)",
							Style:       discordgo.TextInputShort,
							Placeholder: "low, medium, high, urgent のいずれかを入力",
							Required:    false,
							MaxLength:   10,
							Value:       "medium",
						},
					},
				},
			},
		},
	}

	s.InteractionRespond(i.Interaction, &modal)
}

func (h *InteractionHandler) handleTicketCreateModal(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ModalSubmitData()
	guildID := i.GuildID
	userID := i.Member.User.ID

	var subject, description, priority string
	for _, component := range data.Components {
		for _, comp := range component.(*discordgo.ActionsRow).Components {
			textInput := comp.(*discordgo.TextInput)
			switch textInput.CustomID {
			case "ticket_subject":
				subject = textInput.Value
			case "ticket_description":
				description = textInput.Value
			case "ticket_priority":
				priority = textInput.Value
			}
		}
	}

	if priority == "" {
		priority = "medium"
	}

	settings, err := h.db.GetGuildSettings(guildID)
	if err != nil || !settings.TicketEnabled {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ チケットシステムが利用できません",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "🎫 チケットを作成中です...",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})

	ticketNumber := fmt.Sprintf("%d", time.Now().Unix()%10000)
	channelName := fmt.Sprintf("ticket-%s", ticketNumber)
	
	channel, err := s.GuildChannelCreateComplex(guildID, discordgo.GuildChannelCreateData{
		Name:     channelName,
		Type:     discordgo.ChannelTypeGuildText,
		ParentID: settings.TicketCategoryID,
		PermissionOverwrites: []*discordgo.PermissionOverwrite{
			{
				ID:   guildID,
				Type: discordgo.PermissionOverwriteTypeRole,
				Deny: discordgo.PermissionViewChannel,
			},
			{
				ID:    userID,
				Type:  discordgo.PermissionOverwriteTypeMember,
				Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages | discordgo.PermissionReadMessageHistory,
			},
			{
				ID:    settings.TicketSupportRoleID,
				Type:  discordgo.PermissionOverwriteTypeRole,
				Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages | discordgo.PermissionReadMessageHistory | discordgo.PermissionManageMessages,
			},
		},
	})

	if err != nil {
		content := fmt.Sprintf("❌ チケットチャンネルの作成に失敗しました: %v", err)
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &content,
		})
		return
	}

	if settings.TicketAdminRoleID != "" {
		s.ChannelPermissionSet(channel.ID, settings.TicketAdminRoleID, discordgo.PermissionOverwriteTypeRole,
			discordgo.PermissionViewChannel|discordgo.PermissionSendMessages|discordgo.PermissionReadMessageHistory|discordgo.PermissionManageMessages|discordgo.PermissionManageChannels,
			0)
	}

	priorityEmoji := map[string]string{
		"low":    "🟢",
		"medium": "🟡", 
		"high":   "🟠",
		"urgent": "🔴",
	}
	emoji := priorityEmoji[priority]
	if emoji == "" {
		emoji = "🟡"
	}

	ticketEmbed := embed.New().
		SetTitle(fmt.Sprintf("🎫 チケット #%s", ticketNumber)).
		SetDescription(fmt.Sprintf("**件名:** %s", subject)).
		SetColor(embed.M3Colors.Primary).
		AddField("📝 詳細", description, false).
		AddField("⚡ 優先度", fmt.Sprintf("%s %s", emoji, strings.ToUpper(priority)), true).
		AddField("👤 作成者", fmt.Sprintf("<@%s>", userID), true).
		SetFooter("サポートスタッフが対応いたします", "").
		SetTimestamp()

	s.ChannelMessageSendComplex(channel.ID, &discordgo.MessageSend{
		Content: fmt.Sprintf("<@%s> <@&%s>", userID, settings.TicketSupportRoleID),
		Embeds:  []*discordgo.MessageEmbed{ticketEmbed.Build()},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "🔒 チケットを閉じる",
						Style:    discordgo.DangerButton,
						CustomID: fmt.Sprintf("ticket_close_%s", channel.ID),
					},
					discordgo.Button{
						Label:    "📋 トランスクリプト",
						Style:    discordgo.SecondaryButton,
						CustomID: fmt.Sprintf("ticket_transcript_%s", channel.ID),
					},
				},
			},
		},
	})

	successContent := fmt.Sprintf("✅ チケット #%s を作成しました！\n📍 チャンネル: <#%s>", ticketNumber, channel.ID)
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &successContent,
	})
}

func (h *InteractionHandler) handleTicketClose(s *discordgo.Session, i *discordgo.InteractionCreate, customID string) {
	channelID := strings.TrimPrefix(customID, "ticket_close_")
	guildID := i.GuildID
	userID := i.Member.User.ID

	// チケット設定を確認
	settings, err := h.db.GetGuildSettings(guildID)
	if err != nil || !settings.TicketEnabled {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ チケットシステムが利用できません",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// 権限チェック（チケット作成者、サポートロール、管理者ロールのみ）
	member, err := s.GuildMember(guildID, userID)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ 権限の確認に失敗しました",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// 権限確認
	hasPermission := false
	
	// サポートロールまたは管理者ロールを持っているかチェック
	for _, roleID := range member.Roles {
		if roleID == settings.TicketSupportRoleID || roleID == settings.TicketAdminRoleID {
			hasPermission = true
			break
		}
	}

	// チャンネル名からチケット作成者を確認（簡易版）
	channel, err := s.Channel(channelID)
	if err == nil && strings.HasPrefix(channel.Name, "ticket-") {
		// チャンネルの権限を確認してチケット作成者かどうかを判定
		for _, overwrite := range channel.PermissionOverwrites {
			if overwrite.Type == discordgo.PermissionOverwriteTypeMember && overwrite.ID == userID {
				hasPermission = true
				break
			}
		}
	}

	if !hasPermission {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ このチケットを閉じる権限がありません",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// 確認メッセージを表示
	confirmEmbed := embed.New().
		SetTitle("🔒 チケットクローズ確認").
		SetDescription("このチケットを閉じますか？\n\n⚠️ **注意**: チケットを閉じるとチャンネルが削除され、会話履歴は失われます。").
		SetColor(embed.M3Colors.Warning).
		AddField("💡 推奨", "重要な情報がある場合は、先に「📋 トランスクリプト」ボタンでログを保存してください。", false)

	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "🔒 チケットを閉じる",
					Style:    discordgo.DangerButton,
					CustomID: fmt.Sprintf("ticket_close_confirm_%s", channelID),
				},
				discordgo.Button{
					Label:    "❌ キャンセル",
					Style:    discordgo.SecondaryButton,
					CustomID: "ticket_close_cancel",
				},
			},
		},
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{confirmEmbed.Build()},
			Components: components,
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	})
}

func (h *InteractionHandler) handleTicketTranscript(s *discordgo.Session, i *discordgo.InteractionCreate, customID string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "🚧 トランスクリプト機能は近日実装予定です！",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (h *InteractionHandler) handleTicketCloseConfirm(s *discordgo.Session, i *discordgo.InteractionCreate, customID string) {
	channelID := strings.TrimPrefix(customID, "ticket_close_confirm_")
	guildID := i.GuildID
	userID := i.Member.User.ID

	// 設定を取得
	settings, err := h.db.GetGuildSettings(guildID)
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

	// 処理中メッセージ
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "🔒 チケットを閉じています...",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})

	// チャンネル情報を取得
	channel, err := s.Channel(channelID)
	if err != nil {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: func() *string { str := "❌ チャンネル情報の取得に失敗しました"; return &str }(),
		})
		return
	}

	// ログチャンネルに通知（チャンネル削除前）
	if settings.TicketLogChannelID != "" {
		closeEmbed := embed.New().
			SetTitle("🔒 チケットが閉じられました").
			SetColor(embed.M3Colors.Info).
			AddField("チャンネル", channel.Name, true).
			AddField("閉じた人", fmt.Sprintf("<@%s>", userID), true).
			SetTimestamp()

		s.ChannelMessageSendEmbed(settings.TicketLogChannelID, closeEmbed.Build())
	}

	// チャンネル削除（少し遅延をおいて実行）
	go func() {
		time.Sleep(2 * time.Second)
		s.ChannelDelete(channelID)
	}()

	// 成功メッセージ
	successContent := fmt.Sprintf("✅ チケット「%s」を閉じました", channel.Name)
	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &successContent,
	})
}

func (h *InteractionHandler) handleTicketCloseCancel(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    "❌ チケットクローズをキャンセルしました",
			Embeds:     []*discordgo.MessageEmbed{},
			Components: []discordgo.MessageComponent{},
		},
	})
}
