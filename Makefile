help:
	@echo "import_races"
	@echo "import_classes"
	@echo "import_monsters"

import_races:
	go run ./importers/races/main.go

import_classes:
	go run ./importers/classes/main.go

import_monsters:
	go run ./importers/monsters/main.go

examine_actions:
	go run ./importers/examine_actions/examine_actions.go
