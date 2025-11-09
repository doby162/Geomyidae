package tile

import "Geomyidae/internal/shared_structs"

type Tile struct {
	*shared_structs.GameObject
}

func NewTile(gameObject *shared_structs.GameObject) *Tile {
	return &Tile{gameObject}
}

func (t *Tile) ApplyBehavior(deltaTime float64) {
	return
}

func (t *Tile) GetObject() *shared_structs.GameObject {
	return t.GameObject
}
