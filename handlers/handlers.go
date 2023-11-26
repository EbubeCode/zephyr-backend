package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Stutern-128/backend/conf"
	"github.com/Stutern-128/backend/models"
	"github.com/gofiber/fiber/v2"
	"googlemaps.github.io/maps"
	"log"
	"strings"
	"time"
)

// App holds the application state
type App struct {
	MapsClient *maps.Client
	Config     *conf.Configuration
}

type LocationRequest struct {
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	ChartRange string  `json:"chart_range"`
	TimeZone   string  `json:"timeZone"`
}

func (r *LocationRequest) initializeDefaults(config *conf.Configuration) {
	if r.Latitude == 0 || r.Longitude == 0 {
		r.Latitude = config.DEFAULT_LATITUDE
		r.Longitude = config.DEFAULT_LONGITUDE
		r.TimeZone = "America/Los_Angeles"
	}
	r.ChartRange = strings.ToLower(r.ChartRange)
	if r.ChartRange != "day" && strings.ToLower(r.ChartRange) != "week" {
		r.ChartRange = "day"
	}
	// Use the "America/Los_Angeles" timezone for California
	_, err := time.LoadLocation(r.TimeZone)
	if err != nil {
		log.Println("Error loading location:", err)
		r.TimeZone = "America/Los_Angeles"
	}

}

func (r *LocationRequest) getHours() int {
	location, _ := time.LoadLocation(r.TimeZone)

	// Get the current time in the specified timezone
	currentTime := time.Now().In(location)

	// Get the current hour
	if r.ChartRange == "day" {
		return currentTime.Hour()
	}
	return 168
}

func isSupportedCountry(element string) bool {
	supportedCountryCodes := []string{
		"al", "as", "ad", "ar", "am", "au", "at", "az", "bs", "bh", "bd", "by", "be", "ba", "br", "bn", "bg", "ca", "cl", "cn", "co",
		"cr", "hr", "cy", "cz", "dk", "ec", "eg", "ee", "et", "fi", "fr", "ge", "de", "gi", "gr", "gu", "gg", "hk", "hu", "in", "id",
		"ie", "il", "it", "jp", "je", "jo", "ke", "kr", "kw", "lv", "li", "lt", "lu", "my", "mt", "mu", "mx", "md", "mn", "me", "ma",
		"np", "nl", "nz", "mk", "no", "pk", "pe", "ph", "pl", "pt", "pr", "qa", "re", "ro", "ru", "sa", "rs", "sg", "sk", "si", "za",
		"es", "lk", "se", "ch", "tw", "th", "tr", "ug", "ua", "ae", "gb", "us",
	}
	for _, code := range supportedCountryCodes {
		if code == strings.ToLower(element) {
			return true
		}
	}
	return false
}

func (app *App) HandleGetAQI() func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var request LocationRequest
		if err := c.BodyParser(&request); err != nil {
			log.Printf("Invalid request payload: %s\n", err)
		}
		var address string
		if request.Latitude == 0 || request.Longitude == 0 {
			request.initializeDefaults(app.Config)
			address = "CW98+VV Mountain View, CA, USA"
		} else {
			// Perform reverse geocoding
			reverseGeocodeRequest := &maps.GeocodingRequest{
				LatLng: &maps.LatLng{
					Lat: request.Latitude,
					Lng: request.Longitude,
				},
			}

			reverseGeocodeResult, err := app.MapsClient.ReverseGeocode(context.Background(), reverseGeocodeRequest)
			if err != nil || len(reverseGeocodeResult) <= 0 {
				log.Printf("Error decode location: %s\n", err)
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"success": false,
					"error":   "Location not found",
				})
			}
			var countryCode string
			for _, addressComponent := range reverseGeocodeResult[0].AddressComponents {
				for _, typeValue := range addressComponent.Types {
					if typeValue == "country" {
						countryCode = addressComponent.ShortName
						break
					}
				}
				if countryCode != "" {
					break
				}
			}
			address = reverseGeocodeResult[0].FormattedAddress
			log.Printf("Location: %s\n", address)

			if !isSupportedCountry(countryCode) {
				log.Printf("Unsupported location %s\n", err)
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"success": false,
					"error":   "Location not supported",
				})
			}

		}

		url := fmt.Sprintf("%scurrentConditions:lookup?key=%s", app.Config.AIR_QUALITY_BASE_URL, app.Config.API_KEY)
		agent := fiber.Post(url)
		extraComputations := [1]string{"DOMINANT_POLLUTANT_CONCENTRATION"}
		agent.JSON(fiber.Map{
			"location": fiber.Map{
				"longitude": request.Longitude,
				"latitude":  request.Latitude,
			},
			"extraComputations": extraComputations,
		}) // set body received by request
		statusCode, body, errs := agent.Bytes()
		if len(errs) > 0 || statusCode != 200 {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"errs": errs,
			})
		}
		var airQuality models.AirQuality
		err := json.Unmarshal(body, &airQuality)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"err": err,
			})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"dateTime":                       airQuality.DateTime,
			"regionCode":                     airQuality.RegionCode,
			"aqiCode":                        airQuality.Indexes[0].Code,
			"aqiDisplayName":                 airQuality.Indexes[0].DisplayName,
			"aqiValue":                       airQuality.Indexes[0].Aqi,
			"aqiValueDisplay":                airQuality.Indexes[0].AqiDisplay,
			"aqiColor":                       airQuality.Indexes[0].Color,
			"aqiCategory":                    airQuality.Indexes[0].Category,
			"dominantPollutantCode":          airQuality.Indexes[0].DominantPollutant,
			"dominantPollutantDisplayName":   airQuality.Pollutants[0].DisplayName,
			"dominantPollutantFullName":      airQuality.Pollutants[0].FullName,
			"dominantPollutantConcentration": airQuality.Pollutants[0].Concentration,
			"dominantPollutantAdditionInfo":  airQuality.Pollutants[0].AdditionalInfo,
			"location":                       address,
		})
	}
}

