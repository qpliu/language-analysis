package main

import (
	"language-analysis/config"
	thanks "language-analysis/thank-analysis-src"
)

func main() {
	config.Run([]config.Command{
		config.Command{
			Name: "status",
			Run:  thanks.StatusCommand,
		},
		config.Command{
			Name: "collect",
			Run:  thanks.CollectCommand,
		},
	}, config.Command{
		Name: "collect",
		Run:  thanks.CollectCommand,
	}, nil, nil)
}
