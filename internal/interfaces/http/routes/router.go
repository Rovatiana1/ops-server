package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "ops-server/docs"

	"ops-server/configs"
	auditCtrl "ops-server/internal/domain/audit/controller"
	configCtrl "ops-server/internal/domain/config_general/controller"
	metCtrl "ops-server/internal/domain/metrics/controller"
	notifCtrl "ops-server/internal/domain/notification/controller"
	rbacCtrl "ops-server/internal/domain/rbac/controller"
	rbacMod "ops-server/internal/domain/rbac/models"
	userCtrl "ops-server/internal/domain/user/controller"
	"ops-server/internal/interfaces/http/middleware"

	proxy "ops-server/internal/interfaces/http/proxy"

	"ops-server/internal/interfaces/http/handlers"
)

// Setup enregistre toutes les routes applicatives sur le moteur Gin.
func Setup(
	engine *gin.Engine,
	userC *userCtrl.UserController,
	rbacC *rbacCtrl.RBACController,
	notifC *notifCtrl.NotificationController,
	metricsC *metCtrl.MetricsController,
	auditC *auditCtrl.AuditController,
	configGeneralC *configCtrl.ConfigGeneralController,
	jwtCfg configs.JWTConfig,
	proxyCfg configs.ProxyConfig,
) {
	engine.Use(middleware.RequestLogger())
	engine.Use(gin.Recovery())

	// ── Swagger ───────────────────────────────────────────────────────────────
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerFiles.Handler,
		ginSwagger.URL("/swagger/doc.json"),
		ginSwagger.DefaultModelsExpandDepth(2),
		ginSwagger.PersistAuthorization(true),
	))

	// ── Gateway → Sampling Service ─────────────────────────────────────────────
	samplingProxy := proxy.NewReverseProxy(proxyCfg.Sampling, "/api/v1")

	engine.Any("/api/v1/plans/*path", samplingProxy)
	engine.Any("/api/v1/runs/*path", samplingProxy)
	engine.Any("/api/v1/ingestion/*path", samplingProxy)
	engine.Any("/api/v1/pull-configs/*path", samplingProxy)
	engine.Any("/api/v1/release/*path", samplingProxy)
	engine.Any("/api/v1/log/*path", samplingProxy)

	// ── Santé / Observabilité ─────────────────────────────────────────────────
	engine.GET("/api/v1/metrics/prometheus", gin.WrapH(promhttp.Handler()))
	engine.GET("/api/v1/metrics/health", handlers.MetricsHealth)

	v1 := engine.Group("/api/v1")

	// ── Auth (public) ──────────────────────────────────────────────────────────
	auth := v1.Group("/auth")
	{
		auth.POST("/register", userC.Register)
		auth.POST("/signin", userC.SignIn)
		auth.POST("/signin-ldap", userC.SignInLDAP)
		auth.POST("/refresh", userC.RefreshToken)
		auth.POST("/logout", middleware.Auth(jwtCfg), userC.Logout)
	}

	// ── Users (authentifié) ────────────────────────────────────────────────────
	users := v1.Group("/users", middleware.Auth(jwtCfg))
	{
		users.GET("/me", userC.GetMe)

		admin := users.Group("", middleware.RequireRole(rbacMod.RoleNameAdmin))
		{
			admin.GET("", userC.ListUsers)
			admin.GET("/:id", userC.GetUser)
			admin.PATCH("/:id", userC.UpdateUser)
			admin.DELETE("/:id", userC.DeleteUser)
		}
	}

	// ── RBAC (admin uniquement) ────────────────────────────────────────────────
	rbac := v1.Group("/rbac",
		middleware.Auth(jwtCfg),
		middleware.RequireRole(rbacMod.RoleNameAdmin),
	)
	{
		// Roles CRUD
		rbac.POST("/roles", rbacC.CreateRole)
		rbac.GET("/roles", rbacC.ListRoles)
		rbac.GET("/roles/:id", rbacC.GetRole)
		rbac.PATCH("/roles/:id", rbacC.UpdateRole)
		rbac.DELETE("/roles/:id", rbacC.DeleteRole)

		// Permissions d'un rôle
		rbac.GET("/roles/:id/permissions", rbacC.GetRolePermissions)
		rbac.PUT("/roles/:id/permissions", rbacC.SetRolePermissions)              // remplace tout
		rbac.POST("/roles/:id/permissions", rbacC.AddRolePermission)              // ajoute une
		rbac.DELETE("/roles/:id/permissions/:permId", rbacC.RemoveRolePermission) // retire une

		// Permissions CRUD
		rbac.POST("/permissions", rbacC.CreatePermission)
		rbac.GET("/permissions", rbacC.ListPermissions)
		rbac.GET("/permissions/:id", rbacC.GetPermission)
		rbac.PATCH("/permissions/:id", rbacC.UpdatePermission)
		rbac.DELETE("/permissions/:id", rbacC.DeletePermission)

		// Assignation rôles <-> utilisateurs
		rbac.GET("/users/:userId/roles", rbacC.GetUserRoles)
		rbac.POST("/users/:userId/roles", rbacC.AssignRoleToUser)
		rbac.DELETE("/users/:userId/roles/:roleId", rbacC.RemoveRoleFromUser)
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
		middleware.RequireRole(rbacMod.RoleNameAdmin, rbacMod.RoleNameManager),
	)
	{
		met.GET("", metricsC.ListMetrics)
		met.GET("/events", metricsC.ListEvents)
		met.GET("/events/:id", metricsC.GetEvent)
	}

	// ── Audit (admin) ──────────────────────────────────────────────────────────
	audit := v1.Group("/audit",
		middleware.Auth(jwtCfg),
		middleware.RequireRole(rbacMod.RoleNameAdmin),
	)
	{
		audit.GET("/trails", auditC.ListTrails)
		audit.GET("/trails/:id", auditC.GetTrail)
		audit.GET("/logs", auditC.ListLogs)
	}

	// ── Configuration General (admin) ──────────────────────────────────────────────────────────
	configGeneral := v1.Group("/config-general",
		middleware.Auth(jwtCfg),
		middleware.RequireRole(rbacMod.RoleNameAdmin),
	)
	{
		configGeneral.POST("", configGeneralC.Create)
		configGeneral.GET("", configGeneralC.List)
		configGeneral.GET("/by-key", configGeneralC.GetByKey)
		configGeneral.GET("/:id", configGeneralC.GetByID)
		configGeneral.PATCH("/:id", configGeneralC.Update)
		configGeneral.DELETE("/:id", configGeneralC.Delete)
		// configGeneral.DELETE("/ldap/test", configGeneralC.TestLdapConnection)
	}
}
