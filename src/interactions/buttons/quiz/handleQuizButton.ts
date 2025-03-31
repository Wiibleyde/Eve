import { prisma } from '@/utils/database';
import { errorEmbed } from '@/utils/embeds';
import { backSpace } from '@/utils/textUtils';
import { ActionRowBuilder, ButtonBuilder, ButtonInteraction, ButtonStyle, CacheType, MessageFlags } from 'discord.js';

export const quizes: Map<string, QuizType> = new Map();

export const maxTime = 28800000; // 8 hours

/**
 * Handles the interaction when a user clicks a quiz button.
 *
 * @param interaction - The interaction object representing the button click.
 * @returns A promise that resolves to void.
 *
 * This function performs the following steps:
 * 1. Retrieves the message associated with the interaction.
 * 2. Checks if the message exists; if not, replies with an error.
 * 3. Retrieves the quiz associated with the message ID.
 * 4. Checks if the quiz exists; if not, attempts to retrieve the question from the database and replies with an error if not found.
 * 5. Checks if the quiz has expired; if so, replies with an expiration message.
 * 6. Checks if the user has already answered the quiz correctly or incorrectly and replies accordingly.
 * 7. Retrieves the user's answer and compares it with the correct answer.
 * 8. Updates the quiz data and user data in the database based on whether the answer was correct or incorrect.
 * 9. Edits the original message to update the list of users who answered correctly or incorrectly.
 */
