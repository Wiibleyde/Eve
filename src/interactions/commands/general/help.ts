import { CommandInteraction, EmbedBuilder, MessageFlags, SlashCommandBuilder } from 'discord.js';
import { commands } from '@/interactions/commands/index';
import { backSpace } from '@/utils/textUtils';

interface Command {
    data: {
        description: string;
    };
}
interface Commands {
    [key: string]: Command;
}

export const data: SlashCommandBuilder = new SlashCommandBuilder()
    .setName('help')
    .setDescription('Affiche la liste des commandes');

/**
 * Executes the help command, generating an embed with a list of available commands and their descriptions.
 *
 * @param interaction - The interaction object representing the command invocation.
 * @returns A promise that resolves when the reply is sent.
 */
export async function execute(interaction: CommandInteraction): Promise<void> {
    const typedCommands: Commands = commands;
    const commandsList = Object.keys(typedCommands)
        .map((key: string) => {
            return `**/${key}** : ${typedCommands[key].data.description}`;
        })
        .join(backSpace);

    const helpEmbed = new EmbedBuilder()
        .setTitle('Liste des commandes')
        .setDescription(commandsList)
        .setColor(0x4b0082)
        .setTimestamp()
        .setFooter({ text: `Eve – Toujours prête à vous aider.`, iconURL: interaction.client.user.displayAvatarURL() });

    await interaction.reply({ embeds: [helpEmbed], flags: [MessageFlags.Ephemeral] });
}
