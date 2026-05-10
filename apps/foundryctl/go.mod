module github.com/wrapped-owls/gontainer_foundryvtt/apps/foundryctl

go 1.26.2

require (
	github.com/Masterminds/semver/v3 v3.3.1
	github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryacquire v0.0.0
	github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit v0.0.0
	github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch v0.0.0
	github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime v0.0.0
)

require (
	golang.org/x/net v0.54.0 // indirect
	golang.org/x/sys v0.44.0 // indirect
	golang.org/x/term v0.43.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryacquire => ../../libs/foundryacquire
	github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit => ../../libs/foundrykit
	github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch => ../../libs/foundrypatch
	github.com/wrapped-owls/gontainer_foundryvtt/libs/foundryruntime => ../../libs/foundryruntime
)
