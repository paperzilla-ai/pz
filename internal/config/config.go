package config

import "os"

func APIURL() string {
	if v := os.Getenv("PZ_API_URL"); v != "" {
		return v
	}
	return "https://paperzilla.ai"
}
