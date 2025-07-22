import tmi from 'tmi.js';
import { Logger } from '../../utils/core/logger';
import { config } from '../../utils/core/config';

const logger = Logger.init({
    minLevel: 'debug',
    discordMinLevel: 'info',
    discordWebhook: config.LOGS_WEBHOOK_URL,
    showFileInfo: false,
});

/**
 * Interface pour les commandes Twitch
 */
export interface TwitchCommand {
    name: string;
    description: string;
    execute: (client: tmi.Client, channel: string, userstate: tmi.Userstate, args: string[]) => Promise<void>;
}

/**
 * Collection des commandes Twitch
 */
const commands = new Map<string, TwitchCommand>();

/**
 * Commande de test
 */
const testCommand: TwitchCommand = {
    name: 'test',
    description: 'Commande de test pour vérifier que Eve fonctionne sur Twitch',
    execute: async (client: tmi.Client, channel: string, userstate: tmi.Userstate) => {
        const username = (userstate as Record<string, string>)['display-name'] || (userstate as Record<string, string>)['username'] || 'Anonyme';
        const response = `@${username} Salut ! Je suis Eve, et je fonctionne parfaitement sur Twitch ! 🤖✨`;
        
        try {
            await client.say(channel, response);
            logger.info(`Commande !test exécutée par ${username} dans ${channel}`);
        } catch (error) {
            logger.error(`Erreur lors de l'exécution de la commande !test: ${error}`);
        }
    }
};

/**
 * Commande d'aide
 */
const helpCommand: TwitchCommand = {
    name: 'help',
    description: 'Affiche la liste des commandes disponibles',
    execute: async (client: tmi.Client, channel: string, userstate: tmi.Userstate) => {
        const username = (userstate as Record<string, string>)['display-name'] || (userstate as Record<string, string>)['username'] || 'Anonyme';
        const commandList = Array.from(commands.keys()).map(cmd => `!${cmd}`).join(', ');
        const response = `@${username} Commandes disponibles: ${commandList}`;
        
        try {
            await client.say(channel, response);
            logger.info(`Commande !help exécutée par ${username} dans ${channel}`);
        } catch (error) {
            logger.error(`Erreur lors de l'exécution de la commande !help: ${error}`);
        }
    }
};

/**
 * Commande d'information sur Eve
 */
const eveCommand: TwitchCommand = {
    name: 'eve',
    description: 'Affiche des informations sur Eve',
    execute: async (client: tmi.Client, channel: string, userstate: tmi.Userstate) => {
        const username = (userstate as Record<string, string>)['display-name'] || (userstate as Record<string, string>)['username'] || 'Anonyme';
        const response = `@${username} Je suis Eve 🤖 Un bot multi-plateforme présent sur Discord et Twitch ! Je peux vous aider avec diverses tâches. Utilisez !help pour voir mes commandes.`;
        
        try {
            await client.say(channel, response);
            logger.info(`Commande !eve exécutée par ${username} dans ${channel}`);
        } catch (error) {
            logger.error(`Erreur lors de l'exécution de la commande !eve: ${error}`);
        }
    }
};

// Enregistrer les commandes
commands.set('test', testCommand);
commands.set('help', helpCommand);
commands.set('eve', eveCommand);

/**
 * Gère les commandes Twitch
 */
export async function handleTwitchCommands(
    client: tmi.Client,
    channel: string,
    userstate: tmi.Userstate,
    message: string
): Promise<void> {
    // Parse la commande et les arguments
    const args = message.slice(1).trim().split(' ');
    const commandName = args.shift()?.toLowerCase();

    if (!commandName) return;

    // Trouve et exécute la commande
    const command = commands.get(commandName);
    if (command) {
        try {
            await command.execute(client, channel, userstate, args);
        } catch (error) {
            logger.error(`Erreur lors de l'exécution de la commande ${commandName}: ${error}`);
        }
    }
}
