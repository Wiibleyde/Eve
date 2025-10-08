import type { ICommand } from '@bot/commands/command';
import { getSubcommand } from '@bot/utils/commandOptions';
import { errorEmbedGenerator } from '@bot/utils/embeds';
import { config } from '@utils/core/config';
import { prisma } from '@utils/core/database';
import { hasPermission } from '@utils/permission';
import { generateLotoEmbed } from '@utils/rp/loto';
import {
    ActionRowBuilder,
    ButtonBuilder,
    ButtonStyle,
    ChatInputCommandInteraction,
    MessageFlags,
    PermissionFlagsBits,
    SlashCommandBuilder,
} from 'discord.js';

export const loto: ICommand = {
    data: new SlashCommandBuilder()
        .setName('loto')
        .setDescription('[SABS] Commandes du loto')
        .addSubcommand((subcommand) =>
            subcommand
                .setName('create')
                .setDescription('Créer un nouveau loto')
                .addStringOption((option) =>
                    option.setName('name').setDescription('Nom du loto').setRequired(true).setMaxLength(50)
                )
                .addStringOption((option) =>
                    option
                        .setName('ticketprice')
                        .setDescription("Prix d'un ticket (défaut: 500)")
                        .setRequired(false)
                        .setMaxLength(10)
                )
        ),
    guildIds: ['1396778821570793572', config.EVE_HOME_GUILD], // SABS et EVE Home
    execute: async (interaction: ChatInputCommandInteraction) => {
        await interaction.deferReply({
            flags: [MessageFlags.Ephemeral],
        });

        if (!(await hasPermission(interaction, [PermissionFlagsBits.PinMessages]))) {
            await interaction.editReply({
                embeds: [errorEmbedGenerator("Vous n'avez pas la permission d'épingler les messages")],
            });
            return;
        }

        const subcommand = getSubcommand(interaction);
        switch (subcommand) {
            case 'create': {
                const name = interaction.options.getString('name', true).trim();
                const ticketPriceInput = interaction.options.getString('ticketprice')?.trim();
                const ticketPrice = ticketPriceInput ? parseInt(ticketPriceInput, 10) : 500;

                if (isNaN(ticketPrice) || ticketPrice <= 0) {
                    await interaction.editReply({
                        embeds: [errorEmbedGenerator('Le prix du ticket doit être un nombre entier positif.')],
                    });
                    return;
                }

                if (name.length === 0) {
                    await interaction.editReply({
                        embeds: [errorEmbedGenerator('Le nom du loto ne peut pas être vide.')],
                    });
                    return;
                }

                if (name.length > 50) {
                    await interaction.editReply({
                        embeds: [errorEmbedGenerator('Le nom du loto ne peut pas dépasser 50 caractères.')],
                    });
                    return;
                }

                const existingGame = await prisma.lotoGames.findFirst({
                    where: {
                        name: name,
                        isActive: true,
                    },
                });

                if (existingGame) {
                    await interaction.editReply({
                        embeds: [errorEmbedGenerator('Un loto avec ce nom est déjà actif.')],
                    });
                    return;
                }

                const game = await prisma.lotoGames.create({
                    data: {
                        name,
                        ticketPrice,
                        isActive: true,
                    },
                });

                const embed = generateLotoEmbed(game, []);

                const buttons = [
                    new ButtonBuilder()
                        .setCustomId(`lotoBuy--${game.uuid}`)
                        .setLabel('Ajouter des tickets')
                        .setEmoji('🎟️')
                        .setStyle(ButtonStyle.Primary),
                    new ButtonBuilder()
                        .setCustomId(`lotoDraw--${game.uuid}`)
                        .setLabel('Tirer le gagnant')
                        .setEmoji('⚠️')
                        .setStyle(ButtonStyle.Danger),
                    new ButtonBuilder()
                        .setCustomId(`removeTickets--${game.uuid}`)
                        .setLabel('Retirer des tickets')
                        .setEmoji('🗑️')
                        .setStyle(ButtonStyle.Secondary),
                    new ButtonBuilder()
                        .setCustomId(`lotoEditPlayer--${game.uuid}`)
                        .setLabel("Modifier le nom d'un participant")
                        .setEmoji('✏️')
                        .setStyle(ButtonStyle.Secondary),
                ];
                const actionRow = new ActionRowBuilder<ButtonBuilder>().addComponents(buttons);

                await interaction.editReply({
                    content: `Nouveau loto créé par <@${interaction.user.id}> !`,
                    embeds: [embed],
                    components: [actionRow],
                });
                break;
            }
            default: {
                await interaction.editReply({
                    embeds: [errorEmbedGenerator('Sous-commande inconnue.')],
                });
                break;
            }
        }
    },
};
