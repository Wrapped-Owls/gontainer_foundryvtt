module github.com/wrapped-owls/gontainer_foundryvtt/apps/taverncord

go 1.26.2

require (
	github.com/bwmarrin/discordgo v0.28.1
	github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager v0.0.0
	github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit v0.0.0
)

require (
	github.com/gorilla/websocket v1.5.3 // indirect
	golang.org/x/crypto v0.45.0 // indirect
	golang.org/x/sys v0.44.0 // indirect
	golang.org/x/term v0.43.0 // indirect
)

replace (
	github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager => ../foundrymanager
	github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit => ../../libs/foundrykit
)
