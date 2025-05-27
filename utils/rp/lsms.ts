import { ActionRowBuilder, ButtonBuilder, ButtonStyle, type EmbedBuilder, type User } from "discord.js";
import { basicEmbedGenerator } from "../../bot/utils/embeds";

export function lsmsEmbedGenerator() {
    return basicEmbedGenerator()
        .setAuthor({
            name: "Eve - LSMS",
            iconURL: "https://cdn.discordapp.com/icons/872119977946263632/2100e7fea046e84d057cc0009a2b240f.webp",
        });
}

export function lsmsErrorEmbedGenerator(reason: string) {
    return basicEmbedGenerator()
        .setAuthor({
            name: "Eve - LSMS",
            iconURL: "https://cdn.discordapp.com/icons/872119977946263632/2100e7fea046e84d057cc0009a2b240f.webp",
        })
        .setColor(0xff0000)
        .setTitle("Erreur")
        .setDescription(reason);
}

export function lsmsSuccessEmbedGenerator(message: string) {
    return basicEmbedGenerator()
        .setAuthor({
            name: "Eve - LSMS",
            iconURL: "https://cdn.discordapp.com/icons/872119977946263632/2100e7fea046e84d057cc0009a2b240f.webp",
        })
        .setColor(0x00ff00)
        .setTitle("Succès")
        .setDescription(message);
}


export function lsmsDutyEmbedGenerator(onDutyPeople: User[], onCallPeople: User[]): { embed: EmbedBuilder, actionRow: ActionRowBuilder<ButtonBuilder> } {
    const dutyList = onDutyPeople.length > 0 ? onDutyPeople.map(user => `<@${user.id}>`).join("\n") : "Personne n'est en service :(";
    const callList = onCallPeople.length > 0 ? onCallPeople.map(user => `<@${user.id}>`).join("\n") : "Personne n'est en astreinte :(";
    const embed = lsmsEmbedGenerator()
        .setTitle("Gestionnaire de service")
        .setDescription("Cliquez sur les boutons ci-dessous pour gérer les services.")
        .addFields({
            name: "En service :",
            value: dutyList,
            inline: true
        }, {
            name: "En astreinte :",
            value: callList,
            inline: true
        });
    const buttons = [
        new ButtonBuilder().setCustomId("handleLsmsDuty").setLabel("Prendre/Quitter le service").setStyle(ButtonStyle.Primary),
        new ButtonBuilder().setCustomId("handleLsmsOnCall").setLabel("Prendre/Quitter l'astreinte").setStyle(ButtonStyle.Secondary),
    ];
    const actionRow = new ActionRowBuilder<ButtonBuilder>().addComponents(buttons);

    return {
        embed,
        actionRow,
    }
}