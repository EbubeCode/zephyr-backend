package conf

import (
	"github.com/tkanos/gonfig"
)

type Configuration struct {
	Version           string
	ApiKey            string
	AirQualityBaseUrl int
}

func GetConfig() Configuration {
	configuration := Configuration{}
	gonfig.GetConf("./conf/config.json", &configuration)

	return configuration
}
