module github.com/wrapped-owls/gontainer_foundryvtt/apps/foundrymanager

go 1.26

require (
	github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit v0.0.0
	github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime v0.0.0
)

require (
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/term v0.45.0 // indirect
)

replace (
	github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit => ../../libs/foundrykit
	github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime => ../../libs/foundryruntime
)
