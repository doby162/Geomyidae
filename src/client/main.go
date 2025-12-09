package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"log"
	"log/slog"

	"math"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"slices"
	"sync"

	assets "Geomyidae"

	"Geomyidae/internal/shared_structs"

	"github.com/gorilla/websocket"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	graphics "github.com/quasilyte/ebitengine-graphics"
	"github.com/quasilyte/gmath"
)

const (
	screenWidth  = 1280
	screenHeight = 832
)

var sprites map[string]*ebiten.Image

type Game struct{}

var worldMap map[string]*shared_structs.GameObject
var mu sync.Mutex
var oldKeys shared_structs.KeyStruct

type UserConfig struct {
	ConfigDir       string `json:"config_dir"`
	ConfigPath      string `json:"config_path"`
	WindowPositionX int    `json:"window_position_x"`
	WindowPositionY int    `json:"window_position_y"`
	WindowSizeX     int    `json:"window_size_x"`
	WindowSizeY     int    `json:"window_size_y"`
}

var userConfig UserConfig

func (g *Game) Update() error {
	// If worldMap is not yet initialized, skip update
	if len(worldMap) == 0 {
		return nil
	}

	winX, winY := ebiten.WindowPosition()
	winSizeX, winSizeY := ebiten.WindowSize()
	if userConfig.ConfigPath != "" && (winX != userConfig.WindowPositionX || winY != userConfig.WindowPositionY || winSizeX != userConfig.WindowSizeX || winSizeY != userConfig.WindowSizeY) {
		userConfig.WindowSizeX = winSizeX
		userConfig.WindowSizeY = winSizeY
		userConfig.WindowPositionX = winX
		userConfig.WindowPositionY = winY
		configData, _ := json.MarshalIndent(userConfig, "", "  ")
		configFile, err := os.Create(userConfig.ConfigPath)
		if err != nil {
			// Do not crash the game at this point if the config file cannot be updated
			slog.Error("Could not update user config file:", err)
		} else {
			configFile.Write(configData)
			configFile.Close()
		}
	}

	msg := shared_structs.KeyStruct{}

	for _, ekey := range inpututil.AppendPressedKeys([]ebiten.Key{}) {
		msg.Keys = append(msg.Keys, ekey.String())
	}
	if slices.Equal(msg.Keys, oldKeys.Keys) {
		return nil
	}
	oldKeys = msg
	msgBytes, _ := json.Marshal(msg)

	err := socket.WriteMessage(websocket.TextMessage, msgBytes)
	if err != nil {
		log.Fatal("Websocket send error:", err.Error())
	}

	return nil
}

var cameraX float64
var cameraY float64
var socket WSConn

var myPlayerUUID string

func (g *Game) Draw(screen *ebiten.Image) {
	var myPlayerOjbect shared_structs.GameObject
	mu.Lock()
	defer mu.Unlock()
	for _, object := range worldMap {
		if object.UUID == myPlayerUUID {
			myPlayerOjbect = *object
		}
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-float64(object.SpriteWidth/2), -float64(object.SpriteHeight/2))
		if object.SpriteFlipHorizontal {
			op.GeoM.Scale(-1, 1)
		}
		if object.SpriteFlipVertical {
			op.GeoM.Scale(1, -1)
		}
		if object.SpriteFlipDiagonal {
			op.GeoM.Rotate(math.Pi / 2)
		}
		op.GeoM.Rotate(object.Angle)
		op.GeoM.Translate(object.X-cameraX, object.Y-cameraY)
		screen.DrawImage(sprites[object.Sprite].SubImage(image.Rect(object.SpriteOffsetX, object.SpriteOffsetY, object.SpriteOffsetX+object.SpriteWidth, object.SpriteOffsetY+object.SpriteHeight)).(*ebiten.Image), op)
	}

	// Client side UI elements
	// Only used by Client side UI elements
	hudPosition := gmath.Vec{X: myPlayerOjbect.X - cameraX, Y: myPlayerOjbect.Y - cameraY}
	hudOverlay := graphics.NewSprite()
	hudOverlay.Pos.Base = &hudPosition
	hudOverlay.SetImage(sprites["friedEgg"])
	hudOverlay.SetScaleX(2)
	hudOverlay.SetScaleY(2)
	hudOverlay.Draw(screen)

	ebitenutil.DebugPrint(screen, "Camera position: "+fmt.Sprintf("%.2f, %.2f, goroutines:%v", cameraX, cameraY, runtime.NumGoroutine()))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// assets are embedded in package "assets"

