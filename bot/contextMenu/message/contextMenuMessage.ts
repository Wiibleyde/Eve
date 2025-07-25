import { ContextMenuCommandBuilder, MessageContextMenuCommandInteraction } from 'discord.js';
import { createQuote } from './handlers/createQuote';
import type { IBaseCommand } from '../../commands/command';

// Type pour représenter une commande avec sa définition et son exécution
export interface IContextMenuMessageCommand extends IBaseCommand {
    data: ContextMenuCommandBuilder;
    execute: (interaction: MessageContextMenuCommandInteraction) => Promise<void>;
}

// Collection des commandes disponibles
export const messageContextMenuCommands: IContextMenuMessageCommand[] = [createQuote];

// Map pour accéder aux commandes rapidement par leur nom
export const messageContextMenuCommandsMap = new Map<string, IContextMenuMessageCommand>(
    messageContextMenuCommands.map((command) => [command.data.name, command])
);
