import { ContextMenuCommandBuilder, UserContextMenuCommandInteraction } from 'discord.js';
import { profilePicture } from './handlers/profilePicture';
import { banner } from './handlers/banner';
import type { IBaseCommand } from '../../commands/command';

// Type pour représenter une commande avec sa définition et son exécution
export interface IContextMenuUserCommand extends IBaseCommand {
    data: ContextMenuCommandBuilder;
    execute: (interaction: UserContextMenuCommandInteraction) => Promise<void>;
}

// Collection des commandes disponibles
export const userContextMenuCommands: IContextMenuUserCommand[] = [profilePicture, banner];

// Map pour accéder aux commandes rapidement par leur nom
export const userContextMenuCommandsMap = new Map<string, IContextMenuUserCommand>(
    userContextMenuCommands.map((command) => [command.data.name, command])
);
