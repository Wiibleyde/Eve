import { config } from '@/config';
import { prisma } from '@/utils/database';
import { errorEmbed, successEmbed } from '@/utils/embeds';
import { backSpace } from '@/utils/textUtils';
import { EmbedBuilder, MessageFlags, ModalSubmitInteraction, TextChannel } from 'discord.js';

export async function reportQuestionModal(interaction: ModalSubmitInteraction) {
    const user = interaction.user;
    const reportReason = interaction.fields.getTextInputValue('reportReason');
    const uuid = interaction.customId.split('--')[1];
    const question = await prisma.quizQuestions.findFirst({
        where: {
            uuid: uuid,
        },
        include: {
            author: true,
        },
    });
    if (!question) {
        await interaction.reply({
            embeds: [errorEmbed(interaction, new Error('Question introuvable'))],
            ephemeral: true,
        });
        return;
    }

    const channel = (await interaction.client.channels.fetch(config.REPORT_CHANNEL)) as TextChannel;
    if (!channel) {
        await interaction.reply({
            embeds: [errorEmbed(interaction, new Error('Channel de report introuvable'))],
            ephemeral: true,
        });
        return;
    }

    const embed = new EmbedBuilder()
        .setTitle('Signalement de question de quiz')
        .setDescription(`UUID: ${uuid}${backSpace}`)
        .setColor(0xff0000)
        .setTimestamp()
        .setFooter({ text: `Eve – Toujours prête à vous aider.`, iconURL: interaction.client.user.displayAvatarURL() });

    if (reportReason) {
        embed.addFields({
            name: 'Raison du signalement',
            value: reportReason,
            inline: false,
        });
        embed.addFields({
            name: 'Question',
            value: '```' + question.question + '```',
            inline: false,
        });
        embed.addFields({
            name: 'Réponses',
            value: `Bonne réponse: ${'`'}${question.answer}${'`'}${backSpace}Mauvaises réponses: ${'`'}${question.badAnswer1}${'`'}, ${'`'}${question.badAnswer2}${'`'}, ${'`'}${question.badAnswer3}${'`'}`,
            inline: false,
        });
        if (question.author) {
            embed.addFields({
                name: 'Auteur',
                value: `<@${question.author.userId}>`,
                inline: false,
            });
        } else {
            embed.addFields({
                name: 'Auteur',
                value: `API`,
                inline: false,
            });
        }
        embed.addFields({
            name: 'Catégorie',
            value: question.category,
            inline: true,
        });
        embed.addFields({
            name: 'Difficulté',
            value: question.difficulty,
            inline: true,
        });
    }

    await channel.send({ content: `<@${config.OWNER_ID}>, report de : <@${user.id}>`, embeds: [embed] });
    await interaction.reply({
        embeds: [successEmbed(interaction, 'Question signalée')],
        flags: [MessageFlags.Ephemeral],
    });
}
