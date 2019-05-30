package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type preferences struct {
	ApiKey  string
	Country string
	Valid   bool
}

type responseObject struct {
	Articles []article
}

type article struct {
	Author      string
	Title       string
	Description string
}

func main() {
	selectOutput(os.Args[1:])
}

func selectOutput([]string) {
	//first: get user API key if exists
	userPreferences := getUserPreferences()

	if !userPreferences.Valid {
		os.Exit(1)
	}

	if len(os.Args) > 1 {
		parseUserArguments()
	} else {
		performDefaultRequest(userPreferences)
	}

	//get headlines
}

func getUserPreferences() preferences {
	//check if preferences file exists
	if _, err := os.Stat("preferences.json"); os.IsNotExist(err) {
		// create file -- prompt user for api key
		reader := bufio.NewReader(os.Stdin)

		for {
			fmt.Print("Enter API Key: ")
			apiKey, _ := reader.ReadString('\n')
			if validateApiKey(apiKey) {
				//create json and return it
				pref := &preferences{
					ApiKey:  apiKey,
					Country: "gb"}
				prefJson, _ := json.Marshal(pref)
				writeToJsonFile(string(prefJson))
				return *pref
			} else {
				fmt.Print("\nInvalid API Key!")
				return preferences{
					Valid: false}
			}
		}
	} else {
		//preferences file exists so get api key
		//maybe validate it again to check it still works?
		//get json object from file
		fileData, _ := ioutil.ReadFile("preferences.json")
		fileJson := preferences{}
		json.Unmarshal(fileData, &fileJson)
		return fileJson
	}
}

func validateApiKey(key string) bool {
	if len(key) == 0 {
		return false
	}
	url := fmt.Sprintf(`https://newsapi.org/v2/top-headlines?
		country=us&
		apiKey=%d`,
		key)

	resp, _ := http.Get(url)
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true
	}
	return false
}

func writeToJsonFile(json string) {
	//write to file
}

func parseUserArguments() {
	country := flag.String("country", "gb", `Enter the two-character code for the
		country whose headlines you wish to search for. Full list of codes...`)
	flag.Parse()

	if len(*country) > 2 {
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func performDefaultRequest(userPreferences preferences) {
	url := fmt.Sprintf("https://newsapi.org/v2/top-headlines?country=us&apiKey=%s",
		userPreferences.ApiKey)

	resp, _ := http.Get(url)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Print(fmt.Sprintf("Error! %d", resp.StatusCode))
	} else {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print(parseResponse(bodyBytes))
	}
}

func parseResponse(response []byte) string {
	//parse json string into object

	parsedData := responseObject{}
	err := json.Unmarshal(response, &parsedData)

	if err != nil {
		panic(err)
	}

	outValue := ""
	for _, s := range parsedData.Articles {
		outValue += s.Title + "\n"
	}

	return outValue
}
