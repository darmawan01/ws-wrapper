package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/darmawan01/ws-wrapper"
	"github.com/gin-gonic/gin"
)

type health struct {
	router *gin.RouterGroup
}

func (h *health) register() {
	ws.RegisterChannelHandler("public/health", h.healthCheck)
}

func (h *health) healthCheck(msg ws.RequestMessage, c *ws.Client) {
	c.Send(ws.ResponseMessage{
		Result: "online",
	})
}

func main() {
	engine := gin.Default()

	router := engine.Group("/api")

	/* Register websocket route */
	router.GET("/ws", func(ctx *gin.Context) {
		/* Websocket wrapper handler */
		ws.WebsocketHandler(ctx.Writer, ctx.Request)
	})

	h := &health{
		router: router,
	}

	h.register()

	server := http.Server{
		Addr:    ":9898",
		Handler: engine,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server shutdown failed:", err)
	}

	log.Println("Server gracefully stopped")

}
