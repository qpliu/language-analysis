package main

import (
	"language-analysis/config"
	phrases "language-analysis/phrase-analysis-src"
)

func main() {
	config.Run([]config.Command{
		config.Command{
			Name: "status",
			Run:  phrases.StatusCommand,
		},
		config.Command{
			Name: "collect",
			Run:  phrases.CollectCommand,
		},
		config.Command{
			Name: "add",
			Run:  phrases.AddCommand,
		},
	}, config.Command{
		Name: "collect",
		Run:  phrases.CollectCommand,
	}, func() error {
		filename := config.Dir() + "/phrase-analysis.toml"
		return config.ReadConfig(filename, &phrases.Config)
	}, nil)
}
