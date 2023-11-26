package models

import "time"

type AirQuality struct {
	DateTime              time.Time             `json:"dateTime"`
	RegionCode            string                `json:"regionCode"`
	Indexes               []index               `json:"indexes"`
	Pollutants            []pollutant           `json:"pollutants"`
	HealthRecommendations healthRecommendations `json:"healthRecommendations"`
}

type index struct {
	Code              string `json:"code"`
	DisplayName       string `json:"displayName"`
	Aqi               int16  `json:"aqi"`
	AqiDisplay        string `json:"aqiDisplay"`
	Color             color  `json:"color"`
	Category          string `json:"category"`
	DominantPollutant string `json:"dominantPollutant"`
}

type color struct {
	Red   float64 `json:"red"`
	Green float64 `json:"green"`
	Blue  float64 `json:"blue"`
	Alpha float64 `json:"alpha"`
}

type pollutant struct {
	Code           string         `json:"code"`
	DisplayName    string         `json:"displayName"`
	FullName       string         `json:"fullName"`
	Concentration  concentration  `json:"concentration"`
	AdditionalInfo additionalInfo `json:"additionalInfo"`
}

type concentration struct {
	Value float64 `json:"value"`
	Units string  `json:"units"`
}

type additionalInfo struct {
	Sources string `json:"sources"`
	Effects string `json:"effects"`
}

type healthRecommendations struct {
	GeneralPopulation      string `json:"generalPopulation"`
	Elderly                string `json:"elderly"`
	LungDiseasePopulation  string `json:"lungDiseasePopulation"`
	HeartDiseasePopulation string `json:"heartDiseasePopulation"`
	Athletes               string `json:"athletes"`
	PregnantWomen          string `json:"pregnantWomen"`
	Children               string `json:"children"`
}
