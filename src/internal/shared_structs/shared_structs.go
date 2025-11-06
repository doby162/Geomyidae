package shared_structs

import "github.com/jakecoffman/cp/v2"

type GameObject struct {
	X                    float64   `json:"x"`
	Y                    float64   `json:"y"`
	Sprite               string    `json:"sprite"`
	SpriteOffsetX        int       `json:"sprite_x0"`
	SpriteOffsetY        int       `json:"sprite_y0"`
	SpriteWidth          int       `json:"sprite_x1"`
	SpriteHeight         int       `json:"sprite_y1"`
	SpriteFlipHorizontal bool      `json:"sprite_flip_horizontal"`
	SpriteFlipVertical   bool      `json:"sprite_flip_vertical"`
	SpriteFlipDiagonal   bool      `json:"sprite_flip_diagonal"`
	Angle                float64   `json:"rotation"`
	UUID                 string    `json:"uuid"`
	Body                 *cp.Body  `json:"-"`
	Shape                *cp.Shape `json:"-"`
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
