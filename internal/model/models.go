package model

type Service struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Method      string `json:"method"`
	Description string `json:"description"`
	ImageKey    string `json:"imageKey"`
}

type Order struct {
	ID      string   `json:"id"`
	ItemIDs string   `json:"itemIds"`
	Counts  string   `json:"counts"`
	Results []string `json:"results"`
}
