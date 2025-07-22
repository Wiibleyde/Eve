# Configuration Twitch pour Eve

Pour que Eve puisse se connecter au chat Twitch, vous devez configurer les variables d'environnement suivantes dans votre fichier `.env` :

## Variables d'environnement Twitch requises

```env
# Twitch Configuration
TWITCH_CLIENT_ID=votre_client_id
TWITCH_CLIENT_SECRET=votre_client_secret
TWITCH_ACCESS_TOKEN=votre_token_d_acces
```

## Comment obtenir les credentials Twitch

### 1. Créer une application Twitch

1. Allez sur [Twitch Developers Console](https://dev.twitch.tv/console/apps)
2. Connectez-vous avec votre compte Twitch
3. Cliquez sur "Register Your Application"
4. Remplissez les informations :
   - **Name** : Nom de votre bot (ex: "Eve Bot")
   - **OAuth Redirect URLs** : `http://localhost:3000` (pour les tests)
   - **Category** : Chat Bot
5. Cliquez sur "Create"
6. Notez le **Client ID** et générez un **Client Secret**

### 2. Obtenir un Access Token

Vous avez plusieurs options :

#### Option A : Utiliser Twitch CLI (Recommandé)
```bash
# Installer Twitch CLI
# macOS
brew install twitchdev/twitch/twitch-cli

# Configurer avec vos credentials
twitch configure

# Générer un token pour chat
twitch token -u -s 'chat:read chat:edit'
```

#### Option B : Utiliser l'URL d'autorisation manuelle
Remplacez `YOUR_CLIENT_ID` par votre Client ID et visitez cette URL :
```
https://id.twitch.tv/oauth2/authorize?client_id=YOUR_CLIENT_ID&redirect_uri=http://localhost:3000&response_type=token&scope=chat:read+chat:edit
```

### 3. Configuration du bot

1. Modifiez le fichier `twitch/twitchBot.ts` :
   ```typescript
   identity: {
       username: 'votre_nom_utilisateur_twitch',
       password: `oauth:${config.TWITCH_ACCESS_TOKEN}`
   },
   channels: ['votre_canal_twitch']
   ```

2. Remplacez :
   - `votre_nom_utilisateur_twitch` par votre nom d'utilisateur Twitch
   - `votre_canal_twitch` par le nom du canal où le bot doit être présent

## Commandes disponibles

Une fois configuré, Eve répondra aux commandes suivantes dans le chat Twitch :

- `!test` : Commande de test pour vérifier que le bot fonctionne
- `!help` : Affiche la liste des commandes disponibles

## Dépannage

- **Erreur 401** : Vérifiez que votre token d'accès est valide
- **Bot ne répond pas** : Vérifiez que le nom d'utilisateur et le canal sont corrects
- **Erreur de connexion** : Vérifiez votre connexion internet et les credentials

## Sécurité

⚠️ **Important** : Ne partagez jamais vos tokens d'accès. Ajoutez `.env` à votre `.gitignore` pour éviter de les commit accidentellement.
