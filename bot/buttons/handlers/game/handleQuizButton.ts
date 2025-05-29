import { ActionRowBuilder, ButtonBuilder, ButtonStyle, MessageFlags, type ButtonInteraction } from 'discord.js';
import { quizErrorEmbedGenerator, quizes, quizMaxTime } from '../../../../utils/games/quiz';
import { prisma } from '../../../../utils/core/database';

export async function handleQuizButton(interaction: ButtonInteraction): Promise<void> {
    await interaction.deferReply({ withResponse: true, flags: [MessageFlags.Ephemeral] });
    const message = interaction.message;
    if (!message) {
        await interaction.editReply({
            embeds: [quizErrorEmbedGenerator('Aucun message trouvé pour le quiz.')],
        });
        return;
    }

    const quiz = quizes.get(message.id);
    if (!quiz) {
        const questionTxt = message.embeds[0]?.description?.split('```')[1];
        if (!questionTxt) {
            await interaction.editReply({
                embeds: [
                    quizErrorEmbedGenerator(
                        'Une erreur est survenue lors de la récupération de la question de quiz (description introuvable).'
                    ),
                ],
            });
            return;
        }
        const question = await prisma.quizQuestions.findFirst({
            where: {
                question: questionTxt,
            },
        });
        if (!question) {
            await interaction.editReply({
                embeds: [
                    quizErrorEmbedGenerator(
                        'Une erreur est survenue lors de la récupération de la question de quiz (question introuvable).'
                    ),
                ],
            });
            return;
        }
        await interaction.editReply({
            embeds: [
                quizErrorEmbedGenerator(
                    `Le quiz a expiré ou n'existe pas. Mais la réponse à la question était : ${question.answer}`
                ),
            ],
        });
        return;
    }
    if (quiz.createdAt + quizMaxTime < Date.now()) {
        await interaction.editReply({
            embeds: [quizErrorEmbedGenerator('Le quiz a expiré. Veuillez en lancer un nouveau.')],
        });
        return;
    }
    if (quiz.rightUsers && quiz.rightUsers.includes(interaction.user.id)) {
        await interaction.editReply({
            embeds: [quizErrorEmbedGenerator('Vous avez déjà répondu correctement à cette question.')],
        });
        return;
    }
    if (quiz.wrongUsers && quiz.wrongUsers.includes(interaction.user.id)) {
        await interaction.editReply({
            embeds: [quizErrorEmbedGenerator('Vous avez déjà répondu incorrectement à cette question.')],
        });
        return;
    }
    if (!quiz.shuffleAnswers) {
        await interaction.editReply({
            embeds: [
                quizErrorEmbedGenerator(
                    'Une erreur est survenue lors de la récupération des réponses de la question de quiz.'
                ),
            ],
        });
        return;
    }

    const answer = quiz.answer;
    const userAnswerIndexStr = interaction.customId.split('--')[1] ?? '0';
    const userAnswerIndex = parseInt(userAnswerIndexStr) - 1;
    const userAnswer = quiz.shuffleAnswers[userAnswerIndex];

    const isCorrect = userAnswer === answer;
    const responseContent = isCorrect ? 'Bonne réponse !' : `Mauvaise réponse. La bonne réponse était : \`${answer}\``;
    const fieldName = isCorrect ? 'Bonne(s) réponse(s)' : 'Mauvaise(s) réponse(s)';
    const usersList = isCorrect ? 'rightUsers' : 'wrongUsers';

    quiz[usersList] = quiz[usersList] ? [...quiz[usersList], interaction.user.id] : [interaction.user.id];

    const messageFields = message.embeds[0]?.fields ?? [];
    const fieldIndex = messageFields.findIndex((field) => field.name === fieldName);

    if (fieldIndex >= 0) {
        if (messageFields[fieldIndex]) {
            messageFields[fieldIndex].value = ((quiz[usersList] ?? []) as string[])
                .map((user) => `<@${user}>`)
                .join('\n');
        }
    } else {
        messageFields.push({
            name: fieldName,
            value: quiz[usersList].map((user) => `<@${user}>`).join('\n'),
            inline: true,
        });
    }

    const statField = isCorrect ? 'quizGoodAnswers' : 'quizBadAnswers';

    // Utiliser upsert pour optimiser l'accès à la base de données
    await prisma.globalUserData.upsert({
        where: {
            userId: interaction.user.id,
        },
        update: {
            [statField]: {
                increment: 1,
            },
        },
        create: {
            userId: interaction.user.id,
            quizGoodAnswers: isCorrect ? 1 : 0,
            quizBadAnswers: isCorrect ? 0 : 1,
        },
    });

    // Update the message with modified fields regardless of answer correctness
    if (message.embeds[0]) {
        const buttons = [
            new ButtonBuilder().setCustomId('handleQuizButton--1').setLabel('1').setStyle(ButtonStyle.Primary),
            new ButtonBuilder().setCustomId('handleQuizButton--2').setLabel('2').setStyle(ButtonStyle.Primary),
            new ButtonBuilder().setCustomId('handleQuizButton--3').setLabel('3').setStyle(ButtonStyle.Primary),
            new ButtonBuilder().setCustomId('handleQuizButton--4').setLabel('4').setStyle(ButtonStyle.Primary),
        ];

        const actionRow = new ActionRowBuilder<ButtonBuilder>().addComponents(buttons);

        // Create a new embed object that keeps all original properties but updates fields
        const updatedEmbed = {
            ...message.embeds[0].toJSON(), // Preserve all original properties
            fields: messageFields,
        };

        // Execute the update in parallel with the response to the interaction
        await Promise.all([
            message.edit({
                embeds: [updatedEmbed],
                components: [actionRow],
            }),
            interaction.editReply({ content: responseContent }),
        ]);
        return;
    }

    await interaction.editReply({ content: responseContent });
}
