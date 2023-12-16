package actions

import (
	"database/sql"
	"encoding/json"
	"errors"
	"event_planner/models"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
)

// HomeHandler is a default handler to serve up
// a home page.
func HomeHandler(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.HTML("home/index.plush.html"))
}

// EventsListHandler returns GET for list of all events.
func EventsListHandler(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	events := models.Events{}

	err := tx.All(&events)
	if err != nil {
		log.Print(err)
		return c.Redirect(301, "/")
	}

	data, err := json.Marshal(events)
	if err != nil {
		log.Print(err)
		return c.Redirect(301, "/")
	}
	c.Set("events", string(data))
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

	fmt.Printf("got event %s\n\n", event.Title)

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
		log.Printf("error making reservation %s", err)
		return c.Redirect(301, "/")
	}

	c.Flash().Add("info", "Reservation complete")
	return c.Redirect(301, "/events/"+event.ID.String())
}

// AppHandler returns GET for app form.
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

// AppFormHandler responds to POST to add-guest for app form.
func AppFormHandler(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	req := &AppForm{}
	err := c.Bind(req)
	if err != nil {
		log.Printf("form bind error %s", err)
		return c.Redirect(301, "/")
	}

	event := &models.Event{}
	err = tx.Find(event, req.EventID)
	if err != nil {
		log.Printf("error finding event")
		return c.Redirect(301, "/")
	}

	foundGuest := &models.Guest{}
	err = tx.Where("email = ?", req.Email).First(foundGuest)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("guest lookup error %s", err)
		return c.Redirect(301, "/")
	}

	if foundGuest.ID.IsNil() {
		// Try to create new guest.
		foundGuest.Email = req.Email
		foundGuest.FullName = req.FullName

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
		log.Printf("error making reservation %s", err)
		return c.Redirect(301, "/")
	}

	return c.Redirect(301, "/events/"+event.ID.String())
}

// AppRemoteHandler renders the Vue page that makes a remote request to load event list.
func AppRemoteHandler(c buffalo.Context) error {

	return c.Render(http.StatusOK, r.HTML("app-remote"))
}
