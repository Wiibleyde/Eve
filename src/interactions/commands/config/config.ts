import {
    CommandInteraction,
    ChatInputCommandInteraction,
    EmbedBuilder,
    SlashCommandBuilder,
    PermissionFlagsBits,
    InteractionContextType,
    ChannelType,
    MessageFlags,
} from 'discord.js';
import { prisma } from '@/utils/database';
import { errorEmbed, successEmbed } from '@/utils/embeds';
import { backSpace } from '@/utils/textUtils';
import { hasPermission } from '@/utils/permissionTester';

export const data = new SlashCommandBuilder()
    .setName('config')
    .setDescription('Configurer les salons')
    .addSubcommand(subcommand =>
        subcommand
            .setName('view')
            .setDescription('Voir la configuration actuelle')
    )
    .addSubcommand(subcommand =>
        subcommand
            .setName('edit')
            .setDescription('Modifier la configuration')
            .addStringOption((option) =>
                option.setName('key').setDescription('Clé de configuration').setRequired(true).addChoices(
                    {
                        name: 'Salon des anniversaires',
                        value: 'birthdayChannel',
                    },
                    {
                        name: 'Salon des citations',
                        value: 'quoteChannel',
                    },
                    {
                        name: 'Catégorie des formations',
                        value: 'trainingCategory',
                    }
                )
            )
            .addChannelOption((option) =>
                option.setName('channel').setDescription('Salon ou catégorie à configurer').setRequired(true)
            )
    )
    .setContexts([InteractionContextType.Guild, InteractionContextType.PrivateChannel]);

/**
 * Handles the execution of the configuration command for managing channels.
 *
 * @param interaction - The command interaction object.
 * @returns A promise that resolves to void.
 *
 * This function performs the following actions:
 * 1. Defers the reply to the interaction.
 * 2. Checks if the user has the required permissions to manage channels.
 * 3. Based on the subcommand specified in the interaction, it either:
 *    - Views the current channel configuration and sends it as a reply.
 *    - Edits the channel configuration by updating or creating a new entry in the database.
 *
 * The function handles errors and sends appropriate error messages if:
 * - The user does not have the required permissions.
 * - No configuration is found when attempting to view.
 */
export async function execute(interaction: CommandInteraction): Promise<void> {
    await interaction.deferReply({ withResponse: true, flags: [MessageFlags.Ephemeral] });
    if (!(await hasPermission(interaction, [PermissionFlagsBits.ManageChannels], false))) {
        await interaction.editReply({
            embeds: [errorEmbed(interaction, new Error("Vous n'avez pas la permission de changer la configuration."))],
        });
        return;
    }
    const subcommand = (interaction as ChatInputCommandInteraction).options.getSubcommand();

    switch (subcommand) {
        case 'view': {
            const serverConfig = await prisma.config.findMany({
                where: {
                    guildId: interaction.guildId?.toString() as string,
                },
            });
            if (serverConfig.length === 0) {
                await interaction.editReply({
                    embeds: [errorEmbed(interaction, new Error('Aucune configuration trouvée.'))],
                });
                return;
            }
            const responseEmbed = new EmbedBuilder()
                .setTitle('Configuration des salons')
                .setColor(0x00ff00)
                .setDescription(serverConfig.map((config) => `**${config.key}**: <#${config.value}>`).join(backSpace))
                .setTimestamp()
                .setFooter({
                    text: `Eve – Toujours prête à vous aider.`,
                    iconURL: interaction.client.user.displayAvatarURL(),
                });

            await interaction.editReply({ embeds: [responseEmbed] });
            break;
        }
        case 'edit': {
            const key = interaction.options.get('key', true).value as string;
            const channel = interaction.options.get('channel', true).channel;

            const existingConfig = await prisma.config.findFirst({
                where: {
                    guildId: interaction.guildId?.toString() as string,
                    key,
                },
            });

            if (key === 'trainingCategory') {
                const channelData = await interaction.guild?.channels.fetch(channel?.id as string);
                if (channelData?.type !== ChannelType.GuildCategory) {
                    await interaction.editReply({
                        embeds: [errorEmbed(interaction, new Error("Le salon fourni n'est pas une catégorie."))],
                    });
                    return;
                }
            }

            if (existingConfig) {
                await prisma.config.update({
                    where: {
                        uuid: existingConfig.uuid,
                    },
                    data: {
                        value: channel?.id,
                    },
                });
            } else {
                if(!channel) {
                    await interaction.editReply({
                        embeds: [errorEmbed(interaction, new Error('Aucun salon fourni.'))],
                    });
                    return;
                }
                await prisma.config.create({
                    data: {
                        guildId: interaction.guildId?.toString() as string,
                        key,
                        value: channel?.id,
                    },
                });
            }
            await interaction.editReply({ embeds: [successEmbed(interaction, 'Configuration mise à jour.')] });
            break;
        }
    }
}
