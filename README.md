# ops-server — Backend ETL Event-Driven

Backend Go production-ready : REST API + Kafka + Redis + PostgreSQL + Swagger.

---

## Stack

| Composant    | Technologie                          |
|--------------|--------------------------------------|
| Langage      | Go 1.22                              |
| HTTP         | Gin                                  |
| ORM          | GORM + PostgreSQL                    |
| Cache/Locks  | Redis (go-redis v9)                  |
| Messaging    | Apache Kafka (kafka-go)              |
| Auth         | JWT (golang-jwt/v5) + refresh Redis  |
| Docs         | Swagger (swaggo/swag)                |
| Logs         | Zap (JSON structurés)                |
| Métriques    | Prometheus                           |

---

## Démarrage rapide

```bash
# 1. Copier la config d'environnement
cp .env.example .env

# 2. Lancer l'infrastructure (Postgres, Redis, Kafka, Kafka UI)
make docker-up

# 3. Appliquer les migrations SQL
make migrate-up

# 4. Générer la doc Swagger
make swagger

# 5. Lancer l'API
make run
```

Accès :
- **API** → http://localhost:8080
- **Swagger UI** → http://localhost:8080/swagger/index.html
- **Health** → http://localhost:8080/health
- **Kafka UI** → http://localhost:8090
- **Prometheus** → http://localhost:8080/metrics/prometheus

---

## Structure

```
ops-server/
├── cmd/api/main.go              # Entrée — DI + lifecycle + graceful shutdown
├── configs/                     # Config YAML + loader Viper (env: OPS_SERVER_*)
├── docs/                        # Swagger auto-généré (swag init)
├── deployments/                 # Dockerfile multi-stage + docker-compose
├── scripts/migrations/          # SQL migrations + script bash
├── pkg/                         # Librairies partagées (errors, logger, utils)
└── internal/
    ├── domain/
    │   ├── user/                # Modèles: User, Role, Permission, UserRole
    │   ├── notification/        # Notifications multi-canal
    │   ├── metrics/             # Métriques + Events
    │   └── audit/               # Audit trails + Logs persistés
    ├── infrastructure/
    │   ├── postgres/            # GORM + transactions context-aware
    │   ├── redis/               # Cache + Lock distribué (Lua) + Rate limiter
    │   └── kafka/               # Core générique + consumers/producers/handlers
    └── interfaces/
        ├── http/                # Middleware (auth JWT, RBAC, logging) + routes
        └── workers/             # WorkerPool Kafka (goroutines + panic recovery)
```

---

## Domaines

### User
- Inscription / Connexion / Refresh / Logout
- RBAC multi-rôles (admin, ops, user, viewer)
- Permissions granulaires (`resource:action`)
- Table de jointure `user_roles` many2many

### Notification
- Types : email, push, in_app, sms
- Statuts : pending → sent / failed → read
- Comptage des non-lus

### Metrics & Events
- Métriques : counter, gauge, histogram avec labels JSONB
- Événements : info, warning, critical avec filtres temporels

### Audit
- Audit trails : CREATE, UPDATE, DELETE, READ, LOGIN, LOGOUT
- Logs applicatifs persistés avec niveau et tracing

---

## Variables d'environnement (préfixe `OPS_SERVER_`)

```bash
OPS_SERVER_DATABASE_HOST=localhost
OPS_SERVER_DATABASE_PASSWORD=ops-server_secret
OPS_SERVER_JWT_SECRET=your-secret-here
OPS_SERVER_REDIS_HOST=localhost
OPS_SERVER_KAFKA_BROKERS=localhost:9092
```

---

## Commandes utiles

```bash
make build          # Compiler le binaire
make test           # Tests unitaires
make test-coverage  # Coverage HTML
make swagger        # Régénérer la doc Swagger
make lint           # golangci-lint
make mock           # Régénérer les mocks mockgen
make docker-up      # Démarrer l'infra
make docker-down    # Arrêter l'infra
make migrate-up     # Appliquer les migrations SQL
```

---

## Sécurité

- JWT HS256 (access 15min + refresh 7j, rotation à chaque refresh)
- Refresh token stocké en Redis, révocable immédiatement
- RBAC multi-rôles via middleware Gin
- Soft-delete sur toutes les entités sensibles
- Bcrypt pour les mots de passe

---

## Règles architecture

- ❌ Logique métier dans les controllers
- ❌ Accès DB hors repository
- ❌ Couplage fort entre domaines
- ✅ Kafka = transport uniquement
- ✅ Idempotence consumer (Redis event-ID)
- ✅ Retry x3 + DLQ sur chaque consumer
