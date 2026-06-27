package service

import (
	"fmt"
	"log"

	"github.com/Yo0GuitarIT/yo0-backend/internal/config"
	"github.com/bwmarrin/discordgo"
)

// drawTarotButtonID 是「抽牌」按鈕的識別碼，點擊時 Discord 會帶回這個值
const drawTarotButtonID = "draw_tarot"

// StartDiscordBot 連上 Discord Gateway、貼出抽牌按鈕並監聽點擊。
// 與 Telegram 不同：按鈕互動只需要 Guilds intent，不必開 Message Content 特權。
func StartDiscordBot() error {
	token := config.DiscordBotToken()
	if token == "" {
		return fmt.Errorf("DISCORD_BOT_TOKEN 未設定")
	}

	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return fmt.Errorf("無法建立 Discord session: %w", err)
	}
	session.Identify.Intents = discordgo.IntentsGuilds

	// 監聽互動事件（按鈕點擊）
	session.AddHandler(handleDiscordInteraction)

	if err := session.Open(); err != nil {
		return fmt.Errorf("無法連線 Discord Gateway: %w", err)
	}
	defer session.Close()

	log.Printf("Discord bot 已啟動：%s", session.State.User.Username)

	// 啟動時在每個設定的頻道貼出「恩雅婆婆開放占卜摟」按鈕
	for _, chID := range config.DiscordChannelIDs() {
		if err := postStartupButton(session, chID); err != nil {
			log.Printf("[Discord] 啟動時貼按鈕失敗（頻道 %s）: %v", chID, err)
		}
	}

	// 保持 Gateway 連線（直到程式結束）
	select {}
}

// postStartupButton 在指定頻道貼出啟動公告訊息與「🔮 抽牌」按鈕。
func postStartupButton(session *discordgo.Session, channelID string) error {
	if channelID == "" {
		return fmt.Errorf("頻道 ID 為空")
	}
	_, err := session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content: "恩雅婆婆開放占卜摟 🔮",
		Flags:   discordgo.MessageFlagsSuppressNotifications,
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "🔮 抽牌",
						Style:    discordgo.PrimaryButton,
						CustomID: drawTarotButtonID,
					},
				},
			},
		},
	})
	return err
}

// postTarotButton 在指定頻道貼一則帶「🔮 抽牌」按鈕的訊息。
// 由每次抽完牌後補貼，讓按鈕永遠在最新訊息底部。
func postTarotButton(session *discordgo.Session, channelID string) error {
	if channelID == "" {
		return fmt.Errorf("頻道 ID 為空")
	}

	_, err := session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content: "🔮 今日塔羅占卜，點下面的按鈕抽一張牌：",
		Flags:   discordgo.MessageFlagsSuppressNotifications, // 靜音傳送，不跳通知
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "🔮 抽牌",
						Style:    discordgo.PrimaryButton,
						CustomID: drawTarotButtonID,
					},
				},
			},
		},
	})
	return err
}
