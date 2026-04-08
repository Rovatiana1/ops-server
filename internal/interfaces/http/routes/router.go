package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "ops-server/docs" // Swagger docs (générés par swag init)

	"ops-server/configs"
	auditCtrl "ops-server/internal/domain/audit/controller"
	metricsCtrl "ops-server/internal/domain/metrics/controller"
	notifCtrl "ops-server/internal/domain/notification/controller"
	userCtrl "ops-server/internal/domain/user/controller"
	"ops-server/internal/domain/user/models"
	"ops-server/internal/interfaces/http/middleware"
)

// Setup enregistre toutes les routes applicatives sur le moteur Gin.
func Setup(
	engine *gin.Engine,
	userC *userCtrl.UserController,
	notifC *notifCtrl.NotificationController,
	metricsC *metricsCtrl.MetricsController,
	auditC *auditCtrl.AuditController,
	jwtCfg configs.JWTConfig,
) {
	// ── Middlewares globaux ────────────────────────────────────────────────────
	engine.Use(middleware.RequestLogger())
	engine.Use(gin.Recovery())

	// ── Santé / Observabilité ─────────────────────────────────────────────────
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "version": "1.0.0"})
	})
	engine.GET("/metrics/prometheus", gin.WrapH(promhttp.Handler()))

	// ── Swagger UI ────────────────────────────────────────────────────────────
	// URL: http://localhost:8080/swagger/index.html
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.URL("/swagger/doc.json"),
		ginSwagger.DefaultModelsExpandDepth(2),
	))

	// ── API v1 ─────────────────────────────────────────────────────────────────
	v1 := engine.Group("/api/v1")

	// ── Auth (public) ──────────────────────────────────────────────────────────
	auth := v1.Group("/auth")
	{
		auth.POST("/register", userC.Register)
		auth.POST("/signin", userC.SignIn)
		auth.POST("/refresh", userC.RefreshToken)
		auth.POST("/logout", middleware.Auth(jwtCfg), userC.Logout)
	}

	// ── Users (authentifié) ────────────────────────────────────────────────────
	users := v1.Group("/users", middleware.Auth(jwtCfg))
	{
		users.GET("/me", userC.GetMe)

		// Admin seulement
		adminOnly := users.Group("",
			middleware.RequireRole(models.RoleNameAdmin),
		)
		{
			adminOnly.GET("", userC.ListUsers)
			adminOnly.GET("/:id", userC.GetUser)
			adminOnly.PATCH("/:id", userC.UpdateUser)
			adminOnly.DELETE("/:id", userC.DeleteUser)
			adminOnly.POST("/:id/roles", userC.AssignRole)
		}
	}

	// ── Notifications (authentifié) ────────────────────────────────────────────
	notifs := v1.Group("/notifications", middleware.Auth(jwtCfg))
	{
		notifs.GET("", notifC.ListMyNotifications)
		notifs.GET("/unread-count", notifC.CountUnread)
		notifs.PATCH("/:id/read", notifC.MarkAsRead)
		notifs.DELETE("/:id", notifC.DeleteNotification)
	}

	// ── Metrics (admin + manager) ──────────────────────────────────────────────
	met := v1.Group("/metrics",
		middleware.Auth(jwtCfg),
		middleware.RequireRole(models.RoleNameAdmin, models.RoleNameManager),
	)
	{
		met.GET("", metricsC.ListMetrics)
		met.GET("/events", metricsC.ListEvents)
		met.GET("/events/:id", metricsC.GetEvent)
	}

	// ── Audit (admin seulement) ────────────────────────────────────────────────
	audit := v1.Group("/audit",
		middleware.Auth(jwtCfg),
		middleware.RequireRole(models.RoleNameAdmin),
	)
	{
		audit.GET("/trails", auditC.ListTrails)
		audit.GET("/trails/:id", auditC.GetTrail)
		audit.GET("/logs", auditC.ListLogs)
	}
}
