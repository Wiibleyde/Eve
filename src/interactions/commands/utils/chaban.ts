import { logger } from '@/index';
import { CommandInteraction, EmbedBuilder, MessageFlags, SlashCommandBuilder } from 'discord.js';

const url = `https://opendata.bordeaux-metropole.fr/api/explore/v2.1/catalog/datasets/previsions_pont_chaban/exports/json?lang=fr&timezone=Europe%2FBerlin`;

export const data: SlashCommandBuilder = new SlashCommandBuilder()
    .setName('chaban')
    .setDescription('Informations sur le pont Chaban-Delmas');

/**
 * Executes the command to check the current status of the Pont Chaban-Delmas.
 * It fetches the latest closure data, determines if the bridge is currently closed,
 * and constructs an appropriate response embed to send back to the user.
 *
 * @param {CommandInteraction} interaction - The interaction object representing the command invocation.
 * @returns {Promise<void>} A promise that resolves when the reply has been edited.
 */
export async function execute(interaction: CommandInteraction) {
    await interaction.deferReply({ withResponse: true, flags: [MessageFlags.Ephemeral] });
    const data = await downloadJsonFile(url);
    const sortedData = data
        .sort(
            (a: { date_passage: string | number | Date }, b: { date_passage: string | number | Date }) =>
                new Date(a.date_passage).getTime() - new Date(b.date_passage).getTime()
        )
        .reverse();
    const record = sortedData[0];

    const [endHour, endMinute] = record.re_ouverture_a_la_circulation.split(':');
    const [startHour, startMinute] = record.fermeture_a_la_circulation.split(':');

    const closureStart = new Date(record.date_passage);
    closureStart.setHours(parseInt(startHour), parseInt(startMinute), 0, 0);

    const closureEnd = new Date(record.date_passage);
    closureEnd.setHours(parseInt(endHour), parseInt(endMinute), 0, 0);
    if (closureEnd < closureStart) {
        closureEnd.setDate(closureEnd.getDate() + 1);
    }

    const now = new Date();
    const isClosedNow = now >= closureStart && now <= closureEnd;
    const isNextClosureToday = closureStart.getDate() === now.getDate() && closureStart > now;

    const embed = new EmbedBuilder()
        .setTitle('Pont Chaban-Delmas')
        .setDescription(`Le pont est actuellement ${isClosedNow ? 'fermé' : 'ouvert'}.`);

    if (isClosedNow) {
        embed.setColor('Red');
        embed.addFields({
            name: 'Ouverture',
            value: `<t:${Math.floor(closureEnd.getTime() / 1000)}:R>`,
        });
    } else if (isNextClosureToday) {
        embed.setColor('Yellow');
        embed.addFields({
            name: 'Prochaine fermeture',
            value: `<t:${Math.floor(closureStart.getTime() / 1000)}:R>`,
        });
    } else {
        embed.setColor('Green');
        if (closureStart.getTime() > now.getTime()) {
            embed.addFields({
                name: 'Prochaine fermeture',
                value: `<t:${Math.floor(closureStart.getTime() / 1000)}:R>`,
            });
        } else {
            embed.addFields({
                name: 'Prochaine fermeture',
                value: 'Aucune fermeture prévue.',
            });
        }
    }

    await interaction.editReply({ embeds: [embed] });
}

/**
 * Represents the data related to the Chaban bridge's passage and traffic status.
 */
interface ChabanData {
    date_passage: string;
    re_ouverture_a_la_circulation: string;
    fermeture_a_la_circulation: string;
}

/**
 * Downloads a JSON file from the specified URL and returns the parsed data as an array of ChabanData objects.
 *
 * @param url - The URL of the JSON file to download.
 * @returns A promise that resolves to an array of ChabanData objects.
 * @throws Will log an error and return an empty array if the fetch operation fails.
 */
async function downloadJsonFile(url: string): Promise<ChabanData[]> {
    try {
        const response = await fetch(url);
        const data: ChabanData[] = await response.json();
        return data;
    } catch (error) {
        logger.error(`Failed to fetch Chaban-Delmas data: ${error}`);
        return [];
    }
}
