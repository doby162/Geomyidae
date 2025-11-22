package shared_structs

import (
	"Geomyidae/internal/constants"

	"github.com/jakecoffman/cp/v2"
)

type GameObject struct {
	X                    float64                `json:"x"`
	Y                    float64                `json:"y"`
	Sprite               string                 `json:"s"`
	SpriteOffsetX        int                    `json:"sx0"`
	SpriteOffsetY        int                    `json:"sy0"`
	SpriteWidth          int                    `json:"sx1"`
	SpriteHeight         int                    `json:"sy1"`
	SpriteFlipHorizontal bool                   `json:"sfh"`
	SpriteFlipVertical   bool                   `json:"sfv"`
	SpriteFlipDiagonal   bool                   `json:"sfd"`
	Angle                float64                `json:"rot"`
	UUID                 string                 `json:"id"`
	Body                 *cp.Body               `json:"-"`
	Shape                *cp.Shape              `json:"-"`
	Delete               bool                   `json:"del"`
	ShootFlag            bool                   `json:"-"`
	NeedsStatics         bool                   `json:"-"`
	IsStatic             bool                   `json:"-"`
	BombDrop             bool                   `json:"-"`
	Identity             constants.UserDataCode `json:"-"`
	Inbox                chan string            `json:"-"`
}

type HasBehavior interface {
	ApplyBehavior(deltaTime float64)
	GetObject() *GameObject
}

type KeyStruct struct {
	Keys []string `json:"keys"`
}

type WorldData struct {
	Name    string       `json:"name"`
	Objects []GameObject `json:"objects"`
}
