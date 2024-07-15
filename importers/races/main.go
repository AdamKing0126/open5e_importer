package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"reflect"
	"strings"

	"net/http"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	nextUrl := "https://api.open5e.com/v1/races/"

	db, err := sqlx.Open("sqlite3", "../mud/sql_database/race_imports.db")
	if err != nil {
		log.Fatalf("Failed to open SQLite database: %v", err)
	} else {
		err := db.Ping()
		if err != nil {
			log.Fatalf("Failed to ping database: %v", err)
		}
		fmt.Println("Race Imports Database opened successfully")
	}
	defer db.Close()

	for {
		res, err := http.Get(nextUrl)
		if err != nil {
			log.Fatal(err)
		}

		bodyBytes, err := io.ReadAll(res.Body)
		res.Body.Close()
		if res.StatusCode > 299 {
			log.Fatalf("Response failed with status code: %d and \nbody: %s\n", res.StatusCode, bodyBytes)
		}
		if err != nil {
			log.Fatal(err)
		}

		races, next := convertJsonToRaceImports(bodyBytes)
		writeRacesToDB(db, races)
		fmt.Println(len(races))

		if next == "DONE" {
			break
		}

		nextUrl = next
		fmt.Printf("next url to fetch: %s\n", nextUrl)
	}
}

type Open5eResponse struct {
	Count    int          `json:"count"`
	Next     string       `json:"next"`
	Previous string       `json:"previous"`
	Results  []RaceImport `json:"results"`
}

type RaceImport struct {
	Age                string        `json:"age" db:"age"`
	Alignment          string        `json:"alignment" db:"alignment"`
	Asi                []interface{} `json:"asi" db:"asi"`
	AsiDescription     string        `json:"asi_desc" db:"asi_description"`
	Description        string        `json:"desc" db:"description"`
	DocumentLicenseUrl string        `json:"document__licence_url" db:"document_license_url"`
	DocumentSlug       string        `json:"document__slug" db:"document_slug"`
	DocumentTitle      string        `json:"document__title" db:"document_title"`
	DocumentUrl        string        `json:"document__url" db:"document_url"`
	Languages          string        `json:"languages" db:"languages"`
	Name               string        `json:"name" db:"name"`
	Size               string        `json:"size" db:"size"`
	SizeRaw            string        `json:"size_raw" db:"size_raw"`
	Slug               string        `json:"slug" db:"slug"`
	Speed              interface{}   `json:"speed" db:"speed"`
	SpeedDescription   string        `json:"speed_desc" db:"speed_description"`
	Subraces           []interface{} `json:"subraces" db:"subraces"`
	Traits             string        `json:"traits" db:"traits"`
	Vision             string        `json:"vision" db:"vision"`
}

// for used when comparing json field names with RaceImport field names
// Some field names need to be renamed to match the field name on the struct
func snakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	for i := 0; i < len(parts); i++ {
		parts[i] = strings.Title(parts[i])
	}
	res := strings.Join(parts, "")
	switch res {
	case "SpeedDesc":
		return "SpeedDescription"
	case "AsiDesc":
		return "AsiDescription"
	default:
		return res
	}
}

// when importing, we want to examine the field and see if it's
// present on the RaceImport object.  If it's not, we'll decide if
// it should be added or not.  this is a manual process.
func examineResult(result interface{}) {
	raceImportType := reflect.TypeOf(RaceImport{})
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		log.Fatal("could not assert result")
	}
	for key, value := range resultMap {
		camelKey := snakeToCamel(key)
		_, ok := raceImportType.FieldByName(camelKey)

		if !ok && key != "page_no" {
			fmt.Printf("Key/Value pair not found on MonsterImport:\nKey: %s, value: %v\n", key, value)
		}
	}
}

func convertJsonToRaceImports(jsonData []byte) ([]RaceImport, string) {
	var races []RaceImport
	var data map[string]interface{}
	err := json.Unmarshal(jsonData, &data)
	if err != nil {
		log.Fatal(err)
	}

	results, ok := data["results"].([]interface{})
	if !ok {
		log.Fatal("could not assert `results` as slice")
	}
	for i, result := range results {
		examineResult(result)
		resultJson, err := json.Marshal(result)
		if err != nil {
			log.Fatalf("could not marshal result at index %d: %v", i, err)
		}

		var race RaceImport
		err = json.Unmarshal(resultJson, &race)
		if err != nil {
			log.Fatalf("could not decode race at index %d: %v", i, err)
		}
		races = append(races, race)

	}
	nextUrl, ok := data["next"].(string)
	if !ok {
		return races, "DONE"
	}

	return races, nextUrl
}

func writeRacesToDB(db *sqlx.DB, races []RaceImport) {

	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS race_imports (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			age TEXT,
			alignment TEXT,
			asi TEXT,
			asi_description TEXT,
			description TEXT,
			document_license_url TEXT,
			document_slug TEXT,
			document_title TEXT,
			document_url TEXT,
			languages TEXT,
			name TEXT,
			size TEXT,
			size_raw TEXT,
			slug TEXT,
			speed TEXT,
			speed_description TEXT,
			subraces TEXT,
			traits TEXT,
			vision TEXT
		);
	`)
	if err != nil {
		log.Fatalf("Failed to create race_imports table: %v", err)
	}

	for idx := range races {
		query := `INSERT INTO race_imports (
			age, alignment, asi, asi_description, description, document_license_url, document_slug, document_title,
			document_url, languages, name, size, size_raw, slug, speed, speed_description, subraces, traits, vision) VALUES
			(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`

		race := races[idx]

		asi, err := json.Marshal(race.Asi)
		if err != nil {
			log.Fatalf("Failed to marshal Asi: %v", err)
		}

		speed, err := json.Marshal(race.Speed)
		if err != nil {
			log.Fatalf("Failed to marshal Size: %v", err)
		}

		subraces, err := json.Marshal(race.Subraces)
		if err != nil {
			log.Fatalf("Failed to marshal Subraces: %v", err)
		}

		_, err = db.Exec(
			query, race.Age, race.Alignment, asi, race.AsiDescription, race.Description, race.DocumentLicenseUrl,
			race.DocumentSlug, race.DocumentTitle, race.DocumentUrl, race.Languages, race.Name, race.Size, race.SizeRaw,
			race.Slug, speed, race.SpeedDescription, subraces, race.Traits, race.Vision)
		if err != nil {
			log.Fatalf("Failed to insert row into race_imports table: %v", err)
		}
	}
}
