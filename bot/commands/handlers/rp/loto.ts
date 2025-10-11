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
    PermissionFlagsBits,
    SlashCommandBuilder,
} from 'discord.js';

const PRIZE_OPTIONS = Array.from({ length: 9 }, (_, index) => ({
    name: `prize${index + 1}`,
    description: `Gain n°${index + 1} ${index === 0 ? '(obligatoire)' : '(optionnel)'}`,
    required: index === 0,
}));

const PRIZE_OPTION_NAMES = PRIZE_OPTIONS.map((option) => option.name);
const REQUIRED_PRIZE_OPTIONS = PRIZE_OPTIONS.filter((option) => option.required);
const OPTIONAL_PRIZE_OPTIONS = PRIZE_OPTIONS.filter((option) => !option.required);

export const loto: ICommand = {
    data: new SlashCommandBuilder()
        .setName('loto')
        .setDescription('[SABS] Commandes du loto')
        .addSubcommand((subcommand) => {
            subcommand
                .setName('create')
                .setDescription('Créer un nouveau loto')
                .addStringOption((option) =>
                    option.setName('name').setDescription('Nom du loto').setRequired(true).setMaxLength(50)
                );

            REQUIRED_PRIZE_OPTIONS.forEach((prizeOption) => {
                subcommand.addStringOption((option) =>
                    option
                        .setName(prizeOption.name)
                        .setDescription(prizeOption.description)
                        .setRequired(prizeOption.required)
                        .setMaxLength(255)
                );
            });

            subcommand.addStringOption((option) =>
                option
                    .setName('ticketprice')
                    .setDescription("Prix d'un ticket (défaut: 500)")
                    .setRequired(false)
                    .setMaxLength(10)
            );

            subcommand.addIntegerOption((option) =>
                option
                    .setName('cooldown')
                    .setDescription('Cooldown entre deux achats pour un même joueur (en minutes, défaut: 0)')
                    .setRequired(false)
                    .setMinValue(0)
                    .setMaxValue(10080)
            );

            subcommand.addIntegerOption((option) =>
                option
                    .setName('maxtickets')
                    .setDescription('Nombre maximum de tickets par achat (obligatoire si un cooldown est défini)')
                    .setRequired(false)
                    .setMinValue(1)
                    .setMaxValue(1000)
            );

            OPTIONAL_PRIZE_OPTIONS.forEach((prizeOption) => {
                subcommand.addStringOption((option) =>
                    option
                        .setName(prizeOption.name)
                        .setDescription(prizeOption.description)
                        .setRequired(prizeOption.required)
                        .setMaxLength(255)
                );
            });

            return subcommand;
        }),
    guildIds: ['1396778821570793572', config.EVE_HOME_GUILD], // SABS et EVE Home
    execute: async (interaction: ChatInputCommandInteraction) => {
        await interaction.deferReply({});

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
                const cooldownMinutesInput = interaction.options.getInteger('cooldown');
                const cooldownMinutes = cooldownMinutesInput ?? 0;
                const maxTicketsInput = interaction.options.getInteger('maxtickets');
                const prizes = PRIZE_OPTION_NAMES.map((optionName) =>
                    interaction.options.getString(optionName)?.trim()
                ).filter((value): value is string => Boolean(value && value.length > 0));

                if (isNaN(ticketPrice) || ticketPrice <= 0) {
                    await interaction.editReply({
                        embeds: [errorEmbedGenerator('Le prix du ticket doit être un nombre entier positif.')],
                    });
                    return;
                }

                if (cooldownMinutes < 0) {
                    await interaction.editReply({
                        embeds: [errorEmbedGenerator('Le cooldown doit être positif.')],
                    });
                    return;
                }

                if (cooldownMinutes > 0 && (maxTicketsInput == null || maxTicketsInput <= 0)) {
                    await interaction.editReply({
                        embeds: [
                            errorEmbedGenerator(
                                'Vous devez spécifier un nombre maximum de tickets par achat strictement positif lorsque vous définissez un cooldown.'
                            ),
                        ],
                    });
                    return;
                }

                if (maxTicketsInput != null && maxTicketsInput <= 0) {
                    await interaction.editReply({
                        embeds: [
                            errorEmbedGenerator(
                                'Le nombre maximum de tickets doit être un entier strictement positif.'
                            ),
                        ],
                    });
                    return;
                }

                if (prizes.length === 0) {
                    await interaction.editReply({
                        embeds: [errorEmbedGenerator('Vous devez fournir au moins un gain.')],
                    });
                    return;
                }

                if (prizes.length > PRIZE_OPTION_NAMES.length) {
                    await interaction.editReply({
                        embeds: [
                            errorEmbedGenerator(
                                `Le nombre de gains ne peut pas dépasser ${PRIZE_OPTION_NAMES.length}.`
                            ),
                        ],
                    });
                    return;
                }

                const oversizedPrize = prizes.find((prize) => prize.length > 255);
                if (oversizedPrize) {
                    await interaction.editReply({
                        embeds: [
                            errorEmbedGenerator(
                                `Le gain "${oversizedPrize.slice(0, 50)}..." dépasse la longueur maximale autorisée (255 caractères).`
                            ),
                        ],
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
                        cooldownMinutes,
                        maxTicketsPerPurchase: maxTicketsInput ?? null,
                        isActive: true,
                        prizes: {
                            create: prizes.map((label, index) => ({
                                label,
                                position: index,
                            })),
                        },
                    },
                    include: {
                        prizes: {
                            orderBy: { position: 'asc' },
                            include: {
                                winnerPlayer: {
                                    select: {
                                        name: true,
                                    },
                                },
                            },
                        },
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
