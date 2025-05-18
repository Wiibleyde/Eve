import { ActivityType, Events } from "discord.js";
import type { Event } from "../event";
import { deployCommands } from "../../commands/deploy-commands";
import { logger } from "../../..";

export const event: Event<Events.ClientReady> = {
    name: Events.ClientReady,
    once: true,
    execute: async (client) => {
        client.user?.setPresence({
            status: 'dnd',
            activities: [
                {
                    name: `le démarrage...`,
                    type: ActivityType.Watching,
                },
            ],
        })

        await deployCommands();

        logger.info(`Logged in as ${client.user?.tag}`);
    }
};