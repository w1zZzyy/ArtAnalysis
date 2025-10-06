package model

type ArtCenter struct {
	ArtID          string `json:"artId"`
	Title          string `json:"title"`
	Algorithm      string `json:"algorithm"`
	ArtDescription string `json:"artDescription"`
	ArtImageKey    string `json:"artImageKey"`
}

type Basket struct {
	BasketID string `json:"basketId"`
	ArtIDs   string `json:"artIds"`
	Counts   string `json:"counts"`
}

type AnalysisResult struct {
	BasketID string
	Results  map[string]string // map[ArtID]Coordinates
}
