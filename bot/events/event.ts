import type { ClientEvents } from 'discord.js';
import { client } from '../bot';
import { logger } from '../..';
import { interactionCreateEvent } from './handlers/interactionCreate';
import { clientReadyEvent } from './handlers/clientReady';
import { messageCreateEvent } from './handlers/messageCreate';
import { guildMemberUpdateEvent } from './handlers/guildMemberUpdate';

export interface Event<K extends keyof ClientEvents> {
    name: K;
    once?: boolean;
    execute: (...args: ClientEvents[K]) => void | Promise<void>;
}

const eventHandlers = [interactionCreateEvent, clientReadyEvent, messageCreateEvent, guildMemberUpdateEvent];

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
