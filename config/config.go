package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

var options = map[string]string{
	"dir": "./data",

	"fetcher-sleep":        "15s",
	"thank-collect-count":  "50",
	"phrase-collect-count": "500",
}

type Command struct {
	Name string
	Run  func() error
}

func Run(commands []Command, defaultCommand Command, initialize func() error, processArg func(string) (bool, error)) {
	run := false
	for _, arg := range os.Args[1:] {
		if len(arg) > 0 && arg[0] == '-' {
			if eq := strings.Index(arg, "="); eq > 0 {
				options[arg[1:eq]] = arg[eq+1:]
			} else {
				options[arg[1:eq]] = ""
			}
			continue
		}
		run = false
		for _, c := range commands {
			if arg == c.Name {
				run = true
				if initialize != nil {
					if err := initialize(); err != nil {
						fmt.Fprintf(os.Stderr, "%s: %s: %v\n", os.Args[0], arg, err)
						os.Exit(1)
					}
				}
				if err := c.Run(); err != nil {
					fmt.Fprintf(os.Stderr, "%s: %s: %v\n", os.Args[0], arg, err)
					os.Exit(1)
				}
				break
			}
		}
		if !run && processArg != nil {
			processed, err := processArg(arg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: %s: %v\n", os.Args[0], arg, err)
				os.Exit(1)
			}
			run = processed
		}
		if !run {
			fmt.Fprintf(os.Stderr, "%s: unknown command %s.  Available commands:", os.Args[0])
			for _, c := range commands {
				fmt.Fprintf(os.Stderr, " %s", c.Name)
			}
			fmt.Fprintf(os.Stderr, "\n")
			os.Exit(1)
		}
	}
	if !run {
		if initialize != nil {
			if err := initialize(); err != nil {
				fmt.Fprintf(os.Stderr, "%s: %v\n", os.Args[0], err)
				os.Exit(1)
			}
		}
		if err := defaultCommand.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", os.Args[0], err)
			os.Exit(1)
		}
	}
}

func Dir() string {
	return String("dir", "./data")
}

func String(name string, defaultValue string) string {
	if value, ok := options[name]; ok {
		return value
	} else {
		return defaultValue
	}
}

func Int(name string, defaultValue int) (int, error) {
	if value, ok := options[name]; ok {
		var v int
		_, err := fmt.Sscanf(value, "%d", &v)
		return v, err
	} else {
		return defaultValue, nil
	}
}

func Duration(name string, defaultValue time.Duration) (time.Duration, error) {
	if value, ok := options[name]; ok {
		return time.ParseDuration(value)
	} else {
		return defaultValue, nil
	}
}

func ReadConfig(filename string, conf any) error {
	if _, err := toml.DecodeFile(filename, conf); err != nil {
		return err
	}
	return nil
}
