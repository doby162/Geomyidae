package constants

type UserDataCode string

const (
	Unknown UserDataCode = "" // a zero value should only show up if someone didn't initialize their GameObject correctly
	Player  UserDataCode = "player"
	Turret  UserDataCode = "turret"
	Bullet  UserDataCode = "bullet"
	Tile    UserDataCode = "tile"
	Tracker UserDataCode = "tracker"
	Bomb    UserDataCode = "bomb"
	Pickup  UserDataCode = "pickup"
)