export async function handleQuizButton(interaction: ButtonInteraction<CacheType>): Promise<void> {
    const message = interaction.message;
    if (!message) {
        await interaction.reply({
            embeds: [
                errorEmbed(
                    interaction,
                    new Error(
                        'Une erreur est survenue lors de la récupération de la question de quiz (message introuvable).'
                    )
                ),
            ],
            ephemeral: true,
        });
        return;
    }
    const quiz = quizes.get(message.id);
    if (!quiz) {
        const questionTxt = message.embeds[0].description?.split('```')[1];
        if (!questionTxt) {
            await interaction.reply({
                embeds: [
                    errorEmbed(
                        interaction,
                        new Error(
                            'Une erreur est survenue lors de la récupération de la question de quiz (description introuvable).'
                        )
                    ),
                ],
                ephemeral: true,
            });
            return;
        }
        const question = await prisma.quizQuestions.findFirst({
            where: {
                question: questionTxt,
            },
        });
        if (!question) {
            await interaction.reply({
                embeds: [
                    errorEmbed(
                        interaction,
                        new Error(
                            'Une erreur est survenue lors de la récupération de la question de quiz (question introuvable).'
                        )
                    ),
                ],
                ephemeral: true,
            });
            return;
        }
        await interaction.reply({
            embeds: [errorEmbed(interaction, new Error('Quiz expiré (la réponse était: ||' + question.answer + '||)'))],
            ephemeral: true,
        });
        return;
    }

    if (quiz.createdAt + maxTime < Date.now()) {
        await interaction.reply({
            embeds: [errorEmbed(interaction, new Error('Quiz expiré (la réponse était: ||' + quiz.answer + '||)'))],
            ephemeral: true,
        });
        return;
    }

    if (quiz.rightUsers && quiz.rightUsers.includes(interaction.user.id)) {
        await interaction.reply({ content: 'Vous avez déjà répondu correctement à cette question !', flags: [MessageFlags.Ephemeral] });
        return;
    }
    if (quiz.wrongUsers && quiz.wrongUsers.includes(interaction.user.id)) {
        await interaction.reply({
            content: 'Vous avez déjà répondu incorrectement à cette question !',
            ephemeral: true,
        });
        return;
    }

    const answer = quiz.answer;
    if (!quiz.shuffleAnswers) {
        await interaction.reply({
            embeds: [
                errorEmbed(
                    interaction,
                    new Error('Une erreur est survenue lors de la récupération des réponses de la question de quiz.')
                ),
            ],
            ephemeral: true,
        });
        return;
    }
    const userAnswer = quiz.shuffleAnswers[parseInt(interaction.customId.split('--')[1]) - 1];

    const buttons = [
        new ButtonBuilder().setCustomId('handleQuizButton--1').setLabel('1').setStyle(ButtonStyle.Primary),
        new ButtonBuilder().setCustomId('handleQuizButton--2').setLabel('2').setStyle(ButtonStyle.Primary),
        new ButtonBuilder().setCustomId('handleQuizButton--3').setLabel('3').setStyle(ButtonStyle.Primary),
        new ButtonBuilder().setCustomId('handleQuizButton--4').setLabel('4').setStyle(ButtonStyle.Primary),
        new ButtonBuilder()
            .setCustomId('reportQuestionButton')
            .setLabel('Signaler une erreur')
            .setStyle(ButtonStyle.Danger),
    ];

    const actionRow = new ActionRowBuilder<ButtonBuilder>().addComponents(buttons);

    if (userAnswer === answer) {
        quiz.rightUsers = quiz.rightUsers ? [...quiz.rightUsers, interaction.user.id] : [interaction.user.id];
        await interaction.reply({ content: 'Bonne réponse !', flags: [MessageFlags.Ephemeral] });
        const messageFields = message.embeds[0].fields;
        let found = false;
        for (const field of messageFields) {
            if (field.name === 'Bonne(s) réponse(s)') {
                let finalField = '';
                for (const user of quiz.rightUsers) {
                    finalField = finalField + `${backSpace}<@${user}>`;
                }
                field.value = finalField;
                found = true;
                break;
            }
        }
        if (!found) {
            messageFields.push({
                name: 'Bonne(s) réponse(s)',
                value: quiz.rightUsers.map((user) => `<@${user}>`).join(backSpace),
                inline: true,
            });
        }
        const user = await prisma.globalUserData.findUnique({
            where: {
                userId: interaction.user.id,
            },
        });
        if (!user) {
            await prisma.globalUserData.create({
                data: {
                    userId: interaction.user.id,
                    quizBadAnswers: 0,
                    quizGoodAnswers: 1,
                },
            });
        } else {
            await prisma.globalUserData.update({
                where: {
                    userId: interaction.user.id,
                },
                data: {
                    quizGoodAnswers: {
                        increment: 1,
                    },
                },
            });
        }
        await message.edit({ embeds: [message.embeds[0]], components: [actionRow] });
    } else {
        quiz.wrongUsers = quiz.wrongUsers ? [...quiz.wrongUsers, interaction.user.id] : [interaction.user.id];
        await interaction.reply({
            content: 'Mauvaise réponse ! (La bonne réponse était: ||' + answer + '||)',
            ephemeral: true,
        });
        const messageFields = message.embeds[0].fields;
        let found = false;
        for (const field of messageFields) {
            if (field.name === 'Mauvaise(s) réponse(s)') {
                let finalField = '';
                for (const user of quiz.wrongUsers) {
                    finalField = finalField + `${backSpace}<@${user}>`;
                }
                field.value = finalField;
                found = true;
                break;
            }
        }
        if (!found) {
            messageFields.push({
                name: 'Mauvaise(s) réponse(s)',
                value: quiz.wrongUsers.map((user) => `<@${user}>`).join(backSpace),
                inline: true,
            });
        }
        const user = await prisma.globalUserData.findUnique({
            where: {
                userId: interaction.user.id,
            },
        });
        if (!user) {
            await prisma.globalUserData.create({
                data: {
                    userId: interaction.user.id,
                    quizBadAnswers: 1,
                    quizGoodAnswers: 0,
                },
            });
        } else {
            await prisma.globalUserData.update({
                where: {
                    userId: interaction.user.id,
                },
                data: {
                    quizBadAnswers: {
                        increment: 1,
                    },
                },
            });
        }
        await message.edit({ embeds: [message.embeds[0]], components: [actionRow] });
    }
}
