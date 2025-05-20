import {
    CommandInteraction,
    SlashCommandBuilder,
    type SlashCommandOptionsOnlyBuilder,
    type SlashCommandSubcommandsOnlyBuilder,
} from 'discord.js';
import { ping } from './handlers/ping';
import { talk } from './handlers/talk';
import { birthday } from './handlers/birthday';
import { blague } from './handlers/blague';
import { config } from './handlers/config';

// Type pour représenter une commande avec sa définition et son exécution
export interface ICommand {
    data: SlashCommandBuilder | SlashCommandOptionsOnlyBuilder | SlashCommandSubcommandsOnlyBuilder;
    execute: (interaction: CommandInteraction) => Promise<void>;
}

// Collection des commandes disponibles
export const commands: ICommand[] = [ping, talk, birthday, blague, config];

// Map pour accéder aux commandes rapidement par leur nom
export const commandsMap = new Map<string, ICommand>(commands.map((command) => [command.data.name, command]));
