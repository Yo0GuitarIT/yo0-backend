package service

import (
	"bytes"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

// handleDiscordInteraction 處理按鈕點擊：抽一張牌並以占卜訊息回覆點擊者。
func handleDiscordInteraction(session *discordgo.Session, i *discordgo.InteractionCreate) {
	// 只處理「抽牌」按鈕，其餘互動忽略
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}
	if i.MessageComponentData().CustomID != drawTarotButtonID {
		return
	}

	// Discord 要求 3 秒內回應；逆位要下載+旋轉圖會超時，所以先 defer。
	// 用 DeferredMessageUpdate（靜默確認按鈕點擊），不會冒出「思考中…」佔位訊息，
	// 因此也不會有那則無法靜音的通知聲；結果改用一般靜音頻道訊息送出。
	if err := session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	}); err != nil {
		log.Printf("[Discord] defer 失敗: %v", err)
		return
	}

	// 點擊後立即把被點的這顆按鈕變灰，只留稍後補貼的最新一顆可點
	disableTarotButton(session, i.Message)

	username := discordUsername(i)

	card, _, err := GetRandomTarot()
	if err != nil {
		session.ChannelMessageSendComplex(i.ChannelID, &discordgo.MessageSend{
			Content: "❌ 抽牌失敗，請稍後再試",
			Flags:   discordgo.MessageFlagsSuppressNotifications,
		})
		return
	}

	orientation := "正位"
	embed := &discordgo.MessageEmbed{
		Description: fmt.Sprintf("**%s** 已為您占卜了", username),
	}
	params := &discordgo.MessageSend{}

	if card.Reversed {
		orientation = "逆位"
		// 逆位：下載牌面圖、旋轉 180° 後以附件上傳
		imgBytes, rotErr := fetchRotatedImage(card.Image)
		if rotErr != nil {
			// 旋轉失敗就退回正向網址圖
			embed.Image = &discordgo.MessageEmbedImage{URL: card.Image}
		} else {
			params.Files = []*discordgo.File{{
				Name:        "tarot.jpg",
				ContentType: "image/jpeg",
				Reader:      bytes.NewReader(imgBytes),
			}}
			embed.Image = &discordgo.MessageEmbedImage{URL: "attachment://tarot.jpg"}
		}
	} else {
		// 正位：直接給網址，由 Discord 抓圖
		embed.Image = &discordgo.MessageEmbedImage{URL: card.Image}
	}

	embed.Title = fmt.Sprintf("🔮 %s（%s）· %s", card.NameZh, card.Name, orientation)
	if card.Meaning != "" {
		embed.Fields = []*discordgo.MessageEmbedField{
			{Name: "牌義", Value: card.Meaning},
		}
	}
	params.Embeds = []*discordgo.MessageEmbed{embed}
	params.Flags = discordgo.MessageFlagsSuppressNotifications // 靜音傳送，不跳通知

	if _, err := session.ChannelMessageSendComplex(i.ChannelID, params); err != nil {
		log.Printf("[Discord] 發送占卜結果失敗: %v", err)
	}

	// 抽完後在結果下方補貼一則新的「🔮 抽牌」按鈕，讓按鈕永遠在最新訊息底部
	if err := postTarotButton(session, i.ChannelID); err != nil {
		log.Printf("[Discord] 補貼按鈕失敗: %v", err)
	}
}

// disableTarotButton 把指定訊息上的「抽牌」按鈕改成灰色不可點（已被抽過）。
func disableTarotButton(session *discordgo.Session, msg *discordgo.Message) {
	if msg == nil {
		return
	}
	disabled := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "🔮 已抽過",
					Style:    discordgo.SecondaryButton,
					CustomID: drawTarotButtonID,
					Disabled: true,
				},
			},
		},
	}
	if _, err := session.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Channel:    msg.ChannelID,
		ID:         msg.ID,
		Components: &disabled,
	}); err != nil {
		log.Printf("[Discord] 停用舊按鈕失敗: %v", err)
	}
}

// discordUsername 取點擊者的顯示名稱：伺服器暱稱優先，否則用帳號名。
func discordUsername(i *discordgo.InteractionCreate) string {
	if i.Member != nil && i.Member.User != nil {
		if i.Member.Nick != "" {
			return i.Member.Nick
		}
		return i.Member.User.Username
	}
	if i.User != nil {
		return i.User.Username
	}
	return "神祕旅人"
}
