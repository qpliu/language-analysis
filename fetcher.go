package main

import (
	"language-analysis/config"
	fetcher "language-analysis/fetcher-src"
)

func main() {
	config.Run([]config.Command{
		config.Command{
			Name: "status",
			Run:  fetcher.StatusCommand,
		},
		config.Command{
			Name: "fetch",
			Run:  fetcher.FetchCommand,
		},
		config.Command{
			Name: "add-feeds",
			Run:  fetcher.AddFeedsCommand,
		},
	}, config.Command{
		Run: fetcher.FetchLoopCommand,
	}, func() error {
		filename := config.Dir() + "/fetcher.toml"
		return config.ReadConfig(filename, &fetcher.Config)
	}, nil)
}
