import type { ClientEvents } from 'discord.js';
import { client } from '../bot';
import { logger } from '../..';

// Import all event handlers directly
import { interactionCreateEvent } from './handlers/InteractionCreate';
import { clientReadyEvent } from './handlers/ClientReady';
import { messageCreateEvent } from './handlers/MessageCreate';
// Import any other event handlers you have

export interface Event<K extends keyof ClientEvents> {
    name: K;
    once?: boolean;
    execute: (...args: ClientEvents[K]) => void | Promise<void>;
}

// Use a more generic approach to avoid type issues
const eventHandlers = [interactionCreateEvent, clientReadyEvent, messageCreateEvent];

export const loadEvents = () => {
    try {
        for (const event of eventHandlers) {
            if (event.once) {
                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                client.once(event.name, (...args) => event.execute(...(args as any)));
            } else {
                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                client.on(event.name, (...args) => event.execute(...(args as any)));
            }
        }

        logger.info(`Total events loaded: ${eventHandlers.length}`);
    } catch (error) {
        logger.error(`Failed to load events: ${error}`);
    }
};
