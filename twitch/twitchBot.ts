import tmi from 'tmi.js';
import { config } from '../utils/core/config';
import { Logger } from '../utils/core/logger';
import { handleTwitchCommands } from './commands/commandHandler.js';

const logger = Logger.init({
    minLevel: 'debug',
    discordMinLevel: 'info',
    discordWebhook: config.LOGS_WEBHOOK_URL,
    showFileInfo: false,
});

export interface TwitchClient {
    client: tmi.Client;
    isConnected: boolean;
}

let twitchClient: TwitchClient | null = null;

/**
 * Initialize and connect to Twitch chat
 */
export async function initTwitchBot(): Promise<void> {
    if (twitchClient?.isConnected) {
        logger.warn('Le bot Twitch est déjà connecté');
        return;
    }

    // Vérifier si le token d'accès Twitch est configuré
    if (!config.TWITCH_ACCESS_TOKEN) {
        logger.warn('TWITCH_ACCESS_TOKEN non configuré. Le bot Twitch ne sera pas démarré.');
        logger.info('Consultez twitch/README.md pour les instructions de configuration.');
        return;
    }

    // Configuration du client Twitch
    const client = new tmi.Client({
        options: { debug: false },
        connection: {
            reconnect: true,
            secure: true
        },
        identity: {
            username: 'E.V.E', // Remplacez par votre nom d'utilisateur Twitch
            password: `oauth:${config.TWITCH_ACCESS_TOKEN}`
        },
        channels: ['E.V.E'] // Remplacez par votre canal Twitch
    });

    // Event handlers
    client.on('connected', (addr, port) => {
        logger.info(`Bot Twitch connecté à ${addr}:${port}`);
        if (twitchClient) {
            twitchClient.isConnected = true;
        }
    });

    client.on('disconnected', (reason) => {
        logger.warn(`Bot Twitch déconnecté: ${reason}`);
        if (twitchClient) {
            twitchClient.isConnected = false;
        }
    });

    client.on('message', async (channel, userstate, message, self) => {
        // Ignore messages from the bot itself
        if (self) return;

        // Log the message
        logger.debug(`[${channel}] ${userstate.username}: ${message}`);

        // Handle commands (messages starting with !)
        if (message.startsWith('!')) {
            await handleTwitchCommands(client, channel, userstate, message);
        }
    });

    try {
        await client.connect();
        twitchClient = {
            client,
            isConnected: true
        };
        logger.info('Bot Twitch initialisé avec succès');
    } catch (error) {
        logger.error(`Erreur lors de la connexion au bot Twitch: ${error}`);
        throw error;
    }
}

/**
 * Disconnect from Twitch chat
 */
export async function stopTwitchBot(): Promise<void> {
    if (!twitchClient?.isConnected) {
        logger.warn('Le bot Twitch n\'est pas connecté');
        return;
    }

    try {
        await twitchClient.client.disconnect();
        twitchClient.isConnected = false;
        logger.info('Bot Twitch déconnecté avec succès');
    } catch (error) {
        logger.error(`Erreur lors de la déconnexion du bot Twitch: ${error}`);
    }
}

/**
 * Get the current Twitch client
 */
export function getTwitchClient(): tmi.Client | null {
    return twitchClient?.client || null;
}

/**
 * Check if the Twitch bot is connected
 */
export function isTwitchConnected(): boolean {
    return twitchClient?.isConnected || false;
}
