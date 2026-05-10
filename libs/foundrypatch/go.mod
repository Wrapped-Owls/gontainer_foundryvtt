module github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrypatch

go 1.26.2

require (
	github.com/Masterminds/semver/v3 v3.3.1
	github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit v0.0.0
	gopkg.in/yaml.v3 v3.0.1
)

replace github.com/wrapped-owls/gontainer_foundryvtt/libs/foundrykit => ../foundrykit
