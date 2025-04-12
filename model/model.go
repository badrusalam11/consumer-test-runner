package model

type AutomationRequest struct {
	Project     string `json:"project"`
	TestsuiteID string `json:"testsuite_id"`
	TotalSteps  int    `json:"total_steps"`
	Retries     int    `json:"retries"`
}
