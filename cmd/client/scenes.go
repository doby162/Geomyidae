package main

// each character in this "string" will be interpreted as rune, or a numeric representation of a character, indexed
// by the column and row, read top to bottom, left to right. Each unique character could correspond to a tile type, right now
// I'm using 0 for empty and 1 for a rigid tile.
// Layers could be achieved through duplicating this structure for each layer.
// Yes, this could have been json, slices, a struct, etc, and I made it a string. Sue me.
const scene01 = "11111111111111111111\n" +
	"10000000000000000001\n" +
	"10000000000000000001\n" +
	"10000000000000000001\n" +
	"10000000000000111001\n" +
	"10000000000000000001\n" +
	"10001110000000000001\n" +
	"10000000000000000001\n" +
	"10000000000011100001\n" +
	"10000000000000000001\n" +
	"10000111000000000001\n" +
	"10000000000000000001\n" +
	"11111111111111111111\n"
