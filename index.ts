import { Logger } from './utils/core/logger';
import { client, endingScripts, stopBot } from './bot/bot';
import { config } from './utils/core/config';
import { loadEvents } from './bot/events/event';
import { disconnectDatabase, connectDatabase } from './utils/core/database';
import { initAi } from './utils/intelligence';
import { initTwitchBot, stopTwitchBot } from './twitch/twitchBot';

export const logger = Logger.init({
    minLevel: 'debug',
    discordMinLevel: 'info',
    discordWebhook: config.LOGS_WEBHOOK_URL,
    showFileInfo: false,
});

async function main() {
    logger.info("Démarrage de l'application...");

    // Test de la connexion à la base de données
    logger.info('Test de la connexion à la base de données...');
    const isDatabaseConnected = await connectDatabase();
    if (!isDatabaseConnected) {
        logger.error('Impossible de se connecter à la base de données. Arrêt de l\'application.');
        process.exit(1);
    }
    logger.info('Connexion à la base de données réussie.');

    initAi();

    try {
        logger.info('Connexion du bot Discord en cours...');
        loadEvents();
        await client.login(config.DISCORD_TOKEN);
        
        // Initialiser le bot Twitch
        logger.info('Initialisation du bot Twitch...');
        try {
            await initTwitchBot();
            logger.info('Bot Twitch initialisé avec succès.');
        } catch (twitchError) {
            logger.warn(`Impossible d'initialiser le bot Twitch: ${twitchError}`);
            logger.info('Le bot Discord continuera de fonctionner normalement.');
        }
    } catch (error) {
        logger.error(`Erreur lors du démarrage de l'application: ${error}`);
        process.exit(1);
    }
}

// Graceful shutdown
process.on('SIGINT', async () => {
    logger.info('SIGINT reçu, arrêt du bot Discord et Twitch...');
    await endingScripts();
    await stopBot();
    await stopTwitchBot();
    await disconnectDatabase();
    process.exit(0);
});
process.on('SIGTERM', async () => {
    logger.info('SIGTERM reçu, arrêt du bot Discord et Twitch...');
    await endingScripts();
    await stopBot();
    await stopTwitchBot();
    await disconnectDatabase();
    process.exit(0);
});

process.on('uncaughtException', (error) => {
    logger.error(`Exception non gérée: ${error}`);
});

process.on('unhandledRejection', (reason) => {
    logger.error(`Rejet de promesse non géré: ${reason}`);
});


main();
