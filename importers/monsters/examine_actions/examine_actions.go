package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type MonsterImport struct {
	ID                    int
	Actions               string  `db:"actions"`
	Alignment             string  `db:"alignment"`
	ArmorClass            int32   `db:"armor_class"`
	ArmorDescription      string  `db:"armor_description"`
	BonusActions          []uint8 `db:"bonus_actions"`
	ChallengeRating       float32 `db:"challenge_rating"`
	Charisma              int32   `db:"charisma"`
	CharismaSave          int32   `db:"charisma_save"`
	ConditionImmunities   string  `db:"condition_immunities"`
	Constitution          int32   `db:"constitution"`
	ConstitutionSave      int32   `db:"constitution_save"`
	DamageImmunities      string  `db:"damage_immunities"`
	DamageResistances     string  `db:"damage_resistances"`
	DamageVulnerabilities string  `db:"damage_vulnerabilities"`
	Description           string  `db:"description"`
	Dexterity             int32   `db:"dexterity"`
	DexteritySave         int32   `db:"dexterity_save"`
	DocumentLicenseUrl    string  `db:"document_license_url"`
	DocumentSlug          string  `db:"document_slug"`
	DocumentTitle         string  `db:"document_title"`
	DocumentUrl           string  `db:"document_url"`
	Environments          []uint8 `db:"environments"`
	Group                 string  `db:"group_name"`
	HP                    int32   `db:"hp"`
	HitDice               string  `db:"hit_dice"`
	Image                 string  `db:"image"`
	Intelligence          int32   `db:"intelligence"`
	IntelligenceSave      int32   `db:"intelligence_save"`
	Languages             string  `db:"languages"`
	LegendaryActions      []uint8 `db:"legendary_actions"`
	LegendaryDescription  string  `db:"legendary_description"`
	Name                  string  `db:"name"`
	Perception            int32   `db:"perception"`
	Reactions             []uint8 `db:"reactions"`
	Senses                string  `db:"senses"`
	Size                  string  `db:"size"`
	Skills                []uint8 `db:"skills"`
	Slug                  string  `db:"slug"`
	SpecialAbilities      []uint8 `db:"special_abilities"`
	Speed                 []uint8 `db:"speed"`
	SpellList             []uint8 `db:"spell_list"`
	Strength              int32   `db:"strength"`
	StrengthSave          int32   `db:"strength_save"`
	Subtype               string  `db:"subtype"`
	Type                  string  `db:"type"`
	Wisdom                int32   `db:"wisdom"`
	WisdomSave            int32   `db:"wisdom_save"`
}

func main() {
	// Connect to the database:
	db, err := sqlx.Open("sqlite3", "../../sql_database/monster_imports.db")
	if err != nil {
		log.Fatalln(err)
	}

	// Query all monsters:
	monsters := []MonsterImport{}
	err = db.Select(&monsters, "SELECT * FROM mob_imports")
	if err != nil {
		log.Fatalln(err)
	}

	// Open a file for writing:
	file, err := os.Create("actions.txt")
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	// Write all actions to the file:
	for _, monster := range monsters {
		var result interface{}
		if monster.Actions != "null" {
			err := json.Unmarshal([]byte(monster.Actions), &result)
			if err != nil {
				log.Fatalf("%v", err)
			}
			data, ok := result.([]interface{})
			if !ok {
				log.Fatalf("json is not an array")
			}
			for i, item := range data {
				obj, ok := item.(map[string]interface{})
				if !ok {
					log.Fatalf("item %d is not an object", i)
				}
				desc, ok := obj["desc"].(string)
				if !ok {
					log.Fatalf("desc is not a string")
				}
				_, err = file.WriteString(fmt.Sprintf("%s\n", desc))
				if err != nil {
					log.Fatalln(err)
				}
			}
		}
	}
}
