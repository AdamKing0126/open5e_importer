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
	nextUrl := "https://api.open5e.com/v1/monsters/"

	db, err := sqlx.Open("sqlite3", "../mud/sql_database/monster_imports.db")
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

		monsters, next := convertJsonToMonsterImports(bodyBytes)
		writeMonstersToDB(db, monsters)

		if next == "DONE" {
			break
		}

		nextUrl = next
		fmt.Printf("next url to fetch: %s\n", nextUrl)
	}
}

type Open5eResponse struct {
	Count    int             `json:"count"`
	Next     string          `json:"next"`
	Previous string          `json:"previous"`
	Results  []MonsterImport `json:"results"`
}

type MonsterImport struct {
	Actions               []interface{}            `json:"actions"`
	Alignment             string                   `json:"alignment"`
	ArmorClass            int32                    `json:"armor_class"`
	ArmorDescription      string                   `json:"armor_desc"`
	BonusActions          []interface{}            `json:"bonus_actions"`
	ChallengeRating       float32                  `json:"cr"`
	Charisma              int32                    `json:"charisma"`
	CharismaSave          int32                    `json:"charisma_save"`
	ConditionImmunities   string                   `json:"condition_immunities"`
	Constitution          int32                    `json:"constitution"`
	ConstitutionSave      int32                    `json:"constitution_save"`
	DamageImmunities      string                   `json:"damage_immunities"`
	DamageResistances     string                   `json:"damage_resistances"`
	DamageVulnerabilities string                   `json:"damage_vulnerabilities"`
	Description           string                   `json:"desc"`
	Dexterity             int32                    `json:"dexterity"`
	DexteritySave         int32                    `json:"dexterity_save"`
	DocumentLicenseUrl    string                   `json:"document__license_url"`
	DocumentSlug          string                   `json:"document__slug"`
	DocumentTitle         string                   `json:"document__title"`
	DocumentUrl           string                   `json:"document__url"`
	Environments          []interface{}            `json:"environments"`
	Group                 string                   `json:"group"`
	HP                    int32                    `json:"hit_points"`
	HitDice               string                   `json:"hit_dice"`
	Image                 string                   `json:"img_main"`
	Intelligence          int32                    `json:"intelligence"`
	IntelligenceSave      int32                    `json:"intelligence_save"`
	Languages             string                   `json:"languages"`
	LegendaryActions      []interface{}            `json:"legendary_actions"`
	LegendaryDescription  string                   `json:"legendary_desc"`
	Name                  string                   `json:"name"`
	Perception            int32                    `json:"perception"`
	Reactions             []interface{}            `json:"reactions"`
	Senses                string                   `json:"senses"`
	Size                  string                   `json:"size"`
	Skills                map[string]interface{}   `json:"skills"`
	Slug                  string                   `json:"slug"`
	SpecialAbilities      []map[string]interface{} `json:"special_abilities"`
	Speed                 interface{}              `json:"speed"`
	SpellList             []string                 `json:"spell_list"`
	Strength              int32                    `json:"strength"`
	StrengthSave          int32                    `json:"strength_save"`
	Subtype               string                   `json:"subtype"`
	Type                  string                   `json:"type"`
	Wisdom                int32                    `json:"wisdom"`
	WisdomSave            int32                    `json:"wisdom_save"`
}

// for used when comparing json field names with MonsterImport field names
func snakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	for i := 0; i < len(parts); i++ {
		parts[i] = strings.Title(parts[i])
	}
	res := strings.Join(parts, "")
	switch res {
	case "Cr":
		return "ChallengeRating"
	case "LegendaryDesc":
		return "LegendaryDescription"
	case "HitPoints":
		return "HP"
	case "ImgMain":
		return "Image"
	case "ArmorDesc":
		return "ArmorDescription"
	case "Desc":
		return "Description"
	default:
		return res
	}
}

// when importing, we want to examine the field and see if it's
// present on the MonsterImport object.  If it's not, we'll decide if
// it should be added or not.  this is a manual process.
func examineResult(result interface{}) {
	monsterType := reflect.TypeOf(MonsterImport{})
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		log.Fatal("could not assert result")
	}
	for key, value := range resultMap {
		camelKey := snakeToCamel(key)
		_, ok := monsterType.FieldByName(camelKey)

		if !ok && key != "page_no" {
			fmt.Printf("Key/Value pair not found on MonsterImport:\nKey: %s, value: %v\n", key, value)
		}
	}
}

