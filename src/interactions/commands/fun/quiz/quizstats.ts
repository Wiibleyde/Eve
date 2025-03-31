import { CommandInteraction, EmbedBuilder, MessageFlags, SlashCommandBuilder } from 'discord.js';
import { prisma } from '@/utils/database';

export const data: SlashCommandBuilder = new SlashCommandBuilder()
    .setName('quizstats')
    .setDescription('Affiche les statistiques du quiz');

/**
 * Executes the quiz statistics command.
 *
 * This function retrieves the user's quiz data from the database and generates an embed
 * containing the user's quiz statistics, including the number of questions answered,
 * the number of correct and incorrect answers, and the percentage of correct answers.
 * The embed color is determined based on the user's correct answer ratio.
 *
 * @param interaction - The command interaction object.
 * @returns A promise that resolves when the interaction reply is sent.
 */
export async function execute(interaction: CommandInteraction): Promise<void> {
    const userData = await prisma.globalUserData.findFirst({
        where: {
            userId: interaction.user.id,
        },
    });
    if (!userData) {
        await interaction.reply({ content: "Vous n'avez pas encore joué au quiz", flags: [MessageFlags.Ephemeral] });
        return;
    }

    const nbQuestions = userData.quizGoodAnswers + userData.quizBadAnswers;
    if (nbQuestions === 0) {
        await interaction.reply({ content: "Vous n'avez pas encore joué au quiz", flags: [MessageFlags.Ephemeral] });
        return;
    }
    const ratio = (userData.quizGoodAnswers / nbQuestions) * 100;
    const embed = new EmbedBuilder()
        .setTitle('Statistiques du quiz')
        .setDescription(`Vous avez répondu à ${nbQuestions} questions.`)
        .addFields(
            {
                name: 'Bonnes réponses',
                value: `${userData.quizGoodAnswers}`,
                inline: true,
            },
            {
                name: 'Mauvaises réponses',
                value: `${userData.quizBadAnswers}`,
                inline: true,
            },
            {
                name: 'Ratio de bonnes réponses',
                value: `${ratio.toFixed(2)}%`,
                inline: true,
            }
        )
        .setTimestamp()
        .setFooter({ text: `Eve – Toujours prête à vous aider.`, iconURL: interaction.client.user.displayAvatarURL() });
    if (ratio < 50) {
        embed.setColor('#FF0000');
    } else if (ratio < 75) {
        embed.setColor('#FFA500');
    } else {
        embed.setColor('#00FF00');
    }

    await interaction.reply({ embeds: [embed], flags: [MessageFlags.Ephemeral] });
}
