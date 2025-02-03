import { prisma } from '@/utils/database';
import { errorEmbed } from '@/utils/embeds';
import {
    ActionRowBuilder,
    ButtonInteraction,
    CacheType,
    ModalActionRowComponentBuilder,
    ModalBuilder,
    TextInputBuilder,
    TextInputStyle,
} from 'discord.js';

/**
 * Handles the interaction when a user clicks the "report question" button in a quiz.
 *
 * This function retrieves the quiz question from the message embed, searches for the question
 * in the database, and if found, presents a modal to the user to report the question.
 *
 * @param interaction - The button interaction that triggered this function.
 *
 * @returns A promise that resolves when the interaction is handled.
 *
 * @throws Will reply with an error embed if the message, question description, or quiz question is not found.
 */
export async function reportQuestionButton(interaction: ButtonInteraction<CacheType>): Promise<void> {
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
    const question = message.embeds[0].description?.split('```')[1];
    if (!question) {
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
    const quiz = await prisma.quizQuestions.findFirst({
        where: {
            question: question,
        },
    });
    if (!quiz) {
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

    const modal = new ModalBuilder().setCustomId('reportQuestionModal--' + quiz.uuid).setTitle(quiz.uuid);

    const reportReason = new TextInputBuilder()
        .setCustomId('reportReason')
        .setPlaceholder('Raison du signalement')
        .setLabel('Raison du signalement')
        .setStyle(TextInputStyle.Paragraph)
        .setRequired(true);

    const actionRow = new ActionRowBuilder<ModalActionRowComponentBuilder>().addComponents(reportReason);
    modal.addComponents(actionRow);

    await interaction.showModal(modal);
}
