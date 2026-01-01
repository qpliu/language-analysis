package fetcher

import (
	"fmt"
	"io"
	"net/http"
)

func fetch(url string) ([]byte, error) {
	fmt.Printf("Fetching %s...", url)
	defer fmt.Printf("done\n")
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	return io.ReadAll(response.Body)
}
