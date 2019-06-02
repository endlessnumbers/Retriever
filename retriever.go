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
	//first: get user API key if exists
	//*** TODO ***
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
					APIKey:  apiKey,
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

func performRequest(userPreferences preferences, url string) {
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

func parseResponse(responseBytes []byte) string {
	//parse json string into object

	parsedData := response{}
	err := json.Unmarshal(responseBytes, &parsedData)

	if err != nil {
		panic(err)
	}

	outValue := ""
	for _, s := range parsedData.Articles {
		outValue += s.Title + "\n"
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
	// if len(parameters.sources) > 0 {

	// }
	// if len(parameters.from) > 0 {

	// }
	// if len(parameters.to) > 0 {

	// }
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

	allowedCountries := []string{"ae", "ar", "at", "au", "be", "bg", "br", "ca", "ch", "cn", "co", "cu", "cz", "de",
		"eg", "fr", "gb", "gr", "hk", "hu", "id", "ie", "il", "in", "it", "jp", "kr", "lt", "lv", "ma", "mx", "my",
		"ng", "nl", "no", "nz", "ph", "pl", "pt", "ro", "rs", "ru", "sa", "se", "sg", "si", "sk", "th", "tr", "tw",
		"ua", "us", "ve", "za"}

	allowedCategories := []string{"business", "entertainment", "general", "health", "science", "sports", "technology"}

	if len(parameters.keywords) > 0 {
		headlineString += "q=" + parameters.keywords
	}
	if len(parameters.country) == 2 {
		for _, co := range allowedCountries {
			if co == parameters.country {
				headlineString += "country=" + co + "&"
			}
		}
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
