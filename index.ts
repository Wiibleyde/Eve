import { Logger } from './utils/logger';
import { client } from './bot/bot';
import { config } from './utils/config';

export const logger = Logger.init({ minLevel: 'debug' });

async function main() {
    logger.info("Démarrage de l'application...");

    try {
        logger.info('Connexion du bot Discord en cours...');
        await client.login(config.DISCORD_TOKEN);
    } catch (error) {
        logger.error(`Erreur lors du démarrage de l'application: ${error}`);
        process.exit(1);
    }
}

main();
