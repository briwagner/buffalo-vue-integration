package grifts

import (
	"event_planner/models"
	"time"

	"github.com/gobuffalo/grift/grift"
)

var _ = grift.Namespace("db", func() {

	grift.Desc("seed", "Seeds a database")
	grift.Add("seed", func(c *grift.Context) error {
		// Add DB seeding stuff here

		e1 := models.Event{Title: "Demo event", Description: "Demo event description", Date: time.Now()}
		_, err := models.DB.ValidateAndCreate(&e1)
		if err != nil {
			return err
		}
		return nil
	})

})
