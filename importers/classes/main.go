package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	nextUrl := "https://api.open5e.com/v1/classes/"

	db, err := sqlx.Open("sqlite3", "./sql_database/class_imports.db")
	if err != nil {
		log.Fatalf("Failed to open SQLite database: %v", err)
	} else {
		err := db.Ping()
		if err != nil {
			log.Fatalf("Failed to ping database: %v", err)
		}
		fmt.Println("Database opened successfully")
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
			log.Fatalf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, bodyBytes)
		}
		if err != nil {
			log.Fatal(err)
		}

		classes, next := convertJsonToClassImports(bodyBytes)
		writeClassesToDB(db, classes)

		if next == "DONE" {
			break
		}

		nextUrl = next
		fmt.Printf("next url to fetch: %s\n", nextUrl)
	}
}

type Open5eResponse struct {
	Count    int           `json:"count"`
	Next     string        `json:"next"`
	Previous string        `json:"previous"`
	Results  []ClassImport `json:"results"`
}

type ClassImport struct {
	Name                      string        `json:"name" db:"name"`
	Slug                      string        `json:"slug" db:"slug"`
	Description               string        `json:"desc" db:"description"`
	HitDice                   string        `json:"hit_dice" db:"hit_dice"`
	HpAtFirstLevel            string        `json:"hp_at_1st_level" db:"hp_at_first_level"`
	HpAtHigherLevels          string        `json:"hp_at_higher_levels" db:"hp_at_higher_levels"`
	ProficienciesArmor        string        `json:"prof_armor" db:"proficiencies_armor"`
	ProficienciesWeapons      string        `json:"prof_weapons" db:"proficiencies_weapons"`
	ProficienciesTools        string        `json:"prof_tools" db:"proficiencies_tools"`
	ProficienciesSavingThrows string        `json:"prof_saving_throws" db:"proficiencies_saving_throws"`
	ProficienciesSkills       string        `json:"prof_skills" db:"proficiencies_skills"`
	Equipment                 string        `json:"equipment" db:"equipment"`
	Table                     string        `json:"table" db:"class_table"`
	SpellcastingAbility       string        `json:"spellcasting_ability" db:"spellcasting_ability"`
	SubtypesName              string        `json:"subtypes_name" db:"subtypes_name"`
	Archetypes                []interface{} `json:"archetypes" db:"archetypes"`
	DocumentSlug              string        `json:"document__slug" db:"document_slug"`
	DocumentTitle             string        `json:"document__title" db:"document_title"`
	DocumentLicenseUrl        string        `json:"document__license_url" db:"document_license_url"`
	DocumentUrl               string        `json:"document__url" db:"document_url"`
}

// for used when comparing json field names with ClassImport field names
func snakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	for i := 0; i < len(parts); i++ {
		parts[i] = strings.Title(parts[i])
	}
	res := strings.Join(parts, "")
	switch res {
	case "Desc":
		return "Description"
	case "HpAt1stLevel":
		return "HpAtFirstLevel"
	case "ProfArmor":
		return "ProficienciesArmor"
	case "ProfWeapons":
		return "ProficienciesWeapons"
	case "ProfTools":
		return "ProficienciesTools"
	case "ProfSavingThrows":
		return "ProficienciesSavingThrows"
	case "ProfSkills":
		return "ProficienciesSkills"
	case "ClassTable":
		return "Table"
	default:
		return res
	}
}

// when importing, we want to examine the field and see if it's
// present on the ClassImport object.  If it's not, we'll decide if
// it should be added or not.  this is a manual process.
func examineResult(result interface{}) {
	classType := reflect.TypeOf(ClassImport{})
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		log.Fatal("could not assert result")
	}
	for key, value := range resultMap {
		camelKey := snakeToCamel(key)
		_, ok := classType.FieldByName(camelKey)

		if !ok && key != "page_no" {
			fmt.Printf("Key/Value pair not found on ClassImport:\nKey: %s, value: %v\n", key, value)
		}
	}
}

func writeClassesToDB(db *sqlx.DB, classes []ClassImport) {

	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS class_imports (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			slug TEXT,
			description TEXT,
			hit_dice TEXT,
			hp_at_first_level TEXT,
			hp_at_higher_levels TEXT,
			proficiencies_armor TEXT,
			proficiencies_weapons TEXT,
			proficiencies_tools TEXT,
			proficiencies_saving_throws TEXT,
			proficiencies_skills TEXT,
			equipment TEXT,
			class_table TEXT,
			spellcasting_ability TEXT,
			subtypes_name TEXT,
			archetypes TEXT,
			document_slug TEXT,
			document_title TEXT,
			document_license_url TEXT,
			document_url TEXT
		);
	`)
	if err != nil {
		log.Fatalf("Failed to create class_imports %v", err)
	}

	for idx := range classes {
		query := `
			INSERT INTO class_imports (
				name, slug, description, hit_dice, hp_at_first_level,
				hp_at_higher_levels, proficiencies_armor, proficiencies_weapons,
				proficiencies_tools, proficiencies_saving_throws, proficiencies_skills,
				equipment, class_table, spellcasting_ability, subtypes_name, archetypes,
				document_slug, document_title, document_license_url, document_url
			)
			VALUES
			(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
		class := classes[idx]
		archetypesJson, err := json.Marshal(class.Archetypes)
		if err != nil {
			log.Fatalf("%v", err)
		}

		_, err = db.Exec(query, class.Name, class.Slug, class.Description, class.HitDice,
			class.HpAtFirstLevel, class.HpAtHigherLevels, class.ProficienciesArmor,
			class.ProficienciesWeapons, class.ProficienciesTools, class.ProficienciesSavingThrows,
			class.ProficienciesSkills, class.Equipment, class.Table, class.SpellcastingAbility,
			class.SubtypesName, archetypesJson, class.DocumentSlug, class.DocumentTitle,
			class.DocumentLicenseUrl, class.DocumentUrl)
		if err != nil {
			log.Fatal(err)
		}

	}

}

func convertJsonToClassImports(jsonData []byte) ([]ClassImport, string) {
	var classes []ClassImport
	var data map[string]interface{}
	err := json.Unmarshal(jsonData, &data)
	if err != nil {
		log.Fatal(err)
	}

	results, ok := data["results"].([]interface{})
	if !ok {
		log.Fatal("Could not assert 'results' as slice")
	}

	for i, result := range results {
		// check if there are keys in the json which are not
		// present as fields on the ClassImport
		examineResult(result)
		resultJson, err := json.Marshal(result)
		if err != nil {
			log.Fatalf("could not marshal result at index %d: %v", i, err)
		}

		var class ClassImport
		err = json.Unmarshal(resultJson, &class)
		if err != nil {
			log.Fatalf("could not decode class at index %d: %v", i, err)
		}
		classes = append(classes, class)
	}
	nextUrl, ok := data["next"].(string)
	if !ok {
		return classes, "DONE"
	}
	return classes, nextUrl
}
