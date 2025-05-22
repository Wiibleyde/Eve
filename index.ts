import { Logger } from './utils/logger';
import { client } from './bot/bot';
import { config } from './utils/config';
import { loadEvents } from './bot/events/event';

export const logger = Logger.init({ minLevel: 'debug' });

async function main() {
    logger.info("Démarrage de l'application...");

    try {
        logger.info('Connexion du bot Discord en cours...');
        loadEvents();
        await client.login(config.DISCORD_TOKEN);
    } catch (error) {
        logger.error(`Erreur lors du démarrage de l'application: ${error}`);
        process.exit(1);
    }
}

main();

//TODO: PREPARE FOR COMPILATION USING : bun build index.ts --compile --outfile eve-bot // WHICH COMPILE THE CODE AND CREATE A BINARY FILE
