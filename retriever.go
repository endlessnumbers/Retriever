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

const endpoint = "https://newsapi.org/v2/"
const headlinesEndpoint = endpoint + "top-headlines?"
const searchEndpoint = endpoint + "everything?"

type preferences struct {
	APIKey  string
	Country string
	Valid   bool
}

type response struct {
	Articles []article
}

type article struct {
	Author      string
	Title       string
	Description string
}

type request struct {
	keywords string
	sources  []string
}

type topHeadlinesRequest struct {
	country  string
	category string
	request
}

type everythingRequest struct {
	from     string
	to       string
	language string
	sortBy   string
	request
}

func main() {
	selectOutput(os.Args[1:])
}

func selectOutput([]string) {
	userPreferences := getUserPreferences()

	if !userPreferences.Valid {
		os.Exit(1)
	}

	searchEverything := flag.Bool("e", false, "Use this flag if you wish to search for stories rather than returning the headlines.")
	//figure out what args we have and create request accordingly
	keywords := flag.String("keyword", "", "Keyword or phrase to search for.")
	//Need to get list of sources for user for this - can use /sources endpoint
	//sources := flag.String("sources", "", "News sources to fetch headlines from.")

	//should provide list of possible values to user
	country := flag.String("country", userPreferences.Country, "ISO 3166-1 code of the country to get headlines for.")
	//again, provide list of options
	category := flag.String("category", "", "Provide a category to get the headlines for. Note: you can't mix this with the sources parameter!")

	//EVERYTHING
	from := flag.String("from", "", "Date for the oldest article allowed.")
	to := flag.String("to", "", "Date for the newest article allowed.")
	//again, return options:
	language := flag.String("lang", "", "ISO-631-9 code for the languages to search in.")
	sortBy := flag.String("sort", "relevancy", "Sort order for the articles.")

	flag.Parse()

	if len(os.Args) > 1 {
		if *searchEverything {
			search := &everythingRequest{
				request:  request{keywords: *keywords},
				from:     *from,
				to:       *to,
				language: *language,
				sortBy:   *sortBy}
			performEverythingSearch(*search, userPreferences)
		} else {
			headlineRequest := &topHeadlinesRequest{
				request:  request{keywords: *keywords},
				country:  *country,
				category: *category}
			fetchHeadlines(*headlineRequest, userPreferences)
		}
	} else {
		performDefaultRequest(userPreferences)
	}
}

func getUserPreferences() preferences {
	//check if preferences file exists
	if _, err := os.Stat("preferences.json"); os.IsNotExist(err) {
		// create file -- prompt user for api key
		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Enter API Key: ")
		apiKey, _ := reader.ReadString('\n')
		apiKey = apiKey[:len(apiKey)-1]
		if validateAPIKey(apiKey) {
			//get default country
			fmt.Print("Enter country code: ")
			country, _ := reader.ReadString('\n')
			country = country[:len(country)-1]
			if validateCountryCode(country) {
				pref := &preferences{
					APIKey:  apiKey,
					Country: country,
					Valid:   true}
				prefJSON, _ := json.Marshal(pref)
				if writeToJSONFile(string(prefJSON)) {
					fmt.Print("Preferences saved.\n")
					return *pref
				}
				return preferences{
					Valid: false}
			}
			fmt.Print("\nInvalid country code!")
		} else {
			fmt.Print("\nInvalid API Key!")
		}
		//if we've got to this point, preferences are invalid!
		return preferences{
			Valid: false}
	} else {
		//preferences file exists
		fileData, _ := ioutil.ReadFile("preferences.json")
		fileJSON := preferences{}
		json.Unmarshal(fileData, &fileJSON)
		return fileJSON
	}
}

func validateAPIKey(key string) bool {
	if len(key) == 0 {
		return false
	}
	url := "https://newsapi.org/v2/top-headlines?country=us&apiKey=" + key

	resp, err := http.Get(url)

	if err != nil {
		return false
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true
	}
	return false
}

func validateCountryCode(countryCode string) bool {
	if len(countryCode) != 2 {
		return false
	}

	allowedCountries := []string{"ae", "ar", "at", "au", "be", "bg", "br", "ca", "ch", "cn", "co", "cu", "cz", "de",
		"eg", "fr", "gb", "gr", "hk", "hu", "id", "ie", "il", "in", "it", "jp", "kr", "lt", "lv", "ma", "mx", "my",
		"ng", "nl", "no", "nz", "ph", "pl", "pt", "ro", "rs", "ru", "sa", "se", "sg", "si", "sk", "th", "tr", "tw",
		"ua", "us", "ve", "za"}

	for _, co := range allowedCountries {
		if co == countryCode {
			return true
		}
	}
	return false
}

func writeToJSONFile(json string) bool {
	output := []byte(json)
	err := ioutil.WriteFile("preferences.json", output, 0644)

	if err != nil {
		return false
	}
	return true
}

func performRequest(userPreferences preferences, url string) {
	resp, err := http.Get(url)

	if err != nil {
		fmt.Print("Error! Could not carry out request: " + err.Error())

	}

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

func parseResponse(responseBytes []byte) string {
	parsedData := response{}
	err := json.Unmarshal(responseBytes, &parsedData)

	if err != nil {
		panic(err)
	}

	outValue := ""
	for _, s := range parsedData.Articles {
		outValue += " - \t" + s.Title + "\n"
	}

	return outValue
}

func performDefaultRequest(userPreferences preferences) {
	url := headlinesEndpoint + "country=" + userPreferences.Country + "&apiKey=" + userPreferences.APIKey
	performRequest(userPreferences, url)
}

func performEverythingSearch(parameters everythingRequest, userPreferences preferences) {
	searchString := searchEndpoint

	allowedLanguages := []string{"ar", "de", "en", "es", "fr", "he", "it", "nl", "no", "pt", "ru", "se", "ud", "zh"}
	allowedSorts := []string{"relevancy", "popularity", "publishedAt"}

	if len(parameters.keywords) > 0 {
		searchString += "q=" + parameters.keywords
	}
	if len(parameters.language) == 2 {
		for _, lang := range allowedLanguages {
			if lang == parameters.language {
				searchString += "language=" + lang + "&"
			}
		}
	}

	if len(parameters.sortBy) > 0 {
		for _, srt := range allowedSorts {
			if srt == parameters.sortBy {
				searchString += "sortBy=" + srt + "&"
			}
		}
	}

	searchString += "apiKey=" + userPreferences.APIKey
	performRequest(userPreferences, searchString)
}

func fetchHeadlines(parameters topHeadlinesRequest, userPreferences preferences) {
	headlineString := headlinesEndpoint

	allowedCategories := []string{"business", "entertainment", "general", "health", "science", "sports", "technology"}

	if len(parameters.keywords) > 0 {
		headlineString += "q=" + parameters.keywords
	}
	if validateCountryCode(parameters.country) {
		headlineString += "country=" + parameters.country + "&"
	} else {
		headlineString += "country=" + userPreferences.Country + "&"
	}
	if len(parameters.category) > 0 {
		for _, cat := range allowedCategories {
			if cat == parameters.category {
				headlineString += "category=" + cat + "&"
			}
		}
	}

	headlineString += "apiKey=" + userPreferences.APIKey
	performRequest(userPreferences, headlineString)
}
