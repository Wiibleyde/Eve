import { config } from '@/config';
import BlaguesAPI from 'blagues-api';
import { Category } from 'blagues-api/dist/types/types';
import {
    ColorResolvable,
    CommandInteraction,
    EmbedBuilder,
    SlashCommandBuilder,
    SlashCommandOptionsOnlyBuilder,
} from 'discord.js';

const blagues = new BlaguesAPI(config.BLAGUE_API_TOKEN);

export const data: SlashCommandOptionsOnlyBuilder = new SlashCommandBuilder()
    .setName('blague')
    .setDescription('Affiche une blague (Eve et ses développeurs ne sont pas responsables des blagues affichées)')
    .addStringOption((option) =>
        option
            .setName('type')
            .setDescription('Le type de blague à afficher')
            .setRequired(false)
            .addChoices([
                {
                    name: 'Générale',
                    value: blagues.categories.GLOBAL,
                },
                {
                    name: 'Développeur',
                    value: blagues.categories.DEV,
                },
                /* {
                    name: "Humour noir",
                    value: blagues.categories.DARK
                }, */
                /* {
                    name: "Limite limite",
                    value: blagues.categories.LIMIT
                }, */
                {
                    name: 'Beauf',
                    value: blagues.categories.BEAUF,
                },
                /* {
                    name: "Blondes",
                    value: blagues.categories.BLONDES
                } */
            ])
    );

function randomEmbedColor(): ColorResolvable {
    return Math.floor(Math.random() * 16777215);
}

export async function execute(interaction: CommandInteraction): Promise<void> {
    await interaction.deferReply();
    const type = interaction.options.get('type')?.value as Category;
    const joke = await blagues.randomCategorized(type);

    const embed = new EmbedBuilder()
        .setTitle('Blague')
        .setDescription(joke.joke)
        .addFields({
            name: 'Réponse',
            value: joke.answer,
        })
        .setColor(randomEmbedColor())
        .setFooter({
            text: `Eve et ses développeurs ne sont pas responsables des blagues affichées.`,
            iconURL: interaction.client.user?.displayAvatarURL(),
        })
        .setTimestamp();

    await interaction.editReply({ embeds: [embed] });
}
