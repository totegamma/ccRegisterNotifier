package main

import (
	"os"
	"fmt"
	"time"
	"bytes"
	"net/http"
	"strconv"
	"encoding/json"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"github.com/totegamma/concurrent/core"
)

type WebhookData struct {
	Username string `json:"username,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
	Content string `json:"content"`
}

func main() {
	domainName := os.Getenv("DOMAIN_NAME")
	dsn := os.Getenv("DSN")
	spanStr := os.Getenv("SPAN")
	span, err := strconv.Atoi(spanStr)
	if err != nil {
		panic("failed to parse span")
	}
	webhookURL := os.Getenv("WEBHOOK_URL")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	var entities []core.Entity
	err = db.Where("c_date > ? and domain = ?", time.Now().Add(-1*time.Duration(span)*time.Minute), domainName).Find(&entities).Error
	if err != nil {
		panic("failed to get registration")
	}

	if len(entities) == 0 {
		fmt.Println("no registration")
		return
	}

	var ids []string
	for _, entity := range entities {
		ids = append(ids, entity.ID)
	}

	// get registration current-5min
	var metas []core.EntityMeta
	err = db.Where("id IN ?", ids).Find(&metas).Error
	if err != nil {
		panic("failed to get registration")
	}

	var report = "=== Registration Report ===\n"
	for _, meta := range metas {

		report += fmt.Sprintf("https://concrnt.world/%s\n", meta.ID)

		var info map[string]any
		err = json.Unmarshal([]byte(meta.Info), &info)

		nameAny, ok := info["name"]
		if ok {
			name := nameAny.(string)
			report += fmt.Sprintf("  Name: %s\n", name)
		}

		emailAny, ok := info["email"]
		if ok {
			email := emailAny.(string)
			report += fmt.Sprintf("  Email: %s\n", email)
		}

		socialAny, ok := info["social"]
		if ok {
			social := socialAny.(string)
			report += fmt.Sprintf("  Social: %s\n", social)
		}
		report += "\n"
	}

	data := WebhookData{
		Username: domainName,
		AvatarURL: os.Getenv("WEBHOOK_AVATAR_URL"),
		Content: report,
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		panic("failed to marshal data")
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(dataBytes))
	if err != nil {
		panic("failed to create request")
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic("failed to send request")
	}

	fmt.Println(resp.Status)
}

