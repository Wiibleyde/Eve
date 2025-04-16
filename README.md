# Eve

![Eve Banner](./eve-banner.png)

![GitHub](https://img.shields.io/github/license/wiibleyde/eve) ![GitHub package.json version](https://img.shields.io/github/package-json/v/wiibleyde/eve) ![GitHub issues](https://img.shields.io/github/issues/wiibleyde/eve) ![GitHub pull requests](https://img.shields.io/github/issues-pr/wiibleyde/eve)![GitHub top language](https://img.shields.io/github/languages/top/wiibleyde/eve) ![GitHub code size in bytes](https://img.shields.io/github/languages/code-size/wiibleyde/eve) ![GitHub repo size](https://img.shields.io/github/repo-size/wiibleyde/eve) ![GitHub last commit](https://img.shields.io/github/last-commit/wiibleyde/eve) ![GitHub commit activity](https://img.shields.io/github/commit-activity/m/wiibleyde/eve) ![GitHub contributors](https://img.shields.io/github/contributors/wiibleyde/eve)

![GitHub forks](https://img.shields.io/github/forks/wiibleyde/eve?style=social) ![GitHub stars](https://img.shields.io/github/stars/wiibleyde/eve?style=social) ![GitHub watchers](https://img.shields.io/github/watchers/wiibleyde/eve?style=social) ![GitHub followers](https://img.shields.io/github/followers/wiibleyde?style=social)

# COMPLETE REWRITE OF THE BOT IN GOLANG

For now, the bot is working in TypeScript, but I will rewrite it in Golang to improve performance and maintainability.

## Description

Eve is a simple discord bot that can be used for multiple things, including IA functionalities, and more.

## Table of Contents

- [Description](#description)
- [Table of Contents](#table-of-contents)
- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [License](#license)
- [Authors](#authors)
- [Acknowledgments](#acknowledgments)

## Features

- [x] IA functionalities
  - [x] Chat with the bot by simply mentioning it (with a real discussion)
- [x] Music fonctionalities
  - [x] Play music from youtube
  - [x] Pause/Resume music
  - [x] Skip music
  - [x] Stop music
  - [x] Queue management
- [x] Birthday reminder
- [x] Quiz functionalities
- [x] Fun commands
  - [x] cat pictures
  - [x] dog pictures
  - [x] jokes
  - [x] quotes
- [x] Debug functionalities

## Prerequisites

- Node.js
- Yarn
- ffmpeg
- MariaDB database
  - *Windows docker don't work well with the Prisma ORM*
- Discord bot token
  - With good intents

## Installation

1. Clone the repository
2. Install the dependencies with `pnpm`
3. Create a `.env` file at the root of the project with the following content:

    ```env
    DISCORD_TOKEN=your_discord_token
    DISCORD_CLIENT_ID=your_discord_bot_client_id

    EVE_HOME_GUILD=glados_home_guild_id
    REPORT_CHANNEL=report_channel_id
    MP_CHANNEL=mp_channel_id

    OWNER_ID=your_discord_user_id

    LOGS_WEBHOOK_URL=your_discord_webhook_url

    DATABASE_URL="mysql_connection_url"

    GOOGLE_API_KEY=your_google_api_key

    BLAGUE_API_TOKEN=your_blague_api_token
    ```

4. Run the prisma migrations with `yarn prisma migrate deploy`
5. Build the project with `yarn build`
6. Start the bot with `yarn start`

## License

This project is licensed under the GPL-2.0 License - see the [LICENSE](LICENSE) file for details

## Authors

- [**Wiibleyde**](https://github.com/wiibleyde)

## Acknowledgments

- [**Discord.js**](https://discord.js.org/)
- [**Prisma ORM**](https://www.prisma.io/)
- [**Gemini**](https://gemini.google.com/)
- [**Discord Player**](https://discord-player.js.org/)
- [**ZerioDev/Music-bot**](https://github.com/ZerioDev/Music-bot)

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

![built-with-love](https://forthebadge.com/images/badges/built-with-love.svg)
![made-with-typescript](https://forthebadge.com/images/badges/made-with-typescript.svg)
![open-source](https://forthebadge.com/images/badges/open-source.svg)
![uses-git](https://forthebadge.com/images/badges/uses-git.svg)
