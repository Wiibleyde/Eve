import {
    CommandInteraction,
    EmbedBuilder,
    MessageFlags,
    SlashCommandBuilder,
    SlashCommandOptionsOnlyBuilder,
} from 'discord.js';
import { prisma } from '@/utils/database';
import { client } from '@/index';

export const data: SlashCommandOptionsOnlyBuilder = new SlashCommandBuilder()
    .setName('leaderboard')
    .setDescription('Affiche le classement du quiz')
    .addStringOption((option) =>
        option.setName('type').setDescription('Type de classement').setRequired(true).addChoices(
            {
                name: 'Ratio',
                value: 'ratio',
            },
            {
                name: 'Bonnes réponses',
                value: 'good',
            },
            {
                name: 'Mauvaises réponses',
                value: 'bad',
            }
        )
    );

/**
 * Executes the leaderboard command, which retrieves and displays the top 10 users
 * with the best quiz performance based on the specified type (ratio, good answers, or bad answers).
 *
 * @param {CommandInteraction} interaction - The interaction object representing the command invocation.
 * @returns {Promise<void>} A promise that resolves when the leaderboard has been sent as a reply.
 */
export async function execute(interaction: CommandInteraction): Promise<void> {
    // Get all users with quiz data
    let users = await prisma.globalUserData.findMany({
        select: {
            userId: true,
            quizGoodAnswers: true,
            quizBadAnswers: true,
        },
    });

    // Filter out users with no answers and bot itself
    users = users.filter((user) => user.quizGoodAnswers + user.quizBadAnswers > 0);
    users = users.filter((user) => user.userId !== client.user?.id);

    const type = interaction.options.get('type')?.value as string;
    switch (type) {
        case 'ratio':
            users.sort((a, b) => {
                const ratioA = a.quizGoodAnswers / (a.quizGoodAnswers + a.quizBadAnswers) || 0;
                const ratioB = b.quizGoodAnswers / (b.quizGoodAnswers + b.quizBadAnswers) || 0;
                return ratioB - ratioA;
            });
            break;
        case 'good':
            users.sort((a, b) => {
                return b.quizGoodAnswers - a.quizGoodAnswers;
            });
            break;
        case 'bad':
            users.sort((a, b) => {
                return b.quizBadAnswers - a.quizBadAnswers;
            });
            break;
    }
    users = users.slice(0, 10);

    const embed = new EmbedBuilder()
        .setTitle('Classement du quiz')
        .setDescription('Top 10 des meilleurs joueurs du quiz')
        .setTimestamp()
        .setFooter({ text: `Eve – Toujours prête à vous aider.`, iconURL: interaction.client.user.displayAvatarURL() });

    users.forEach((user, index) => {
        const total = user.quizGoodAnswers + user.quizBadAnswers;
        const ratio = total > 0 ? (user.quizGoodAnswers / total) * 100 : 0;
        
        embed.addFields({
            name: `#${index + 1} Ratio: ${ratio.toFixed(2)}%`,
            value: `<@${user.userId}> Bonnes réponses: ${user.quizGoodAnswers} | Mauvaises réponses: ${user.quizBadAnswers}`,
        });
    });

    await interaction.reply({ embeds: [embed], flags: [MessageFlags.Ephemeral] });
}
