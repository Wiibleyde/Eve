import { errorEmbed, successEmbed } from '@/utils/embeds';
import { creatEmbedForRadio } from '@/utils/rp/radio';
import { hasPermission } from '@/utils/permissionTester';
import {
    SlashCommandOptionsOnlyBuilder,
    SlashCommandBuilder,
    CommandInteraction,
    TextChannel,
    PermissionFlagsBits,
    MessageFlags,
} from 'discord.js';

export const data: SlashCommandOptionsOnlyBuilder = new SlashCommandBuilder()
    .setName('createradio')
    .setDescription('Créer un message pour la radio')
    .addStringOption((option) => option.setName('nom').setDescription('Nom du service').setRequired(true));

/**
 * Executes the command to create or update a radio configuration.
 *
 * @param interaction - The command interaction object.
 * @returns A promise that resolves to void.
 *
 * This function performs the following steps:
 * 1. Checks if the user has the required permissions.
 * 2. Retrieves the guild ID and service name from the interaction options.
 * 3. Checks if a radio configuration already exists for the guild.
 * 4. If the radio configuration does not exist, it creates a new one and sends a message to the channel.
 * 5. If the radio configuration exists, it updates the existing message in the channel.
 * 6. Sends a success reply to the interaction.
 */
export async function execute(interaction: CommandInteraction): Promise<void> {
    if (!(await hasPermission(interaction, [PermissionFlagsBits.Administrator], false))) {
        await interaction.editReply({
            embeds: [errorEmbed(interaction, new Error("Vous n'avez pas la permission de changer la configuration."))],
        });
        return;
    }
    const serviceName = interaction.options.get('nom')?.value as string;
    const channel = interaction.channel as TextChannel;

    const { embed, actionRow, files } = creatEmbedForRadio(interaction, serviceName, [
        { name: serviceName, frequency: '0.0' },
    ]);

    await channel.send({ embeds: [embed], components: [actionRow], files: files });

    await interaction.reply({
        embeds: [successEmbed(interaction, `La radio du ${serviceName} a été créée avec succès !`)],
        flags: [MessageFlags.Ephemeral],
    });
}
