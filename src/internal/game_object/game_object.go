package game_object

type GameObject struct {
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	Sprite  string  `json:"sprite"`
	OffsetX int     `json:"sprite_x0"`
	OffsetY int     `json:"sprite_y0"`
	Width   int     `json:"sprite_x1"`
	Height  int     `json:"sprite_y1"`
	Name    string  `json:"name"`
	Angle   float64 `json:"rotation"`
}
