// backend/settings.go

package main

type Settings struct {
	SelectedTheme   string `json:"selectedTheme"`
	DefaultUsername string `json:"defaultUsername"`
	DefaultHost     string `json:"defaultHost"`
	DefaultPort     string `json:"defaultPort"`
}
