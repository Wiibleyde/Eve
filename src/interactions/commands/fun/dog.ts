import { errorEmbed } from '@/utils/embeds';
import { CommandInteraction, SlashCommandBuilder } from 'discord.js';

const dogImgUrl = 'https://dog.ceo/api/breeds/image/random';

export const data: SlashCommandBuilder = new SlashCommandBuilder()
    .setName('dog')
    .setDescription('Affiche une image de chien');

/**
 * Executes the command to fetch and send a random dog image.
 *
 * @param interaction - The command interaction object.
 * @returns A promise that resolves when the interaction reply is sent.
 *
 * Fetches a random dog image from the specified URL and sends it as a reply to the interaction.
 * If an error occurs during the fetch or if the image URL is not found, an error message is sent instead.
 */
export async function execute(interaction: CommandInteraction): Promise<void> {
    const response = await fetch(dogImgUrl);
    const data = await response.json();
    const dogImg = data.message;

    if (!dogImg) {
        await interaction.reply({
            embeds: [
                errorEmbed(
                    interaction,
                    new Error("Une erreur est survenue lors de la récupération de l'image de dog.")
                ),
            ],
            ephemeral: true,
        });
        return;
    }

    await interaction.reply({ files: [dogImg], ephemeral: true });
}
