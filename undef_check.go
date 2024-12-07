package main

import (
	"bufio"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"

	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/projectdiscovery/gologger"
)

const (
	apiKey       = "9d2eca52f53de4f2aeeb9624bffda71e"
	apiBaseURL   = "https://undef.info/api/"
	pollInterval = 5 * time.Second // Poll every 5 seconds
)

func checkCard() {
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


func parseCardsForUndef() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter card details: ")
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading input: %s\n", err)
		return
	}

	// Parse the input
	details := strings.Split(strings.TrimSpace(input), ";")
	if len(details) < 3 {
		fmt.Println("Invalid input format.")
		return
	}

	cardNumber := details[0]
	expiry := strings.Split(details[1], "/")
	expMonth := expiry[0]
	expYear := "20" + expiry[1] // Assuming the year is in YY format and needs to be converted to YYYY
	cvv := details[2]

	gologger.Info().Msgf("Card Number: %s\n", cardNumber)
	gologger.Info().Msgf("Expiry: %s/%s\n", expMonth, expYear)
	gologger.Info().Msgf("CVV: %s\n", cvv)
	gologger.Info().Msgf("API Key: %s\n", apiKey)

	// Prepare CSV data
	csvData := fmt.Sprintf("%s\t%s\t%s\t%s\n", cardNumber, expMonth, expYear, cvv)
	encodedData := base64.StdEncoding.EncodeToString([]byte(csvData))

	// Construct the API request for card check
	apiURL := apiBaseURL + "cards.html"
	gologger.Info().Msgf("API URL: %s", apiURL)

	data := url.Values{
		"checker":   {"pre"},
		"data":      {encodedData},
		"encrypted": {"0"},
	}
	gologger.Info().Msgf("Payload: %s\n", data.Encode())

	// Send the request for card check
	resp, err := sendRequest(apiURL, data)
	if err != nil {
		fmt.Printf("Error sending request: %s\n", err)
		return
	}

	// Parse the response to get the api_group_id
	apiGroupID, err := parseResponse(resp)
	if err != nil {
		fmt.Printf("Error parsing response: %s\n", err)
		return
	}

	// Polling for status resolution
	for {
		// Construct the API request for result check
		apiURL = apiBaseURL + "ccres.html"
		data = url.Values{
			"api_group_id": {apiGroupID},
			"encrypted":    {"0"},
		}

		// Send the request for result check
		resp, err = sendRequest(apiURL, data)
		if err != nil {
			fmt.Printf("Error sending request: %s\n", err)
			return
		}

		gologger.Info().Msgf("Response: %s\n", resp)
		// Check if the status is resolved
		statusResolved, err := checkStatusResolved(resp)
		if err != nil {
			fmt.Printf("Error checking status: %s\n", err)
			return
		}

		if statusResolved {
			fmt.Println("Final Response:")
			fmt.Println(resp)
			break
		}

		fmt.Println("Status not resolved, polling again in 5 seconds...")
		time.Sleep(pollInterval)
	}
}

func sendUndefRequest(apiURL string, data url.Values) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.URL.RawQuery = "key=" + apiKey

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(respBody), nil
}

func parseResponse(response string) (string, error) {
	reader := csv.NewReader(strings.NewReader(response))
	reader.Comma = '\t'
	records, err := reader.ReadAll()
	if err != nil {
		return "", err
	}

	for _, record := range records {
		if len(record) >= 7 {
			return record[6], nil // api_group_id is the 7th field in the response
		}
	}

	return "", fmt.Errorf("api_group_id not found in the response")
}

func checkStatusResolved(response string) (bool, error) {
	reader := csv.NewReader(strings.NewReader(response))
	reader.Comma = '\t'
	records, err := reader.ReadAll()
	if err != nil {
		return false, err
	}

	for _, record := range records {
		gologger.Info().Msgf("Record: %s", record)
		if len(record) >= 7 {
			status := record[7] // status is the 8th field in the response
			if status == "check_wait" || status == "checking" {
				return false, nil
			}
		}
	}

	return true, nil
}
