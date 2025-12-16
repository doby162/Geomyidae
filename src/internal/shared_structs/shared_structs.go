package shared_structs

import (
	"Geomyidae/internal/constants"

	"github.com/jakecoffman/cp/v2"

	"strconv"
)

// A special float that gets rounded to 2 decimal places when marshaled to JSON
// https://stackoverflow.com/a/61811599
type RoundedFloat2 float64

func (r RoundedFloat2) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatFloat(float64(r), 'f', 2, 32)), nil
}

type GameObject struct {
	X                    int                    `json:"x"`
	Y                    int                    `json:"y"`
	Sprite               string                 `json:"s"`
	SpriteOffsetX        int                    `json:"sx0"`
	SpriteOffsetY        int                    `json:"sy0"`
	SpriteWidth          int                    `json:"sx1"`
	SpriteHeight         int                    `json:"sy1"`
	SpriteFlipHorizontal bool                   `json:"sfh"`
	SpriteFlipVertical   bool                   `json:"sfv"`
	SpriteFlipDiagonal   bool                   `json:"sfd"`
	Angle                RoundedFloat2          `json:"rot"`
	UUID                 string                 `json:"id"`
	Delete               bool                   `json:"del"`
	Body                 *cp.Body               `json:"-"`
	Shape                *cp.Shape              `json:"-"`
	NeedsStatics         bool                   `json:"-"`
	IsStatic             bool                   `json:"-"`
	Identity             constants.UserDataCode `json:"-"`
	Inbox                chan string            `json:"-"`
	Portal               bool                   `json:"-"`
}

type HasBehavior interface {
	ApplyBehavior(deltaTime float64, spawnerPipeline chan HasBehavior)
	GetObject() *GameObject
}

type KeyStruct struct {
	Keys []string `json:"keys"`
}

type GameData struct {
	Portal     bool   `json:"portal"`
	PlayerUUID string `json:"pud"`
}

type WorldData struct {
	Objects  []GameObject `json:"objects"`
	GameData GameData     `json:"gd"`
}