func writeMonstersToDB(db *sqlx.DB, monsters []MonsterImport) {

	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS mob_imports (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			actions TEXT,
			alignment TEXT,
			armor_class INTEGER,
			armor_description TEXT,
			bonus_actions TEXT,
			challenge_rating FLOAT,
			charisma INTEGER,
			charisma_save INTEGER,
			condition_immunities TEXT,
			constitution INTEGER,
			constitution_save INTEGER,
			damage_immunities TEXT,
			damage_resistances TEXT,
			damage_vulnerabilities TEXT,
			description TEXT,
			dexterity INTEGER,
			dexterity_save INTEGER,
			document_license_url TEXT,
			document_slug TEXT,
			document_title TEXT,
			document_url TEXT,
			environments TEXT,
			group_name TEXT,
			hp INTEGER,
			hit_dice TEXT,
			image TEXT,
			intelligence INTEGER,
			intelligence_save INTEGER,
			languages TEXT,
			legendary_actions TEXT,
			legendary_description TEXT,
			name TEXT,
			perception INTEGER,
			reactions TEXT,
			senses TEXT,
			size TEXT,
			skills TEXT,
			slug TEXT,
			special_abilities TEXT,
			speed TEXT,
			spell_list TEXT,
			strength INTEGER,
			strength_save INTEGER,
			subtype TEXT,
			type TEXT,
			wisdom INTEGER,
			wisdom_save INTEGER
		);
	`)
	if err != nil {
		log.Fatalf("Failed to create mob_imports %v", err)
	}

	for _, monster := range monsters {
		query := `
			INSERT INTO mob_imports
			(
				actions, 
				alignment, 
				armor_class,
				armor_description,
				bonus_actions,
				challenge_rating,
				charisma,
				charisma_save,
				condition_immunities,
				constitution,
				constitution_save,
				damage_immunities,
				damage_resistances,
				damage_vulnerabilities,
				description,
				dexterity,
				dexterity_save,
				document_license_url,
				document_slug,
				document_title,
				document_url,
				environments,
				group_name,
				hp,
				hit_dice,
				image,
				intelligence,
				intelligence_save,
				languages,
				legendary_actions,
				legendary_description,
				name,
				perception,
				reactions,
				senses,
				size,
				skills,
				slug,
				special_abilities,
				speed,spell_list,
				strength,
				strength_save,
				subtype,
				type,
				wisdom,
				wisdom_save)
			VALUES
			(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		`
		actionsJson, err := json.Marshal(monster.Actions)
		if err != nil {
			log.Fatalf("%v", err)
		}

		bonusActionsJson, err := json.Marshal(monster.BonusActions)
		if err != nil {
			log.Fatalf("%v", err)
		}

		environmentsJson, err := json.Marshal(monster.Environments)
		if err != nil {
			log.Fatalf("%v", err)
		}

		legendaryActionsJson, err := json.Marshal(monster.LegendaryActions)
		if err != nil {
			log.Fatalf("%v", err)
		}

		reactionsJson, err := json.Marshal(monster.Reactions)
		if err != nil {
			log.Fatalf("%v", err)
		}

		skillsJson, err := json.Marshal(monster.Skills)
		if err != nil {
			log.Fatalf("%v", err)
		}

		specialAbilitiesJson, err := json.Marshal(monster.SpecialAbilities)
		if err != nil {
			log.Fatalf("%v", err)
		}

		speedJson, err := json.Marshal(monster.Speed)
		if err != nil {
			log.Fatalf("%v", err)
		}

		spellListJson, err := json.Marshal(monster.SpellList)
		if err != nil {
			log.Fatalf("%v", err)
		}

		languagesJson, err := json.Marshal(monster.Languages)
		if err != nil {
			log.Fatalf("%v", err)
		}

		_, err = db.Exec(query,
			actionsJson,
			monster.Alignment,
			monster.ArmorClass,
			monster.ArmorDescription,
			bonusActionsJson,
			monster.ChallengeRating,
			monster.Charisma,
			monster.CharismaSave,
			monster.ConditionImmunities,
			monster.Constitution,
			monster.ConstitutionSave,
			monster.DamageImmunities,
			monster.DamageResistances,
			monster.DamageVulnerabilities,
			monster.Description,
			monster.Dexterity,
			monster.DexteritySave,
			monster.DocumentLicenseUrl,
			monster.DocumentSlug,
			monster.DocumentTitle,
			monster.DocumentUrl,
			environmentsJson,
			monster.Group,
			monster.HP,
			monster.HitDice,
			monster.Image,
			monster.Intelligence,
			monster.IntelligenceSave,
			languagesJson,
			legendaryActionsJson,
			monster.LegendaryDescription,
			monster.Name,
			monster.Perception,
			reactionsJson,
			monster.Senses,
			monster.Size,
			skillsJson,
			monster.Slug,
			specialAbilitiesJson,
			speedJson,
			spellListJson,
			monster.Strength,
			monster.StrengthSave,
			monster.Subtype,
			monster.Type,
			monster.Wisdom,
			monster.WisdomSave)
		if err != nil {
			log.Fatal(err)
		}

	}

}

func convertJsonToMonsterImports(jsonData []byte) ([]MonsterImport, string) {
	var monsters []MonsterImport
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
		// present as fields on the MonsterImport
		examineResult(result)
		resultJson, err := json.Marshal(result)
		if err != nil {
			log.Fatalf("could not marshal result at index %d: %v", i, err)
		}

		var monster MonsterImport
		err = json.Unmarshal(resultJson, &monster)
		if err != nil {
			log.Fatalf("could not decode monster at index %d: %v", i, err)
		}
		monsters = append(monsters, monster)
	}
	nextUrl, ok := data["next"].(string)
	if !ok {
		return monsters, "DONE"
	}
	return monsters, nextUrl
}
