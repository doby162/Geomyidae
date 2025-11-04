package tiled

// This will import a Tiled native .tmx file and extract tile data from it.
// (At least, the bytes from said file.)
// There is no need to "export" from Tiled to JSON or other formats.
// Use base64 encoded and uncompressed data in Tiled, with infinite map enabled.

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"log"
	"strings"
)

// https://blog.kowalczyk.info/tools/xmltogo/
type Map struct {
	XMLName        xml.Name `xml:"map"`
	Text           string   `xml:",chardata"`
	Version        string   `xml:"version,attr"`
	Tiledversion   string   `xml:"tiledversion,attr"`
	Orientation    string   `xml:"orientation,attr"`
	Renderorder    string   `xml:"renderorder,attr"`
	Width          int      `xml:"width,attr"`
	Height         int      `xml:"height,attr"`
	Tilewidth      int      `xml:"tilewidth,attr"`
	Tileheight     int      `xml:"tileheight,attr"`
	Infinite       int      `xml:"infinite,attr"`
	Nextlayerid    int      `xml:"nextlayerid,attr"`
	Nextobjectid   int      `xml:"nextobjectid,attr"`
	Editorsettings struct {
		Text   string `xml:",chardata"`
		Export struct {
			Text   string `xml:",chardata"`
			Target string `xml:"target,attr"`
			Format string `xml:"format,attr"`
		} `xml:"export"`
	} `xml:"editorsettings"`
	Tileset struct {
		Text       string `xml:",chardata"`
		Firstgid   string `xml:"firstgid,attr"`
		Name       string `xml:"name,attr"`
		Tilewidth  int    `xml:"tilewidth,attr"`
		Tileheight int    `xml:"tileheight,attr"`
		Tilecount  int    `xml:"tilecount,attr"`
		Columns    int    `xml:"columns,attr"`
		Image      struct {
			Text   string `xml:",chardata"`
			Source string `xml:"source,attr"`
			Width  string `xml:"width,attr"`
			Height string `xml:"height,attr"`
		} `xml:"image"`
	} `xml:"tileset"`
	Layer struct {
		Text   string `xml:",chardata"`
		ID     string `xml:"id,attr"`
		Name   string `xml:"name,attr"`
		Width  int    `xml:"width,attr"`
		Height int    `xml:"height,attr"`
		Data   struct {
			Text        string `xml:",chardata"`
			Encoding    string `xml:"encoding,attr"`
			Compression string `xml:"compression,attr"`
			Chunk       []struct {
				Text   string `xml:",chardata"`
				X      int    `xml:"x,attr"`
				Y      int    `xml:"y,attr"`
				Width  int    `xml:"width,attr"`
				Height int    `xml:"height,attr"`
			} `xml:"chunk"`
		} `xml:"data"`
	} `xml:"layer"`
}

type tileDatum struct {
	ID               uint32
	FlipHorizontally bool
	FlipVertically   bool
	Row              int
	Col              int
}

var tileData []tileDatum

func GetTileData(tileFileInput []byte) []tileDatum {
	var m Map
	err := xml.Unmarshal(tileFileInput, &m)
	if err != nil {
		log.Fatal(err)
	}

	// Tiled maps must be infinite
	if m.Infinite != 1 {
		log.Fatal("Tiled maps must be infinite")
	}

	// Tiled maps must be base64 encoded and uncompressed
	if m.Layer.Data.Encoding != "base64" || m.Layer.Data.Compression != "" {
		log.Fatal("Tiled maps must be base64 encoded and uncompressed")
	}

	fmt.Println("Tilemap Width:", m.Layer.Width, "Height:", m.Layer.Height)

	// Tiled stores data in chunks for infinite maps
	for _, chunk := range m.Layer.Data.Chunk {
		fmt.Println("Chunk X:", chunk.X, "Y:", chunk.Y)

		data, err := base64.StdEncoding.DecodeString(strings.TrimSpace(chunk.Text))
		if err != nil {
			log.Fatal("error:", err)
		}

		// https://doc.mapeditor.org/en/stable/reference/tmx-map-format/#data
		// The data is an array of bytes, which should be interpreted as an array of unsigned 32-bit integers using little-endian byte ordering.
		// https://doc.mapeditor.org/en/stable/reference/global-tile-ids/#
		thisRow := 0
		thisCol := -1
		for i := 0; i < len(data); i += 4 {
			// https://doc.mapeditor.org/en/stable/reference/global-tile-ids/#tile-flipping
			// The highest four bits of the 32-bit GID are flip flags, and you will need to read and clear them before you can access the GID itself to identify the tile.
			// Bit 32 is used for storing whether the tile is horizontally flipped, bit 31 is used for the vertically flipped tiles. In orthogonal and isometric maps, bit 30 indicates whether the tile is flipped (anti) diagonally, which enables tile rotation, and bit 29 can be ignored. In hexagonal maps, bit 30 indicates whether the tile is rotated 60 degrees clockwise, and bit 29 indicates 120 degrees clockwise rotation.

			// Flip flag is the highest bit
			flipHorizontally := (data[i+3] & 0x80) != 0
			flipVertically := (data[i+3] & 0x40) != 0

			tileID := binary.LittleEndian.Uint32(data[i : i+4])

			// Clear highest four bits of the most significant byte (safe for a single byte)
			tileID &^= 0xF0000000

			thisCol++
			if thisCol >= chunk.Width {
				thisCol = 0
				thisRow++
				fmt.Println()
			}
			fmt.Print(tileID)

			// store tile data
			tileData = append(tileData, tileDatum{
				ID:               tileID,
				FlipHorizontally: flipHorizontally,
				FlipVertically:   flipVertically,
				Row:              thisRow + chunk.Y,
				Col:              thisCol + chunk.X,
			})
		}
		fmt.Println()
	}

	// https://doc.mapeditor.org/en/stable/reference/global-tile-ids/
	return tileData
}