func (app *App) HandleChart() func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		var request LocationRequest
		if err := c.BodyParser(&request); err != nil {
			log.Printf("Invalid request payload: %s\n", err)
		}
		if request.Latitude > 0 && request.Longitude > 0 {
			// Perform reverse geocoding
			reverseGeocodeRequest := &maps.GeocodingRequest{
				LatLng: &maps.LatLng{
					Lat: request.Latitude,
					Lng: request.Longitude,
				},
			}

			reverseGeocodeResult, err := app.MapsClient.ReverseGeocode(context.Background(), reverseGeocodeRequest)
			if err != nil || len(reverseGeocodeResult) <= 0 {
				log.Printf("Error decode location: %s\n", err)
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"success": false,
					"error":   "Location not found",
				})
			}
			var countryCode string
			for _, addressComponent := range reverseGeocodeResult[0].AddressComponents {
				for _, typeValue := range addressComponent.Types {
					if typeValue == "country" {
						countryCode = addressComponent.ShortName
						break
					}
				}
				if countryCode != "" {
					break
				}
			}
			if !isSupportedCountry(countryCode) {
				log.Printf("Unsupported location %s\n", err)
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"success": false,
					"error":   "Location not supported",
				})
			}
		}
		request.initializeDefaults(app.Config)
		extraComputations := []string{"DOMINANT_POLLUTANT_CONCENTRATION", "POLLUTANT_CONCENTRATION"}

		var airQualities models.AirQualities
		var aqiValues []interface{}
		var dominantPollutantValues []interface{}
		var totalAqi int
		var totalDominantPollutantConcentration float64
		var firstAqiValue float64
		var lastAqiValue float64
		size := request.getHours()
		for size > 0 {
			url := fmt.Sprintf("%shistory:lookup?key=%s", app.Config.AIR_QUALITY_BASE_URL, app.Config.API_KEY)
			agent := fiber.Post(url)
			agent.JSON(fiber.Map{
				"location": fiber.Map{
					"longitude": &request.Longitude,
					"latitude":  &request.Latitude,
				},
				"extraComputations": extraComputations,
				"hours":             request.getHours(),
				"pageToken":         airQualities.NextPageToken,
				"pageSize":          72,
			}) // set body received by request
			statusCode, body, errs := agent.Bytes()
			if len(errs) > 0 || statusCode != 200 {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"errs": errs,
				})
			}
			err := json.Unmarshal(body, &airQualities)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"err": err,
				})
			}
			for _, airQuality := range airQualities.HoursInfo {
				totalAqi += airQuality.Indexes[0].Aqi
				totalDominantPollutantConcentration += airQuality.Pollutants[0].Concentration.Value
				aqiValues = append(aqiValues, fiber.Map{
					"dateTime":        airQuality.DateTime,
					"aqiCode":         airQuality.Indexes[0].Code,
					"aqiDisplayName":  airQuality.Indexes[0].DisplayName,
					"aqiValue":        airQuality.Indexes[0].Aqi,
					"aqiValueDisplay": airQuality.Indexes[0].AqiDisplay,
				})
				dominantPollutantValues = append(dominantPollutantValues, fiber.Map{
					"dominantPollutantCode":          airQuality.Indexes[0].DominantPollutant,
					"dominantPollutantDisplayName":   airQuality.Pollutants[0].DisplayName,
					"dominantPollutantConcentration": airQuality.Pollutants[0].Concentration.AddSymbol(),
				})
			}
			if firstAqiValue == 0 {
				firstAqiValue = float64(airQualities.HoursInfo[0].Indexes[0].Aqi)
			}
			lastAqiValue = float64(airQualities.HoursInfo[len(airQualities.HoursInfo)-1].Indexes[0].Aqi)
			size -= len(airQualities.HoursInfo)
		}

		changeInAqi := (firstAqiValue - lastAqiValue) / firstAqiValue
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"aqis":                          aqiValues,
			"dominantPollutants":            dominantPollutantValues,
			"averageAqiValue":               totalAqi / len(airQualities.HoursInfo),
			"averageDominantPollutantValue": totalDominantPollutantConcentration / float64(len(airQualities.HoursInfo)),
			"percentageChangeInAqi":         changeInAqi * 100,
		})
	}
}
