package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"consumer-test-runner/config"
	"consumer-test-runner/model"
)

func HandleRequest(data model.AutomationRequest) error {
	bodyBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	resp, err := http.Post(config.AppConfig.AutomationAPI, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("[✓] Response : %s\n", body)

	return nil
}
