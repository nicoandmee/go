package main

import (
	"log"

	"github.com/gospider007/requests"
)

func gospider() {
	resp, err := requests.Get(nil, "http://httpbin.org/anything")
	if err != nil {
      log.Panic(err)
    }
    log.Print(resp.Text())    // Get content and parse as string
    log.Print(resp.Content()) // Get content as bytes
    log.Print(resp.Json())    // Get content and parse as gjson JSON
    log.Print(resp.Html())    // Get content and parse as goquery DOM
    log.Print(resp.Cookies()) // Get cookies
}
