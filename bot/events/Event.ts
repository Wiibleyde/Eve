import type { ClientEvents } from 'discord.js';
import fs from 'fs';
import path from 'path';
import { client } from '../bot';
import { logger } from '../..';

export interface Event<K extends keyof ClientEvents> {
    name: K;
    once?: boolean;
    execute: (...args: ClientEvents[K]) => void | Promise<void>;
}

export const loadEvents = () => {
    const eventsPath = path.join(__dirname, '../events/handlers');
    const eventFiles = fs.readdirSync(eventsPath).filter((file) => file.endsWith('.ts') || file.endsWith('.js'));

    for (const file of eventFiles) {
        const filePath = path.join(eventsPath, file);
        import(filePath)
            .then((module) => {
                const event = module.event as Event<keyof ClientEvents>;

                if (event.once) {
                    client.once(event.name, (...args) => event.execute(...args));
                } else {
                    client.on(event.name, (...args) => event.execute(...args));
                }

                logger.info(`Loaded event: ${event.name}`);
            })
            .catch((error) => {
                logger.error(`Failed to load event ${file}: ${error}`);
            });
    }
};
