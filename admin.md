D'après la migration `001_initial_schema.sql` que nous avons écrite :

```
Email    : admin@myapp.local
Password : Admin@1234
```

Le hash bcrypt dans le SQL est :
```
$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy
```

Ce hash correspond au mot de passe **`Admin@1234`** (coût bcrypt = 10).

> ⚠️ C'est un compte de seed pour le développement uniquement. En production, changer immédiatement le mot de passe ou supprimer ce seed via la migration.