func main() {
	// slog.SetLogLoggerLevel(slog.LevelDebug)
	worldMap = make(map[string]*shared_structs.GameObject)

	platformPackData, err := assets.FS.ReadFile("assets/img/platformerPack_industrial_tilesheet_64x64.png")
	if err != nil {
		log.Fatal(err)
	}
	platformPackImg, _, _ := image.Decode(bytes.NewReader(platformPackData))

	platformerIndustrialExpansionTilesetData, err := assets.FS.ReadFile("assets/img/kenny_pixel_platformer_industrial_expansion_tileset_64x64.png")
	if err != nil {
		log.Fatal(err)
	}
	platformerIndustrialExpansionTilesetImg, _, _ := image.Decode(bytes.NewReader(platformerIndustrialExpansionTilesetData))

	spaceShooterReduxData, err := assets.FS.ReadFile("assets/img/spaceShooterRedux_sheet.png")
	if err != nil {
		log.Fatal(err)
	}
	spaceShooterReduxImg, _, _ := image.Decode(bytes.NewReader(spaceShooterReduxData))

	friedEggData, err := assets.FS.ReadFile("assets/img/fried_egg.png")
	if err != nil {
		log.Fatal(err)
	}
	friedEggImg, _, _ := image.Decode(bytes.NewReader(friedEggData))

	// Create sprites map
	sprites = make(map[string]*ebiten.Image)
	sprites["platformerPack_industrial_tilesheet_64x64"] = ebiten.NewImageFromImage(platformPackImg)
	sprites["kenny_pixel_platformer_industrial_expansion_tileset_64x64"] = ebiten.NewImageFromImage(platformerIndustrialExpansionTilesetImg)
	sprites["spaceShooterRedux"] = ebiten.NewImageFromImage(spaceShooterReduxImg)
	sprites["friedEgg"] = ebiten.NewImageFromImage(friedEggImg)

	// Connect to WebSocket server
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws"}
	// For website in "production":
	// u := url.URL{Scheme: "wss", Host: "geomyidae-server.ekpyroticfrood.net", Path: "/ws"}
	slog.Debug("connecting to %s", u.String())

	conn, err := DialWS(u.String())
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer func(c WSConn) {
		err := c.Close()
		if err != nil {
			slog.Error(err.Error())
		}
	}(conn)

	socket = conn

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Fatal("Websocket read error:", err)
			}
			slog.Debug(string(bytes.TrimSpace(message)))
			var newState shared_structs.WorldData
			err = json.Unmarshal(message, &newState)
			if err != nil {
				slog.Error("unmarshal:", err)
			}
			mu.Lock()
			if myPlayerUUID == "" {
				myPlayerUUID = newState.Name
			}
			for _, object := range newState.Objects {
				key := object.UUID
				if key == newState.Name {
					cameraX, cameraY = object.X-screenWidth/2, object.Y-screenHeight/2
				}
				if object.Delete {
					delete(worldMap, key)
				} else {
					worldMap[key] = &object
				}
			}
			mu.Unlock()
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	go handleChannels(done, interrupt, conn)

	// Load user config data
	userConfig.ConfigDir, err = os.UserConfigDir()
	if err != nil {
		slog.Debug("Could not get user config dir: %v", err)
		userConfig.ConfigDir = ""
		// This is expected to fail on some systems, so we can just continue without a config dir.
	}
	if userConfig.ConfigDir != "" {
		userConfig.ConfigDir = userConfig.ConfigDir + string(os.PathSeparator) + "Geomyidae"
		err = os.MkdirAll(userConfig.ConfigDir, os.ModePerm)
		if err != nil {
			log.Fatal("Could not create user config dir:", err)
			userConfig.ConfigDir = ""
		}
		// Load config file or create it if it doesn't exist
		userConfig.ConfigPath = userConfig.ConfigDir + string(os.PathSeparator) + "config.json"
		configFileData, err := os.ReadFile(userConfig.ConfigPath)
		if err != nil {
			// Create default config file
			configFile, err := os.Create(userConfig.ConfigPath)
			if err != nil {
				log.Fatal("Could not create user config file:", err)
			} else {
				defaultConfigData, _ := json.MarshalIndent(userConfig, "", "  ")
				_, err = configFile.Write(defaultConfigData)
				if err != nil {
					log.Fatal("Could not write default user config file:", err)
				}
				configFile.Close()
			}
		} else {
			// Load existing config file
			err = json.Unmarshal(configFileData, &userConfig)
			if err != nil {
				log.Fatal("Could not parse user config file:", err)
			}
		}
		slog.Debug("User config file path:", userConfig.ConfigPath)
	}

	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Geomyidae")
	if userConfig.WindowPositionX != 0 || userConfig.WindowPositionY != 0 {
		ebiten.SetWindowPosition(userConfig.WindowPositionX, userConfig.WindowPositionY)
	}
	if userConfig.WindowSizeX != 0 && userConfig.WindowSizeY != 0 {
		ebiten.SetWindowSize(userConfig.WindowSizeX, userConfig.WindowSizeY)
	}
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
