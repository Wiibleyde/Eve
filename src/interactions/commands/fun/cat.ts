import { errorEmbed } from '@/utils/embeds';
import { CommandInteraction, SlashCommandBuilder } from 'discord.js';

const catImgUrl = 'https://api.thecatapi.com/v1/images/search';

export const data: SlashCommandBuilder = new SlashCommandBuilder()
    .setName('cat')
    .setDescription('Affiche une image de chat');

/**
 * Executes the command to fetch and send a random cat image.
 *
 * @param interaction - The command interaction object.
 * @returns A promise that resolves when the interaction reply is sent.
 *
 * Fetches a random cat image from the specified URL and sends it as a reply to the interaction.
 * If an error occurs during the fetch or if the image URL is not found, an error message is sent instead.
 */
export async function execute(interaction: CommandInteraction): Promise<void> {
    const response = await fetch(catImgUrl);
    const data = await response.json();
    const catImg = data[0].url;

    if (!catImg) {
        await interaction.reply({
            embeds: [
                errorEmbed(
                    interaction,
                    new Error("Une erreur est survenue lors de la récupération de l'image de chat.")
                ),
            ],
            ephemeral: true,
        });
        return;
    }

    await interaction.reply({ files: [catImg], ephemeral: true });
}
