package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"log"
	"log/slog"
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
)

const (
	screenWidth  = 1280
	screenHeight = 832
	tileSize     = 64
	tileHalf     = 32
)

var sprites map[string]*ebiten.Image

type Game struct{}

var us string
var name string
var worldMap map[string]*shared_structs.GameObject
var mu sync.Mutex
var oldKeys shared_structs.KeyStruct

type UserConfig struct {
	ConfigDir       string `json:"config_dir"`
	ConfigPath      string `json:"config_path"`
	WindowPositionX int    `json:"window_position_x"`
	WindowPositionY int    `json:"window_position_y"`
}

var userConfig UserConfig

func (g *Game) Update() error {
	winX, winY := ebiten.WindowPosition()
	if winX != userConfig.WindowPositionX || winY != userConfig.WindowPositionY {
		userConfig.WindowPositionX = winX
		userConfig.WindowPositionY = winY
		configData, _ := json.MarshalIndent(userConfig, "", "  ")
		configFile, err := os.Create(userConfig.ConfigPath)
		if err != nil {
			// Do not crash the game at this point if the config file cannot be updated
			slog.Error("Could not update user config file:", err)
		}
		configFile.Write(configData)
		configFile.Close()
	}

	if us == "" {
		for _, obj := range worldMap {
			if obj.Name == name {
				us = obj.Name
			}
		}
	}
	cameraX = (worldMap[us].X * tileSize) - screenWidth/2
	cameraY = (worldMap[us].Y * tileSize) - screenHeight/2 - (2 * tileSize)

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

func (g *Game) Draw(screen *ebiten.Image) {
	mu.Lock()
	defer mu.Unlock()
	for _, object := range worldMap {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-tileHalf, -tileHalf)
		op.GeoM.Rotate(object.Angle)
		op.GeoM.Translate(-cameraX, -cameraY)
		op.GeoM.Translate(object.X*tileSize-tileHalf, object.Y*tileSize-tileHalf)
		screen.DrawImage(sprites[object.Sprite].SubImage(image.Rect(object.OffsetX, object.OffsetY, object.OffsetX+object.Width, object.OffsetY+object.Height)).(*ebiten.Image), op)
	}

	ebitenutil.DebugPrint(screen, "Camera position: "+fmt.Sprintf("%.2f, %.2f, goroutines:%v", cameraX, cameraY, runtime.NumGoroutine()))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// assets are embedded in package "assets"

func main() {
	//slog.SetLogLoggerLevel(slog.LevelDebug)
	worldMap = make(map[string]*shared_structs.GameObject)
	beet, err := assets.FS.ReadFile("assets/img/placeholderSprite.png")
	if err != nil {
		log.Fatal(err)
	}
	bert, _, _ := image.Decode(bytes.NewReader(beet))

	platformPackData, err := assets.FS.ReadFile("assets/img/platformerPack_industrial_tilesheet_64x64.png")
	if err != nil {
		log.Fatal(err)
	}
	platformPackImg, _, _ := image.Decode(bytes.NewReader(platformPackData))

	spaceShooterReduxData, err := assets.FS.ReadFile("assets/img/spaceShooterRedux_sheet.png")
	if err != nil {
		log.Fatal(err)
	}
	spaceShooterReduxImg, _, _ := image.Decode(bytes.NewReader(spaceShooterReduxData))

	// Create sprites map
	sprites = make(map[string]*ebiten.Image)
	sprites["player_01"] = ebiten.NewImageFromImage(bert)
	sprites["tom"] = ebiten.NewImageFromImage(bert)
	sprites["platformerPack_industrial"] = ebiten.NewImageFromImage(platformPackImg)
	sprites["spaceShooterRedux"] = ebiten.NewImageFromImage(spaceShooterReduxImg)

	// Connect to WebSocket server
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/ws"}
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
			slog.Debug("recv: %s", message)
			var newState shared_structs.WorldData
			err = json.Unmarshal(message, &newState)
			if err != nil {
				slog.Error("unmarshal:", err)
			}
			name = newState.Name
			mu.Lock()
			for _, object := range newState.Objects {
				key := object.UUID
				worldMap[key] = &object
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
		log.Fatal("Could not get user config dir:", err)
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
		log.Println("User config file path:", userConfig.ConfigPath)
	}

	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Geomyidae")
	if userConfig.WindowPositionX != 0 || userConfig.WindowPositionY != 0 {
		ebiten.SetWindowPosition(userConfig.WindowPositionX, userConfig.WindowPositionY)
	}
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
