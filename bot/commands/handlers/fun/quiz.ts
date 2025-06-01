import {
    ActionRowBuilder,
    ButtonBuilder,
    ButtonStyle,
    ChatInputCommandInteraction,
    MessageFlags,
    SlashCommandBuilder,
    TextChannel,
} from 'discord.js';
import type { ICommand } from '../../command';
import { prisma } from '../../../../utils/core/database';
import {
    quizEmbedGenerator,
    quizErrorEmbedGenerator,
    quizes,
    quizMaxTime,
    quizSuccessEmbedGenerator,
    type QuizType,
} from '../../../../utils/games/quiz';
import { logger } from '../../../..';
import { client } from '../../../bot';

export const quiz: ICommand = {
    data: new SlashCommandBuilder()
        .setName('quiz')
        .setDescription('Gestionnaire de quiz')
        .addSubcommand((subcommand) => subcommand.setName('launch').setDescription('Lance un quiz'))
        .addSubcommand((subcommand) =>
            subcommand
                .setName('create')
                .setDescription('Crée un quiz')
                .addStringOption((option) =>
                    option.setName('question').setDescription('La question du quiz').setRequired(true)
                )
                .addStringOption((option) =>
                    option.setName('answer').setDescription('La réponse de la question').setRequired(true)
                )
                .addStringOption((option) =>
                    option.setName('bad1').setDescription('Mauvaise réponse 1').setRequired(true)
                )
                .addStringOption((option) =>
                    option.setName('bad2').setDescription('Mauvaise réponse 2').setRequired(true)
                )
                .addStringOption((option) =>
                    option.setName('bad3').setDescription('Mauvaise réponse 3').setRequired(true)
                )
                .addStringOption((option) =>
                    option.setName('category').setDescription('La catégorie du quiz').setRequired(true)
                )
                .addStringOption((option) =>
                    option
                        .setName('difficulty')
                        .setDescription('La difficulté du quiz')
                        .addChoices(
                            {
                                name: 'Facile',
                                value: 'facile',
                            },
                            {
                                name: 'Normal',
                                value: 'normal',
                            },
                            {
                                name: 'Difficile',
                                value: 'difficile',
                            }
                        )
                        .setRequired(true)
                )
        )
        .addSubcommand((subcommand) =>
            subcommand
                .setName('leaderboard')
                .setDescription('Affiche le classement des meilleurs joueurs de quiz')
                .addStringOption((option) =>
                    option
                        .setName('choice')
                        .setDescription('Choisissez le type de classement')
                        .addChoices(
                            {
                                name: 'Meilleurs scores',
                                value: 'best_scores',
                            },
                            {
                                name: 'Meilleurs ratios',
                                value: 'best_ratios',
                            },
                            {
                                name: 'Pires scores',
                                value: 'worst_scores',
                            }
                        )
                        .setRequired(true)
                )
        ),
    execute: async (interaction) => {
        await interaction.deferReply({ flags: [MessageFlags.Ephemeral] });
        const subcommand = (interaction as ChatInputCommandInteraction).options.getSubcommand();
        switch (subcommand) {
            case 'launch': {
                // Logic to launch a quiz
                const questionCount = await prisma.quizQuestions.count();
                if (questionCount === 0) {
                    await interaction.reply({
                        embeds: [
                            quizErrorEmbedGenerator(
                                "Aucun quiz n'a été créé. Veuillez en créer un avant de lancer un quiz."
                            ),
                        ],
                        flags: [MessageFlags.Ephemeral],
                    });
                    return;
                }
                const randomQuiz = await prisma.quizQuestions.findMany({
                    take: 1,
                    skip: Math.floor(Math.random() * questionCount),
                    include: {
                        author: true,
                    },
                });
                const quizJson = randomQuiz[0];
                if (!quizJson) {
                    await interaction.editReply({
                        embeds: [quizErrorEmbedGenerator('Une erreur est survenue lors de la récupération du quiz.')],
                    });
                    return;
                }
                const quiz: QuizType = {
                    question: quizJson.question,
                    answer: quizJson.answer,
                    badAnswers: [quizJson.badAnswer1, quizJson.badAnswer2, quizJson.badAnswer3],
                    category: quizJson.category,
                    difficulty: quizJson.difficulty,
                    createdAt: Date.now(),
                };

                const shuffledAnswers = [quiz.answer, ...quiz.badAnswers].sort(() => Math.random() - 0.5);
                quiz.shuffleAnswers = shuffledAnswers;

                const formattedCategory =
                    quiz.category.charAt(0).toUpperCase() + quiz.category.slice(1).replace(/_/g, ' ');

                const invalidQuizTimestamp = quiz.createdAt + quizMaxTime;

                const quizEmbed = quizEmbedGenerator()
                    .setTitle('Question de quiz')
                    .setDescription(
                        `${'```'}${quiz.question}${'```'}\n1) ${'`'}${shuffledAnswers[0]}${'`'}\n2) ${'`'}${shuffledAnswers[1]}${'`'}\n3) ${'`'}${shuffledAnswers[2]}${'`'}\n4) ${'`'}${shuffledAnswers[3]}${'`'}`
                    )
                    .addFields(
                        {
                            name: 'Catégorie / difficulté',
                            value: `${formattedCategory} / ${quiz.difficulty}`,
                            inline: true,
                        },
                        { name: 'Invalide', value: `<t:${Math.floor(invalidQuizTimestamp / 1000)}:R>`, inline: true }
                    )
                    .setColor(0x4b0082);

                if (quizJson.author) {
                    quizEmbed.addFields({ name: 'Auteur', value: `<@${quizJson.author.userId}>`, inline: true });
                } else {
                    quizEmbed.addFields({ name: 'Demandé par', value: `<@${interaction.user.id}>`, inline: true });
                }

                const buttons = [
                    new ButtonBuilder().setCustomId('handleQuizButton--1').setLabel('1').setStyle(ButtonStyle.Primary),
                    new ButtonBuilder().setCustomId('handleQuizButton--2').setLabel('2').setStyle(ButtonStyle.Primary),
                    new ButtonBuilder().setCustomId('handleQuizButton--3').setLabel('3').setStyle(ButtonStyle.Primary),
                    new ButtonBuilder().setCustomId('handleQuizButton--4').setLabel('4').setStyle(ButtonStyle.Primary),
                ];

                const actionRow = new ActionRowBuilder<ButtonBuilder>().addComponents(buttons);

                const channel = interaction.channel as TextChannel;
                if (!channel) {
                    await interaction.editReply({
                        embeds: [quizErrorEmbedGenerator("Le canal de discussion n'est pas valide.")],
                    });
                    return;
                }

                const messageResponse = await channel.send({ embeds: [quizEmbed], components: [actionRow] });

                quizes.set(messageResponse.id, quiz);

                await interaction.editReply({
                    embeds: [quizSuccessEmbedGenerator('Quiz lancé avec succès !')],
                });

                await prisma.quizQuestions.update({
                    where: {
                        uuid: quizJson.uuid,
                    },
                    data: {
                        lastTimeUsed: new Date(),
                    },
                });

                break;
            }
            case 'create': {
                const question = interaction.options.get('question')?.value as string;
                const answer = interaction.options.get('answer')?.value as string;
                const badAnswer1 = interaction.options.get('bad1')?.value as string;
                const badAnswer2 = interaction.options.get('bad2')?.value as string;
                const badAnswer3 = interaction.options.get('bad3')?.value as string;
                const category = interaction.options.get('category')?.value as string;
                const difficulty = interaction.options.get('difficulty')?.value as string;

                const guildId = interaction.guildId;
                if (!guildId) {
                    await interaction.editReply({
                        embeds: [quizErrorEmbedGenerator("Impossible de récupérer l'ID du serveur.")],
                    });
                    return;
                }

                const userId = interaction.user.id;
                if (!userId) {
                    await interaction.editReply({
                        embeds: [quizErrorEmbedGenerator("Impossible de récupérer l'ID de l'utilisateur.")],
                    });
                    return;
                }

                try {
                    await prisma.quizQuestions.create({
                        data: {
                            question,
                            answer,
                            badAnswer1,
                            badAnswer2,
                            badAnswer3,
                            category,
                            difficulty,
                            guildId: guildId,
                            author: {
                                connectOrCreate: {
                                    where: {
                                        userId: userId,
                                    },
                                    create: {
                                        userId: userId,
                                    },
                                },
                            },
                        },
                    });
                } catch (error) {
                    // eslint-disable-next-line @typescript-eslint/no-explicit-any
                    if ((error as any).code === 'P2002') {
                        await interaction.editReply({
                            embeds: [
                                quizErrorEmbedGenerator(
                                    'Cette question existe déjà *(Si vous souhaitez la supprimer, contactez un administrateur de Eve)*.'
                                ),
                            ],
                        });
                        return;
                    }
                    logger.error(`Erreur lors de l'ajout de la question : ${error}`);
                    await interaction.editReply({
                        embeds: [quizErrorEmbedGenerator("Une erreur est survenue lors de l'ajout de la question.")],
                    });
                    return;
                }

                await interaction.editReply({ embeds: [quizSuccessEmbedGenerator('Question ajoutée avec succès !')] });
                break;
            }
            case 'leaderboard': {
                const choice = interaction.options.get('choice')?.value as string;
                let users = await prisma.globalUserData.findMany({
                    select: {
                        userId: true,
                        quizGoodAnswers: true,
                        quizBadAnswers: true,
                    },
                });

                users = users.filter((user) => user.quizGoodAnswers + user.quizBadAnswers > 0);
                users = users.filter((user) => user.userId !== client.user?.id);

                let stringChoice = '';

                switch (choice) {
                    case 'best_ratios':
                        stringChoice = 'Meilleurs ratios';
                        users.sort((a, b) => {
                            const ratioA = a.quizGoodAnswers / (a.quizGoodAnswers + a.quizBadAnswers) || 0;
                            const ratioB = b.quizGoodAnswers / (b.quizGoodAnswers + b.quizBadAnswers) || 0;
                            return ratioB - ratioA;
                        });
                        break;
                    case 'best_scores':
                        stringChoice = 'Meilleurs scores';
                        users.sort((a, b) => {
                            return b.quizGoodAnswers - a.quizGoodAnswers;
                        });
                        break;
                    case 'worst_scores':
                        stringChoice = 'Pires scores';
                        users.sort((a, b) => {
                            return b.quizBadAnswers - a.quizBadAnswers;
                        });
                        break;
                }
                users = users.slice(0, 10);

                const leaderboardEmbed = quizEmbedGenerator()
                    .setTitle('Classement des joueurs de quiz')
                    .setDescription(`Voici le classement des joueurs de quiz pour ${stringChoice} :`)
                    .setColor(0x4b0082);
                users.forEach(async (user, index) => {
                    const userId = user.userId;
                    const userTag = await client.users.fetch(userId).then((user) => user.tag).catch(() => 'Utilisateur inconnu');
                    if (!userTag) {
                        return;
                    }
                    const ratio = user.quizGoodAnswers / (user.quizGoodAnswers + user.quizBadAnswers) || 0;
                    leaderboardEmbed.addFields({
                        name: `${index + 1}. ${userTag}`,
                        value: `Score : ${user.quizGoodAnswers} | Mauvais : ${user.quizBadAnswers} | Ratio : ${ratio.toFixed(2)}`,
                        inline: false,
                    });
                });

                await interaction.editReply({
                    embeds: [leaderboardEmbed],
                });
                break;
            }
        }
    },
}
