import type { ClientEvents } from 'discord.js';
import { client } from '../bot';
import { logger } from '../..';

// Import all event handlers directly
import { event as interactionCreateEvent } from './handlers/InteractionCreate';
import { event as clientReadyEvent } from './handlers/ClientReady';
// Import any other event handlers you have

export interface Event<K extends keyof ClientEvents> {
    name: K;
    once?: boolean;
    execute: (...args: ClientEvents[K]) => void | Promise<void>;
}

// Use a more generic approach to avoid type issues
const eventHandlers = [
    interactionCreateEvent,
    clientReadyEvent,
    // Add other event handlers here
];

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

            logger.info(`Loaded event: ${event.name}`);
        }

        logger.info(`Total events loaded: ${eventHandlers.length}`);
    } catch (error) {
        logger.error(`Failed to load events: ${error}`);
    }
};
