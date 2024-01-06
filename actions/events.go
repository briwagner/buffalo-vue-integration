package actions

import (
	"database/sql"
	"encoding/json"
	"event_planner/models"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

// EventsListHandler returns GET for list of all events.
func EventsListHandler(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	events := models.Events{}

	err := tx.All(&events)
	if err != nil {
		log.Print(err)
		return c.Redirect(301, "/")
	}

	ct, _ := c.Value("contentType").(string)
	if ct == "application/json" {
		return c.Render(http.StatusOK, r.JSON(events))
	}

	// Marshal to JSON so the Vue app can read it.
	data, err := json.Marshal(events)
	if err != nil {
		log.Print(err)
		return c.Redirect(301, "/")
	}
	c.Set("events", string(data))
	c.Set("eventsGo", events)
	return c.Render(http.StatusOK, r.HTML("events/all"))
}

// EventsListHandler returns JSON list of all events.
func EventsListJSONHandler(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	events := models.Events{}

	err := tx.All(&events)
	if err != nil {
		log.Print(err)
		return c.Redirect(301, "/")
	}

	// r.JSON handles Marshal for us.
	return c.Render(http.StatusOK, r.JSON(events))
}

// EventDetailHandler returns GET for detail on one event.
func EventDetailHandler(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	event := models.Event{}
	eventID := c.Param("id")

	err := tx.Eager().Find(&event, eventID)
	if err != nil {
		log.Print(err)
		return c.Redirect(301, "/")
	}

	g := &models.Guest{} // for partial form
	c.Set("guest", g)
	c.Set("event", event)
	return c.Render(http.StatusOK, r.HTML("events/detail"))
}

// EventNewHandler returns GET for create form.
func EventNewHandler(c buffalo.Context) error {
	e := models.Event{}
	e.Date = time.Now()
	c.Set("event", e)
	c.Set("tFormat", "2006-01-02T15:04")
	return c.Render(http.StatusOK, r.HTML("events/new"))
}

// EventCreateHandler responds to POST to create new event.
func EventCreateHandler(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	event := &models.Event{}

	event.Date = time.Now()
	event.EventGuests = make([]models.Guest, 0)

	err := c.Bind(event)
	if err != nil {
		fmt.Printf("bad bind %s", err)
		return c.Redirect(301, "/")
	}

	err = tx.Create(event)
	if err != nil {
		fmt.Printf("bad create %s", err)
		return c.Redirect(301, "/")
	}

	c.Flash().Add("info", "Event created")
	return c.Redirect(301, "/events/"+event.ID.String())
}

// EventNewGuestHandler returns GET for add-guest form.
func EventNewGuestHandler(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	e := models.Event{}
	g := models.Guest{}

	eventID := c.Param("id")
	err := tx.Find(&e, eventID)
	if err != nil {
		log.Printf("error finding event %s", err)
		return c.Redirect(301, "/")
	}

	c.Set("event", e)
	c.Set("guest", g)
	return c.Render(http.StatusOK, r.HTML("events/add-guest"))
}

// EventAddGuestHandler responds to POST to create add-guest.
func EventAddGuestHandler(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	event := &models.Event{}
	eventID := c.Param("id")

	// Find and validate event.
	err := tx.Find(event, eventID)
	if err != nil {
		log.Printf("event error %s", err)
		return c.Redirect(301, "/")
	}

	// Find and validate guest.
	guest := &models.Guest{}
	err = c.Bind(guest)
	if err != nil {
		log.Printf("form error %s", err)
		return c.Redirect(301, "/")
	}

	foundGuest := &models.Guest{}
	err = tx.Where("email = ?", guest.Email).First(foundGuest)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("guest lookup error %s", err)
		return c.Redirect(301, "/")
	}

	if foundGuest.ID.IsNil() {
		// Try to create new guest.
		foundGuest.Email = guest.Email
		foundGuest.FullName = guest.FullName

		err = tx.Create(foundGuest)
		if err != nil {
			log.Printf("error creating guest %s", err)
			return c.Redirect(301, "/")
		}
	}

	res := &models.EventAttendee{}
	res.GuestID = foundGuest.ID
	res.EventID = event.ID

	err = tx.Create(res)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			c.Flash().Add("warning", "A reservation already exists for that person.")
			return c.Redirect(301, "/events/"+event.ID.String())
		}
		log.Printf("error making reservation %s", err)
		return c.Redirect(301, "/")
	}

	c.Flash().Add("info", "Reservation complete for "+foundGuest.Email)
	return c.Redirect(301, "/events/"+event.ID.String())
}

// AppHandler returns GET for Vue form.
func AppHandler(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	events := &models.Events{}

	err := tx.All(events)
	if err != nil {
		log.Printf("error getting events %s", err)
		return c.Redirect(301, "/")
	}

	eventData, err := json.Marshal(events)
	if err != nil {
		log.Print(err)
		return c.Redirect(301, "/")
	}
	c.Set("events", string(eventData))
	return c.Render(http.StatusOK, r.HTML("app"))
}

type AppForm struct {
	EventID  uuid.UUID `form:"EventID"`
	FullName string    `form:"FullName"`
	Email    string    `form:"Email"`
}

// AppFormHandler responds to POST to add-guest for Vue form.
func AppFormHandler(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	req := &AppForm{}
	err := c.Bind(req)
	if err != nil {
		log.Printf("form bind error %s", err)
		return c.Render(400, r.String("form binding error"))
	}

	event := &models.Event{}
	err = tx.Find(event, req.EventID)
	if err != nil {
		log.Printf("error finding event")
		return c.Render(404, r.String("event not found "+req.EventID.String()))
	}

	foundGuest := &models.Guest{}
	err = tx.Where("email = ?", req.Email).First(foundGuest)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("guest lookup error %s", err)
		return c.Render(404, r.String("guest not found with email: "+req.Email))
	}

	if foundGuest.ID.IsNil() {
		// Try to create new guest.
		foundGuest.Email = req.Email
		foundGuest.FullName = req.FullName

		err = tx.Create(foundGuest)
		if err != nil {
			log.Printf("error creating guest %s", err)
			return c.Render(500, r.String("error creating guest"))
		}
	}

	res := &models.EventAttendee{}
	res.GuestID = foundGuest.ID
	res.EventID = event.ID

	err = tx.Create(res)
	if err != nil {
		log.Printf("error making reservation %s", err)
		return c.Render(500, r.String("error making reservation"))
	}

	log.Printf("reservation made for %s", event.ID.String())
	return c.Render(204, r.String(""))
}

// EventsRemoteHandler renders the Vue page that makes a remote request to load event list.
func EventsRemoteHandler(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.HTML("events/list-remote"))
}
