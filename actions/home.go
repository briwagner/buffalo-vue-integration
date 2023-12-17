package actions

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
)

// HomeHandler shows app routes.
func HomeHandler(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.HTML("home/index.plush.html"))
}
