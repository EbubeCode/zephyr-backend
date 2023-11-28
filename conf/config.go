package conf

import (
	"github.com/tkanos/gonfig"
	"log"
)

type Configuration struct {
	VERSION              string
	AIR_QUALITY_BASE_URL string
	PLACES_BASE_URL      string
	API_KEY              string
	DEFAULT_LONGITUDE    float64
	DEFAULT_LATITUDE     float64
	SUPPORTED_COUNTRIES  []string
}

func GetConfig() Configuration {
	configuration := Configuration{}
	err := gonfig.GetConf("./conf/config.json", &configuration)
	if err != nil {
		log.Printf("Error loading config: %s", err)
	}

	return configuration
}
