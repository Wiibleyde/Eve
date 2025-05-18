import { CommandInteraction, type SlashCommandOptionsOnlyBuilder } from "discord.js";
import { ping } from "./utils/ping";
import { talk } from "./utils/talk";

// Type pour représenter une commande avec sa définition et son exécution
export interface ICommand {
    data: SlashCommandOptionsOnlyBuilder;
    execute: (interaction: CommandInteraction) => Promise<void>;
}

// Collection des commandes disponibles
export const commands: ICommand[] = [ping, talk];

// Map pour accéder aux commandes rapidement par leur nom
export const commandsMap = new Map<string, ICommand>(
    commands.map(command => [command.data.name, command])
);