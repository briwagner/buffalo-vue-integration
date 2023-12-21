package models

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
)

// Event is used by pop to map your events database table to your go code.
type Event struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Title       string    `db:"title"`
	Description string    `db:"desc"`
	Date        time.Time `db:"event_date"`
	EventGuests Guests    `many_to_many:"event_attendees"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// String is not required by pop and may be deleted
func (e Event) String() string {
	je, _ := json.Marshal(e)
	return string(je)
}

func (e Event) HasGuests() bool {
	return len(e.EventGuests) > 0
}

func (e Event) EventDate() string {
	return e.Date.Format("Jan. 02 2006 3:04 PM MST")
}

func (e Event) ToListItem() string {
	return "<li><a href='/events/" + e.ID.String() + "'>" + e.Title + "</a> &mdash; " + e.EventDate() + "</li>"
}

// Events is not required by pop and may be deleted
type Events []Event

// String is not required by pop and may be deleted
func (e Events) String() string {
	je, _ := json.Marshal(e)
	return string(je)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (e *Event) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (e *Event) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (e *Event) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

func (e *Events) SearchTitle(tx *pop.Connection, s string) error {
	// TODO do we have to concat the string like this?
	s = s + "%"
	err := tx.Where("title like ?", s).All(e)
	if err != nil {
		return err
	}
	return nil
}

func (e Events) ToList() string {
	var b strings.Builder
	b.WriteString("<ul>")
	for _, ev := range e {
		b.WriteString(ev.ToListItem())
	}
	b.WriteString("</ul>")
	return b.String()
}
