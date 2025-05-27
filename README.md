# Eve

![Eve Banner](./eve-banner.png)

![GitHub](https://img.shields.io/github/license/wiibleyde/eve) ![GitHub package.json version](https://img.shields.io/github/package-json/v/wiibleyde/eve) ![GitHub issues](https://img.shields.io/github/issues/wiibleyde/eve) ![GitHub pull requests](https://img.shields.io/github/issues-pr/wiibleyde/eve) ![GitHub top language](https://img.shields.io/github/languages/top/wiibleyde/eve) ![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/wiibleyde/eve) ![GitHub repo size](https://img.shields.io/github/repo-size/wiibleyde/eve) ![GitHub last commit](https://img.shields.io/github/last-commit/wiibleyde/eve) ![GitHub commit activity](https://img.shields.io/github/commit-activity/m/wiibleyde/eve) ![GitHub contributors](https://img.shields.io/github/contributors/wiibleyde/eve)

![GitHub forks](https://img.shields.io/github/forks/wiibleyde/eve?style=social) ![GitHub stars](https://img.shields.io/github/stars/wiibleyde/eve?style=social) ![GitHub watchers](https://img.shields.io/github/watchers/wiibleyde/eve?style=social) ![GitHub followers](https://img.shields.io/github/followers/wiibleyde?style=social)

## À propos

Eve est un bot Discord multifonctionnel développé en TypeScript avec Discord.js, conçu pour améliorer les serveurs Discord avec diverses fonctionnalités. Inspiré du robot Eve de Pixar (WALL-E), ce bot est efficace et direct tout en restant chaleureux et curieux.

## Fonctionnalités

### Commandes principales
- `/ping` - Vérifier si le bot est en ligne et son temps de réponse
- `/talk` - Faire parler le bot dans un canal ou en MP à un utilisateur
- `/quote` - Gérer un système de citations mémorables
- `/config` - Configurer les paramètres du bot pour votre serveur
- `/maintenance` - Mettre le bot en mode maintenance (réservé aux propriétaires)
- `/debug` - Commandes de débogage avancées (réservé aux propriétaires)

### Système d'anniversaires
- `/birthday set` - Définir votre date d'anniversaire
- `/birthday get` - Voir votre date d'anniversaire enregistrée
- `/birthday remove` - Supprimer votre date d'anniversaire
- `/birthday list` - Afficher la liste des anniversaires de la communauté
- Notification automatique des anniversaires via un système de cron

### Divertissement
- `/quiz` - Système de quiz avec création et gestion de questions
- `/blague` - Obtenir des blagues aléatoires
- `/motus` - Jouer au jeu Motus dans Discord

### Intégrations
- **Twitch** - Notifications de stream avec `/streamer add/remove`

## Installation

### Prérequis
- Node.js 22+ (+ Bun)
- Base de données MariaDB ou MySQL
- Un compte Discord et un bot créé sur le [portail développeur Discord](https://discord.com/developers/applications)

### Configuration
1. Clonez le dépôt :
   ```bash
   git clone https://github.com/wiibleyde/eve.git
   cd eve
   ```

2. Installez les dépendances :
   ```bash
   bun install
   ```

3. Créez un fichier `.env` à la racine du projet avec les variables suivantes :
   ```
   DISCORD_TOKEN=votre_token_discord
   DISCORD_CLIENT_ID=votre_client_id_discord
   EVE_HOME_GUILD=id_du_serveur_principal
   MP_CHANNEL=id_du_canal_mp
   BLAGUE_API_TOKEN=votre_token_api_blague
   OWNER_ID=votre_id_discord
   LOGS_WEBHOOK_URL=url_webhook_logs
   DATABASE_URL=mysql://user:password@localhost:3306/eve
   GOOGLE_API_KEY=votre_cle_api_google
   TWITCH_CLIENT_ID=votre_client_id_twitch
   TWITCH_CLIENT_SECRET=votre_secret_client_twitch
   ```

4. Initialisez la base de données :
   ```bash
   npx prisma migrate dev
   ```

5. Démarrez le bot :
   ```bash
   bun index.ts
   ```

### Déploiement avec Docker
Vous pouvez également déployer Eve avec Docker en utilisant le `docker-compose.yml` fourni :
```bash
docker-compose up -d
```

## Structure du projet
- `bot/` - Logique principale du bot et gestionnaires de commandes
- `utils/` - Utilitaires et fonctions d'aide
- `cron/` - Tâches programmées (anniversaires, statuts, etc.)
- `prisma/` - Schéma de base de données et migrations
- `assets/` - Ressources graphiques et polices

## Contribution

Les contributions sont les bienvenues ! Consultez notre [guide de contribution](CONTRIBUTING.md) pour plus d'informations sur la façon de participer au développement d'Eve.

## Licence

Ce projet est sous licence GNU General Public License v2.0 (GPL-2.0). Voir le fichier [LICENSE](LICENSE) pour plus de détails.
