package api

import (
	"github.com/wb-go/wbf/ginext"

	"EventBooker/internal/api/handlers"
)

func SetupRoutes(h *handlers.Handlers, g *ginext.Engine) {

	g.Use(ginext.Logger(), ginext.Recovery(), handlers.CORSMiddleware())
	g.LoadHTMLGlob("web/*.html")

	g.GET("/", handlers.GetHome)
	g.GET("/admin_panel", handlers.GetAdminLogin)
	g.GET("/admin", handlers.GetAdmin)

	auth := g.Group("/auth")
	{
		auth.POST("/register", h.User.Register)
		auth.POST("/login", h.User.Login)
	}

	g.GET("/events", h.Event.GetListEvents)
	g.GET("/events/:id", h.Event.GetEvent)

	api := g.Group("/api")
	api.Use(handlers.AuthMiddleware())
	{
		api.POST("/events/:event_id/book", h.Booking.Book)
		api.POST("/events/:event_id/confirm", h.Booking.Confirm)
		api.POST("/events/:event_id/cancel", h.Booking.Cancel)
		api.GET("/books", h.Booking.GetListBooking)

		admin := api.Group("/admin")
		admin.Use(handlers.AdminMiddleware())
		{
			admin.GET("/check")
			admin.POST("/events", h.Event.CreateEvent)
			admin.GET("/users", h.User.GetList)
		}
	}

}
