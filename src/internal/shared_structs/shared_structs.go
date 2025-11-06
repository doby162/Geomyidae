package shared_structs

import "github.com/jakecoffman/cp/v2"

type GameObject struct {
	X       float64   `json:"x"`
	Y       float64   `json:"y"`
	Sprite  string    `json:"sprite"`
	OffsetX int       `json:"sprite_x0"`
	OffsetY int       `json:"sprite_y0"`
	Width   int       `json:"sprite_x1"`
	Height  int       `json:"sprite_y1"`
	Angle   float64   `json:"rotation"`
	UUID    string    `json:"uuid"`
	Body    *cp.Body  `json:"-"`
	Shape   *cp.Shape `json:"-"`
}

type KeyStruct struct {
	Keys []string `json:"keys"`
}

// Objects gets turned into a map later so... maybe it should just be a map
// todo I guess
type WorldData struct {
	Name    string       `json:"name"`
	Objects []GameObject `json:"objects"`
}
