package notification

import (
	"boralabs/config"
	"fmt"
	"github.com/slack-go/slack"
	"log"
)

var SlackLogger Slack

type Slack struct {
}

func (s Slack) SendMessage(message string) error {
	defer func() {
		if r := recover(); r != nil {
			log.Println(fmt.Sprintf("Panic Slack SendMessage %v", r))
		}
	}()

	if config.C.GetString("slack.webhook_url") == "" {
		log.Println("slack.webhook_url is not configured.")
		return nil
	}

	textBlocks := make([]*slack.TextBlockObject, 0)
	textBlocks = append(textBlocks,
		slack.NewTextBlockObject(slack.MarkdownType, message, false, false),
	)
	return slack.PostWebhook(config.C.GetString("slack.webhook_url"), &slack.WebhookMessage{
		Blocks: &slack.Blocks{BlockSet: []slack.Block{
			slack.NewSectionBlock(slack.NewTextBlockObject(slack.MarkdownType, fmt.Sprintf("Environment : %s\n", config.C.GetString("env")), false, false),
				textBlocks, nil,
			),
		},
		}})
}
