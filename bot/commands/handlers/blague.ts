import { ActionRowBuilder, ButtonBuilder, ButtonStyle, MessageFlags, SlashCommandBuilder } from "discord.js";
import type { ICommand } from "../command";
import BlaguesAPI from 'blagues-api';
import { config } from "../../../utils/config";
import type { Category } from "blagues-api/dist/types/types";
import { basicEmbedGenerator, errorEmbedGenerator, successEmbedGenerator } from "../../utils/embeds";

const blagues = new BlaguesAPI(config.BLAGUE_API_TOKEN);


export const blague: ICommand = {
    data: new SlashCommandBuilder()
        .setName("blague")
        .setDescription("Fait une blague")
        .addStringOption((option) =>
        option
            .setName('type')
            .setDescription('Le type de blague à afficher')
            .setRequired(true)
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
    ),
    execute: async (interaction) => {
        await interaction.deferReply({ withResponse: true, flags: [MessageFlags.Ephemeral] });
        const type = interaction.options.get('type')?.value as Category;
        const blague = await blagues.randomCategorized(type);
        const embed = jokeBasicEmbedGenerator();
        embed.setDescription(blague.joke);
        embed.addFields({
            name: "Réponse :",
            value: `||${blague.answer}||`,
        });

        const button = new ButtonBuilder().setCustomId('jokeSetPublicButton').setLabel('Rendre publique').setStyle(ButtonStyle.Primary)
        const actionRow = new ActionRowBuilder<ButtonBuilder>().addComponents(button);

        await interaction.editReply({ embeds: [embed], components: [actionRow] });
    },
};

export function jokeBasicEmbedGenerator() {
    const embed = basicEmbedGenerator();
    const actualFooter = embed.data.footer;
    embed.setTitle("Voici une blague :");
    embed.setFooter({
        text: "⚠️ Eve et ses développeurs ne sont pas responsable des blagues proposées. ⚠️",
        iconURL: actualFooter?.icon_url,
    });
    embed.setAuthor({
        name: "Eve - Blague",
        iconURL: "https://cdn.discordapp.com/attachments/1373968524229218365/1374021216360206388/latest.png?ex=682c887a&is=682b36fa&hm=141ddd2a49ccc4667737e5d19fd5158cfb76c574f98aaa6ff17e29860067f5e0&",
    });
    return embed;
}

export function jokeErrorEmbedGenerator(reason: string) {
    const embed = errorEmbedGenerator(reason);
    const actualFooter = embed.data.footer;
    embed.setFooter({
        text: "⚠️ Eve et ses développeurs ne sont pas responsable des blagues proposées. ⚠️",
        iconURL: actualFooter?.icon_url,
    });
    embed.setAuthor({
        name: "Eve - Blague",
        iconURL: "https://cdn.discordapp.com/attachments/1373968524229218365/1374021216360206388/latest.png?ex=682c887a&is=682b36fa&hm=141ddd2a49ccc4667737e5d19fd5158cfb76c574f98aaa6ff17e29860067f5e0&",
    })
    return embed;
}

export function jokeSuccessEmbedGenerator(reason: string) {
    const embed = successEmbedGenerator(reason);
    const actualFooter = embed.data.footer;
    embed.setFooter({
        text: "⚠️ Eve et ses développeurs ne sont pas responsable des blagues proposées. ⚠️",
        iconURL: actualFooter?.icon_url,
    });
    embed.setAuthor({
        name: "Eve - Blague",
        iconURL: "https://cdn.discordapp.com/attachments/1373968524229218365/1374021216360206388/latest.png?ex=682c887a&is=682b36fa&hm=141ddd2a49ccc4667737e5d19fd5158cfb76c574f98aaa6ff17e29860067f5e0&",
    })
    return embed;
}