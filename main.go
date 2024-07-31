package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
)

type CreateRoom struct {
	Room string `json:"room"`
}

type UpdateParticipant struct {
	Room              string                `json:"room"`
	Identity          string                `json:"identity"`
	CanSubscribe      bool                  `json:"can_subscribe"`
	CanPublish        bool                  `json:"can_publish"`
	CanPublishData    bool                  `json:"can_publish_data"`
	CanPublishSources []livekit.TrackSource `json:"can_publish_sources"`
}

type ConnectToRoom struct {
	Room     string `json:"room"`
	Identity string `json:"identity"`
	Token    string `json:"token"`
}

type HandleParticipant struct {
	Room     string `json:"room"`
	Identity string `json:"identity"`
}

func main() {
	apiKey := os.Getenv("LIVEKIT_API_KEY")
	apiSecret := os.Getenv("LIVEKIT_API_SECRET")
	host := os.Getenv("LIVEKIT_API_HOST")

	roomService := lksdk.NewRoomServiceClient(host, apiKey, apiSecret)

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.POST("/create-room", func(c *gin.Context) {
		var req *CreateRoom
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if req.Room == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "roomName is required"})
			return
		}

		room, err := roomService.CreateRoom(c.Request.Context(), &livekit.CreateRoomRequest{Name: req.Room})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"room": room})
	})

	router.GET("/join-token", func(c *gin.Context) {
		roomName := c.Query("roomName")
		identity := c.Query("identity")
		if roomName == "" || identity == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "roomName and identity are required"})
			return
		}

		at := auth.NewAccessToken(apiKey, apiSecret).
			AddGrant(&auth.VideoGrant{RoomJoin: true, Room: roomName}).
			SetIdentity(identity)

		token, err := at.ToJWT()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": token})
	})

	router.POST("/join-room", func(c *gin.Context) {
		var req *ConnectToRoom
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		room, err := lksdk.ConnectToRoomWithToken(host, req.Token, &lksdk.RoomCallback{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"room": room})
	})

	router.GET("/list-rooms", func(c *gin.Context) {
		rooms, err := roomService.ListRooms(c.Request.Context(), &livekit.ListRoomsRequest{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, rooms.Rooms)
	})

	router.GET("/room", func(c *gin.Context) {
		roomName := c.Query("roomName")
		if roomName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "roomName is required"})
			return
		}

		room, err := roomService.ListRooms(c.Request.Context(), &livekit.ListRoomsRequest{
			Names: []string{roomName},
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if room == nil {
			c.JSON(http.StatusOK, gin.H{
				"room": nil,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"room": room.Rooms[0],
		})
	})

	router.GET("/list-participants", func(c *gin.Context) {
		roomName := c.Query("roomName")
		if roomName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "roomName is required"})
			return
		}

		participants, err := roomService.ListParticipants(c.Request.Context(), &livekit.ListParticipantsRequest{
			Room: roomName,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, participants.Participants)
	})

	router.GET("/detail-participant", func(c *gin.Context) {
		roomName := c.Query("roomName")
		identity := c.Query("identity")
		if roomName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "roomName is required"})
			return
		}

		if identity == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "identity is required"})
			return
		}

		participant, err := roomService.GetParticipant(c.Request.Context(), &livekit.RoomParticipantIdentity{
			Room:     roomName,
			Identity: identity,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, participant)
	})

	router.POST("/remove-participant", func(c *gin.Context) {
		var req *HandleParticipant
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if req.Room == "" || req.Identity == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "roomName and identity are required"})
			return
		}

		_, err := roomService.RemoveParticipant(c.Request.Context(), &livekit.RoomParticipantIdentity{Room: req.Room, Identity: req.Identity})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Participant removed successfully"})
	})

	router.POST("/mute-participant", func(c *gin.Context) {
		var req *HandleParticipant
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if req.Room == "" || req.Identity == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "roomName and identity are required"})
			return
		}

		_, err := roomService.MutePublishedTrack(c.Request.Context(), &livekit.MuteRoomTrackRequest{Room: req.Room, Identity: req.Identity, Muted: true})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Participant muted successfully"})
	})

	router.POST("/unmute-participant", func(c *gin.Context) {
		var req *HandleParticipant
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if req.Room == "" || req.Identity == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "roomName and identity are required"})
			return
		}

		_, err := roomService.MutePublishedTrack(c.Request.Context(), &livekit.MuteRoomTrackRequest{Room: req.Room, Identity: req.Identity, Muted: false})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Participant unmuted successfully"})
	})

	router.PUT("/update-participant", func(c *gin.Context) {
		var req *UpdateParticipant
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		res, err := roomService.UpdateParticipant(c.Request.Context(), &livekit.UpdateParticipantRequest{
			Room:     req.Room,
			Identity: req.Identity,
			Permission: &livekit.ParticipantPermission{
				CanSubscribe:      req.CanSubscribe,
				CanPublish:        req.CanPublish,
				CanPublishData:    req.CanPublishData,
				CanPublishSources: req.CanPublishSources,
			},
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"result": res,
		})
	})

	log.Println("Starting server on :8000")
	router.Run(":8000")
}
