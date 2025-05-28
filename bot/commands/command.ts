import {
    CommandInteraction,
    SlashCommandBuilder,
    type SlashCommandOptionsOnlyBuilder,
    type SlashCommandSubcommandsOnlyBuilder,
} from 'discord.js';
import { ping } from './handlers/ping';
import { talk } from './handlers/talk';
import { birthday } from './handlers/birthday';
import { config } from './handlers/config';
import { streamer } from './handlers/streamer';
import { blague } from './handlers/fun/blague';
import { motus } from './handlers/fun/motus';
import { quiz } from './handlers/fun/quiz';
import { quote } from './handlers/quote';
import { debug } from './handlers/debug';
import { maintenance } from './handlers/maintenance';
import { lsms } from './handlers/rp/lsms';

// Type pour représenter une commande avec sa définition et son exécution
export interface ICommand {
    data: SlashCommandBuilder | SlashCommandOptionsOnlyBuilder | SlashCommandSubcommandsOnlyBuilder;
    execute: (interaction: CommandInteraction) => Promise<void>;
}

// Collection des commandes disponibles
export const commands: ICommand[] = [
    ping,
    talk,
    birthday,
    blague,
    config,
    streamer,
    motus,
    quiz,
    quote,
    debug,
    maintenance,
    lsms,
];

// Map pour accéder aux commandes rapidement par leur nom
export const commandsMap = new Map<string, ICommand>(commands.map((command) => [command.data.name, command]));
