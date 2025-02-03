import { CommandInteraction, SlashCommandBuilder, SlashCommandOptionsOnlyBuilder } from 'discord.js';
import { prisma } from '@/utils/database';
import { successEmbed, errorEmbed } from '@/utils/embeds';
import { logger } from '@/index';

export const data: SlashCommandOptionsOnlyBuilder = new SlashCommandBuilder()
    .setName('addquestion')
    .setDescription("Permet d'ajouter une question au quiz")
    .addStringOption((option) => option.setName('question').setDescription('La question du quiz').setRequired(true))
    .addStringOption((option) => option.setName('answer').setDescription('La réponse de la question').setRequired(true))
    .addStringOption((option) => option.setName('bad1').setDescription('Mauvaise réponse 1').setRequired(true))
    .addStringOption((option) => option.setName('bad2').setDescription('Mauvaise réponse 2').setRequired(true))
    .addStringOption((option) => option.setName('bad3').setDescription('Mauvaise réponse 3').setRequired(true))
    .addStringOption((option) => option.setName('category').setDescription('La catégorie du quiz').setRequired(true))
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
    );

/**
 * Handles the execution of the add question command for a quiz.
 *
 * @param interaction - The command interaction object.
 *
 * @remarks
 * This function retrieves the question, answer, bad answers, category, and difficulty from the interaction options.
 * It then attempts to add the question to the database. If the question already exists, it sends an error message.
 * If the question is added successfully, it sends a success message.
 *
 * @throws Will send an error message if the guild ID or user ID cannot be retrieved.
 * Will also send an error message if there is an issue with adding the question to the database.
 */
export async function execute(interaction: CommandInteraction): Promise<void> {
    await interaction.deferReply({ ephemeral: true });
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
            embeds: [errorEmbed(interaction, new Error("Impossible de récupérer l'ID du serveur."))],
        });
        return;
    }

    const userId = interaction.user.id;
    if (!userId) {
        await interaction.editReply({
            embeds: [errorEmbed(interaction, new Error("Impossible de récupérer l'ID de l'utilisateur."))],
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
                    errorEmbed(
                        interaction,
                        new Error(
                            'Cette question existe déjà *(Si vous souhaitez la supprimer, contactez un administrateur de Eve)*.'
                        )
                    ),
                ],
            });
            return;
        }
        logger.error(`Erreur lors de l'ajout de la question : ${error}`);
        await interaction.editReply({
            embeds: [errorEmbed(interaction, new Error("Une erreur est survenue lors de l'ajout de la question."))],
        });
        return;
    }

    await interaction.editReply({ embeds: [successEmbed(interaction, 'Question ajoutée avec succès !')] });
}
