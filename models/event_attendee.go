package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
)

// EventAttendee is used by pop to map your event_attendees database table to your go code.
type EventAttendee struct {
	ID        uuid.UUID `json:"id" db:"id"`
	GuestID   uuid.UUID `db:"guest_id"`
	Guest     *Guest    `belongs_to:"guests"`
	EventID   uuid.UUID `db:"event_id"`
	Event     *Event    `belongs_to:"events"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// String is not required by pop and may be deleted
func (e EventAttendee) String() string {
	je, _ := json.Marshal(e)
	return string(je)
}

// EventAttendees is not required by pop and may be deleted
type EventAttendees []EventAttendee

// String is not required by pop and may be deleted
func (e EventAttendees) String() string {
	je, _ := json.Marshal(e)
	return string(je)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (e *EventAttendee) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateCreate gets run every time you call "pop.ValidateAndCreate" method.
// This method is not required and may be deleted.
func (e *EventAttendee) ValidateCreate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateAndUpdate" method.
// This method is not required and may be deleted.
func (e *EventAttendee) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
