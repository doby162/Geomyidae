package main

import (
	"github.com/doby162/go-higher-order"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// key checks take a function to run if the key is held
func checkKey(checkKey ebiten.Key, fn func()) {
	if higher_order.AnySlice(heldKeys, func(key ebiten.Key) bool {
		return key == checkKey
	}) {
		fn()
	}
}
func handleKeyState() {
	// keys are added to held when pressed.
	// keys are removed when released
	heldKeys = inpututil.AppendPressedKeys(heldKeys)
	releasedKeys = inpututil.AppendJustReleasedKeys([]ebiten.Key{})
	for _, key := range releasedKeys {
		heldKeys = higher_order.FilterSlice(heldKeys, func(key2 ebiten.Key) bool {
			return key != key2
		})
	}
}
