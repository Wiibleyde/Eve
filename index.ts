import { Logger } from './utils/logger';
import { client, stopBot } from './bot/bot';
import { config } from './utils/config';
import { loadEvents } from './bot/events/event';
import { disconnectDatabase } from './utils/database';
import { initAi } from './utils/intelligence';

export const logger = Logger.init({ minLevel: 'debug' });

async function main() {
    logger.info("Démarrage de l'application...");

    initAi();

    try {
        logger.info('Connexion du bot Discord en cours...');
        loadEvents();
        await client.login(config.DISCORD_TOKEN);
    } catch (error) {
        logger.error(`Erreur lors du démarrage de l'application: ${error}`);
        process.exit(1);
    }
}

// Graceful shutdown
process.on('SIGINT', async () => {
    logger.info('SIGINT reçu, arrêt du bot Discord...');
    await stopBot();
    await disconnectDatabase();
    process.exit(0);
});
process.on('SIGTERM', async () => {
    logger.info('SIGTERM reçu, arrêt du bot Discord...');
    await stopBot();
    await disconnectDatabase();
    process.exit(0);
});

main();
