package main

import (
	"github.com/Stutern-128/backend/conf"
	"github.com/Stutern-128/backend/handlers"
	_ "github.com/Stutern-128/backend/handlers"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"googlemaps.github.io/maps"
	"log"
)

func main() {
	app := fiber.New()
	config := conf.GetConfig()

	// Initialize the Maps client during the application startup
	appInstance := &handlers.App{
		MapsClient: createMapsClient(&config),
		Config:     &config,
	}

	app.Use(cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			return true
		},
	}))

	app.Post("/aqi", appInstance.HandleGetAQI())
	app.Post("/pollutants", appInstance.HandleGetPollutants())
	app.Post("/pollutantsAdditionalInfo", appInstance.HandleGetPollutantsAdditionalInfo())
	app.Post("/nearbyPlaces", appInstance.HandleNearByPlaces())
	app.Post("/searchPlaces", appInstance.HandleSearch())
	app.Post("/chart", appInstance.HandleChart())
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendFile("./index.html")
	})

	err := app.Listen(":3000")
	if err != nil {
		panic(err)
	}
}

// createMapsClient initializes and returns a Google Maps client
func createMapsClient(config *conf.Configuration) *maps.Client {
	client, err := maps.NewClient(maps.WithAPIKey(config.API_KEY))
	if err != nil {
		log.Fatal(err)
	}
	return client
}
