import { Client, GatewayIntentBits, Partials } from 'discord.js';
import { prepareLsmsSummary } from '../utils/rp/lsms';
import { logger } from '..';
import { Player } from 'discord-player';
import { YoutubeiExtractor } from 'discord-player-youtubei';

export const client = new Client({
    intents: [
        GatewayIntentBits.Guilds,
        GatewayIntentBits.GuildMessages,
        GatewayIntentBits.GuildMembers,
        GatewayIntentBits.GuildVoiceStates,
        GatewayIntentBits.GuildMessageTyping,
        GatewayIntentBits.DirectMessages,
        GatewayIntentBits.DirectMessageTyping,
        GatewayIntentBits.MessageContent,
    ],
    partials: [Partials.User, Partials.Channel, Partials.Message, Partials.GuildMember],
});

export const player = new Player(client, {});
player.extractors.register(YoutubeiExtractor, {});

export async function stopBot() {
    await client.destroy();
}

export async function endingScripts() {
    // Send a summary of the LSMS duty before rebooting
    logger.info('Lancement des scripts de fin...');

    // LSMS Summary
    await prepareLsmsSummary();

    logger.info('Fin des scripts de fin...');
}
