# Eve

![Eve Banner](./eve-banner.png)

![GitHub](https://img.shields.io/github/license/wiibleyde/eve) ![GitHub package.json version](https://img.shields.io/github/package-json/v/wiibleyde/eve) ![GitHub issues](https://img.shields.io/github/issues/wiibleyde/eve) ![GitHub pull requests](https://img.shields.io/github/issues-pr/wiibleyde/eve) ![GitHub top language](https://img.shields.io/github/languages/top/wiibleyde/eve) ![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/wiibleyde/eve) ![GitHub repo size](https://img.shields.io/github/repo-size/wiibleyde/eve) ![GitHub last commit](https://img.shields.io/github/last-commit/wiibleyde/eve) ![GitHub commit activity](https://img.shields.io/github/commit-activity/m/wiibleyde/eve) ![GitHub contributors](https://img.shields.io/github/contributors/wiibleyde/eve)

![GitHub forks](https://img.shields.io/github/forks/wiibleyde/eve?style=social) ![GitHub stars](https://img.shields.io/github/stars/wiibleyde/eve?style=social) ![GitHub watchers](https://img.shields.io/github/watchers/wiibleyde/eve?style=social) ![GitHub followers](https://img.shields.io/github/followers/wiibleyde?style=social)

## About

Eve is a multifunctional Discord bot developed in Go, designed to enhance Discord servers with various features. Inspired by Pixar's WALL-E robot, Eve is efficient and direct while remaining warm and curious.

## Main Features

### 🎮 Twitch Integration

- Automatic notifications when streamers start and end their streams
- Display of rich embeds with streamer information and direct links to channels
- Support for specific role mentions for notifications

### 🤖 Artificial Intelligence

- Interaction with Google's Gemini AI for intelligent responses
- Ability to mention and respond to users naturally
- Support for interactions in threads and direct messages

### 🎲 Games and Entertainment

- Quiz system with player statistics tracking
- Integrated games like Motus for more entertainment
- Image quote generator to create fun content

### 📊 Server Management

- Advanced logging system with Discord webhook support
- Administrative commands to manage bot features
- Maintenance mode for updates or maintenance

## Installation

### Prerequisites

- Go (recent version)
- MySQL or any other database system compatible with Prisma
- A Discord developer account for tokens
- [Optional] A Twitch developer account for Twitch integration
- [Optional] A Google API key for Gemini

### Docker Configuration

1. Clone the repository:

```bash
git clone https://github.com/wiibleyde/eve.git
cd eve
```

2. Configure the docker-compose.yml file with your tokens and IDs:

```yml
services:
  eve:
    build: .
    image: wiibleyde/eve:latest
    container_name: eve
    restart: unless-stopped
    volumes:
      - ./logs:/root/logs
    environment:
      - DISCORD_TOKEN=your_discord_token
      - DISCORD_CLIENT_ID=your_discord_client_id
      - EVE_HOME_GUILD=your_main_server_id
      - MP_CHANNEL=your_mp_channel_id
      - BLAGUE_API_TOKEN=your_joke_api_token
      - OWNER_ID=your_user_id
      - LOGS_WEBHOOK_URL=your_webhook_url_for_logs
      - DATABASE_URL=mysql://user:password@host:3306/eve
      - GOOGLE_API_KEY=your_google_api_key
      - TWITCH_CLIENT_ID=your_twitch_client_id
      - TWITCH_CLIENT_SECRET=your_twitch_client_secret
```

3. Launch the bot with Docker Compose:

```bash
docker-compose up -d
```

### Manual Configuration

1. Clone the repository:

```bash
git clone https://github.com/wiibleyde/eve.git
cd eve
```

2. Install dependencies:

```bash
go get
```

3. Configure your Prisma database:

```bash
npx prisma generate
```

4. Create a `.env` file with the variables listed in the docker-compose.yml

5. Launch the bot:

```bash
go run main.go
```

## Commands

The bot offers several Discord slash commands:

- `/streamer` - Manage streamers followed by the bot
- `/quiz` - Interact with the quiz system
- More commands available in the server

## Logs and Maintenance

Eve generates log files in the `logs/` directory that can be used to monitor the bot's behavior and diagnose problems.

## Contributing

Contributions are welcome! Check out our [contribution guide](CONTRIBUTING.md) for more information on how to participate in Eve's development.

## License

This project is under the GNU General Public License v2.0 (GPL-2.0). See the [LICENSE](LICENSE) file for more details.
