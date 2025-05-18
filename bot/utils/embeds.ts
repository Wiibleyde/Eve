import { EmbedBuilder } from "discord.js";
import { client } from "../bot";

export function basicEmbedGenerator(): EmbedBuilder {
    return new EmbedBuilder()
        .setFooter({
            text: "Eve – Toujours prête à vous aider.",
            iconURL: client.user?.displayAvatarURL() || "",
        })
        .setColor("#FFFFFF")
        .setTimestamp()
}

export function errorEmbedGenerator(error: string): EmbedBuilder {
    return new EmbedBuilder()
        .setTitle("Oops... Une erreur est survenue !")
        .setDescription(error)
        .setFooter({
            text: "Eve – Toujours prête à vous aider.",
            iconURL: client.user?.displayAvatarURL() || "",
        })
        .setColor("#FF0000")
        .setTimestamp()
}

export function successEmbedGenerator(success: string): EmbedBuilder {
    return new EmbedBuilder()
        .setTitle("Succès !")
        .setDescription(success)
        .setFooter({
            text: "Eve – Toujours prête à vous aider.",
            iconURL: client.user?.displayAvatarURL() || "",
        })
        .setColor("#00FF00")
        .setTimestamp()
}

export function infoEmbedGenerator(info: string): EmbedBuilder {
    return new EmbedBuilder()
        .setTitle("Info")
        .setDescription(info)
        .setFooter({
            text: "Eve – Toujours prête à vous aider.",
            iconURL: client.user?.displayAvatarURL() || "",
        })
        .setColor("#0000FF")
        .setTimestamp()
}

export function warningEmbedGenerator(warning: string): EmbedBuilder {
    return new EmbedBuilder()
        .setTitle("Avertissement")
        .setDescription(warning)
        .setFooter({
            text: "Eve – Toujours prête à vous aider.",
            iconURL: client.user?.displayAvatarURL() || "",
        })
        .setColor("#FFFF00")
        .setTimestamp()
}