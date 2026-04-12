package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ops-server/configs"

	// Infrastructure

	ldapInfra "ops-server/internal/infrastructure/ldap"
	postgresInfra "ops-server/internal/infrastructure/postgres"
	redisInfra "ops-server/internal/infrastructure/redis"

	// Domaine user
	userCtrl "ops-server/internal/domain/user/controller"
	userRepo "ops-server/internal/domain/user/repository"
	userSvc "ops-server/internal/domain/user/service"

	// Domaine rbac
	rbacCtrl "ops-server/internal/domain/rbac/controller"
	rbacRepo "ops-server/internal/domain/rbac/repository"
	rbacSvc "ops-server/internal/domain/rbac/service"

	// Domaine notification
	notifCtrl "ops-server/internal/domain/notification/controller"
	notifRepo "ops-server/internal/domain/notification/repository"
	notifSvc "ops-server/internal/domain/notification/service"

	// Domaine metrics
	metricsCtrl "ops-server/internal/domain/metrics/controller"
	metricsRepo "ops-server/internal/domain/metrics/repository"
	metricsSvc "ops-server/internal/domain/metrics/service"

	// Domaine audit
	auditCtrl "ops-server/internal/domain/audit/controller"
	auditRepo "ops-server/internal/domain/audit/repository"
	auditSvc "ops-server/internal/domain/audit/service"

	// Domaine config general
	configCtrl "ops-server/internal/domain/config_general/controller"
	configRepo "ops-server/internal/domain/config_general/repository"
	configSvc "ops-server/internal/domain/config_general/service"

	// Interfaces
	"ops-server/internal/interfaces/http/routes"

	"ops-server/pkg/logger"
)

const configPath = "configs/config.yaml"

func main() {
	// ── 1. Configuration ──────────────────────────────────────────────────────
	cfg, err := configs.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// ── 2. Logger ─────────────────────────────────────────────────────────────
	logger.Init(cfg.Observability.LogLevel, cfg.App.Debug)
	defer logger.Sync()
	log := logger.L()
	log.Info("starting", zap.String("app", cfg.App.Name), zap.String("env", cfg.App.Env))

	// ── 3. PostgreSQL ──────────────────────────────────────────────────────────
	db, err := postgresInfra.New(cfg.Database, cfg.App.Debug)
	if err != nil {
		log.Fatal("postgres failed", zap.Error(err))
	}
	defer postgresInfra.Close(db) //nolint:errcheck

	// ── 4. Redis ───────────────────────────────────────────────────────────────
	redisClient, err := redisInfra.NewClient(cfg.Redis)
	if err != nil {
		log.Fatal("redis failed", zap.Error(err))
	}
	cache := redisInfra.NewCache(redisClient)
	_ = redisInfra.NewLock(redisClient)
	_ = redisInfra.NewRateLimiter(redisClient)

	// ── 5. Kafka ───────────────────────────────────────────────────────────────
	// if err := kafkaCore.DialBroker(cfg.Kafka); err != nil {
	// 	log.Fatal("kafka failed", zap.Error(err))
	// }
	// signupWriter := kafkaCore.NewWriter(cfg.Kafka, cfg.Kafka.Topics.Signup)
	// signinWriter := kafkaCore.NewWriter(cfg.Kafka, cfg.Kafka.Topics.Signin)
	// dlqWriter := kafkaCore.NewWriter(cfg.Kafka, cfg.Kafka.Topics.DLQ)
	// dlqProd := kafkaCore.NewProducer(dlqWriter)

	// _ = kafkaProducer.NewSignupProducer(kafkaCore.NewProducer(signupWriter))
	// _ = kafkaProducer.NewSigninProducer(kafkaCore.NewProducer(signinWriter))
	// _ = kafkaProducer.NewDLQProducer(dlqProd)

	// signupHdlr := kafkaHandler.NewSignupHandler(cache)
	// signinHdlr := kafkaHandler.NewSigninHandler(cache)
	// signupCons := kafkaConsumer.NewSignupConsumer(cfg.Kafka, signupHdlr, dlqProd)
	// signinCons := kafkaConsumer.NewSigninConsumer(cfg.Kafka, signinHdlr, dlqProd)
	// retryCons := kafkaConsumer.NewRetryConsumer(cfg.Kafka, signupHdlr, dlqProd)

	// ── 6. DI — domaines & LDAP ──────────────────────────────────────────────────────

	// config
	cRepo := configRepo.NewConfigGeneralRepository(db)
	cSvc := configSvc.NewConfigGeneralService(cRepo)
	cCtrl := configCtrl.NewConfigGeneralController(cSvc)

	// ── LDAP INFRA 🔥 ─────────────────────────────────────────────
	ldapProvider := ldapInfra.NewConfigProvider(cSvc, cache)
	ldapService := ldapInfra.NewService(ldapProvider)

	// User (auth uniquement)
	uRepo := userRepo.NewUserRepository(db)
	uSvc := userSvc.NewUserService(uRepo, cache, cfg.JWT, ldapService)
	uCtrl := userCtrl.NewUserController(uSvc)
	// RBAC (roles + permissions + assignations)
	rRepo := rbacRepo.NewRBACRepository(db)
	rSvc := rbacSvc.NewRBACService(rRepo)
	rCtrl := rbacCtrl.NewRBACController(rSvc)

	// Notification
	nRepo := notifRepo.NewNotificationRepository(db)
	nSvc := notifSvc.NewNotificationService(nRepo)
	nCtrl := notifCtrl.NewNotificationController(nSvc)

	// Metrics
	mRepo := metricsRepo.NewMetricsRepository(db)
	mSvc := metricsSvc.NewMetricsService(mRepo)
	mCtrl := metricsCtrl.NewMetricsController(mSvc)

	// Audit
	aRepo := auditRepo.NewAuditRepository(db)
	aSvc := auditSvc.NewAuditService(aRepo)
	aCtrl := auditCtrl.NewAuditController(aSvc)

	// ── 7. HTTP ────────────────────────────────────────────────────────────────
	if !cfg.App.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // ou ton frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	routes.Setup(engine, uCtrl, rCtrl, nCtrl, mCtrl, aCtrl, cCtrl, cfg.JWT, cfg.Proxy)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.Port),
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// ── 8. Workers ─────────────────────────────────────────────────────────────
	// workerPool := workers.NewWorkerPool(signupCons, signinCons, retryCons)
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()
	// go workerPool.Run(ctx)

	// ── 9. Start ───────────────────────────────────────────────────────────────
	go func() {
		log.Info("http listening",
			zap.String("addr", srv.Addr),
			zap.String("swagger", fmt.Sprintf("http://localhost%s/swagger/index.html", srv.Addr)),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server error", zap.Error(err))
		}
	}()

	// ── 10. Graceful shutdown ──────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutdown initiated")
	// cancel() // besoins avec worker

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutCancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		log.Error("shutdown error", zap.Error(err))
	}

	// _ = signupWriter.Close()
	// _ = signinWriter.Close()
	// _ = dlqWriter.Close()
	log.Info("stopped cleanly")
}
