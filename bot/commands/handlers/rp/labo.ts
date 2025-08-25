import type { ICommand } from "@bot/commands/command";
import { config } from "@utils/core/config";
import { addLaboInQuery, laboInQueryManager, type LaboInQueryEntry } from "@utils/rp/labo";
import { lsmsEmbedGenerator, lsmsErrorEmbedGenerator, lsmsSuccessEmbedGenerator } from "@utils/rp/lsms";
import { ChatInputCommandInteraction, InteractionContextType, MessageFlags, SlashCommandBuilder, TextChannel } from "discord.js";

const bloodTypePercentages: { [key: string]: number } = {
    "O+": 34,
    "A+": 28,
    "B+": 20,
    "AB+": 2,
    "O-": 7,
    "A-": 6,
    "B-": 2,
    "AB-": 1,
};

const alcoholeLevels: { [key: string]: number } = {
    "none": 80,
    "low": 15,
    "medium": 4,
    "high": 1,
};

const drugTypes: { [key: string]: number } = {
    "negative": 70,
    "positive": 30,
};

export const labo: ICommand = {
    data: new SlashCommandBuilder()
        .setName("labo")
        .setDescription("[LSMS] Demander une analyse sanguine au laboratoire")
        .addSubcommand(sub =>
            sub.setName("bloodgroup")
                .setDescription("Analyse du groupe sanguin")
                .addStringOption(option =>
                    option.setName("nom_prenom")
                        .setDescription("Nom et prénom de la personne à tester")
                        .setRequired(true)
                )
                .addStringOption(option =>
                    option.setName("resultat")
                        .setDescription("Résultat truqué du test (optionnel)")
                        .addChoices(
                            { name: "A+", value: "A+" },
                            { name: "A-", value: "A-" },
                            { name: "B+", value: "B+" },
                            { name: "B-", value: "B-" },
                            { name: "AB+", value: "AB+" },
                            { name: "AB-", value: "AB-" },
                            { name: "O+", value: "O+" },
                            { name: "O-", value: "O-" }
                        )
                        .setRequired(false)
                )
                .addStringOption(option =>
                    option.setName("time")
                        .setDescription("Temps de traitement du test (en minutes)")
                        .setRequired(false)
                        .setMinLength(1)
                        .setMaxLength(3)
                )
        )
        .addSubcommand(sub =>
            sub.setName("alcohole")
                .setDescription("Analyse de l'alcoolémie")
                .addStringOption(option =>
                    option.setName("nom_prenom")
                        .setDescription("Nom et prénom de la personne à tester")
                        .setRequired(true)
                )
                .addStringOption(option =>
                    option.setName("resultat")
                        .setDescription("Résultat truqué du test (optionnel)")
                        .addChoices(
                            { name: "Pas d'alcool détecté (0g/L)", value: "none" },
                            { name: "Faible taux d'alcool (0.1 - 0.5g/L)", value: "low" },
                            { name: "Taux d'alcool moyen (0.5 - 1.5g/L)", value: "medium" },
                            { name: "Haut taux d'alcool (>1.5g/L)", value: "high" }
                        )
                        .setRequired(false)
                )
                .addStringOption(option =>
                    option.setName("time")
                        .setDescription("Temps de traitement du test (en minutes)")
                        .setRequired(false)
                        .setMinLength(1)
                        .setMaxLength(3)
                )
        )
        .addSubcommand(sub =>
            sub.setName("drugs")
                .setDescription("Analyse de dépistage de drogues")
                .addStringOption(option =>
                    option.setName("nom_prenom")
                        .setDescription("Nom et prénom de la personne à tester")
                        .setRequired(true)
                )
                .addStringOption(option =>
                    option.setName("depistage")
                        .setDescription("Type de dépistage")
                        .setRequired(true)
                        .addChoices(
                            { name: "Méthamphétamine", value: "methamphetamine" },
                            { name: "Cocaïne", value: "cocaine" },
                            { name: "Héroïne", value: "heroin" },
                            { name: "Opium", value: "opium" },
                            { name: "LSD", value: "lsd" },
                            { name: "Champignons hallucinogènes", value: "hallucinogenicmushrooms" },
                            { name: "Ecstasy", value: "ecstasy" },
                            { name: "Amphétamines", value: "amphetamines" },
                            { name: "Cannabis", value: "cannabis" }
                        )
                )
                .addStringOption(option =>
                    option.setName("resultat")
                        .setDescription("Résultat truqué du test (optionnel)")
                        .addChoices(
                            { name: "Négatif", value: "negative" },
                            { name: "Positif", value: "positive" }
                        )
                        .setRequired(false)
                )
                .addStringOption(option =>
                    option.setName("time")
                        .setDescription("Temps de traitement du test (en minutes)")
                        .setRequired(false)
                        .setMinLength(1)
                        .setMaxLength(3)
                )
        )
        .addSubcommand(sub =>
            sub.setName("diseases")
                .setDescription("Analyse de maladies")
                .addStringOption(option =>
                    option.setName("nom_prenom")
                        .setDescription("Nom et prénom de la personne à tester")
                        .setRequired(true)
                )
                .addStringOption(option =>
                    option.setName("resultat")
                        .setDescription("Résultat truqué du test (optionnel)")
                        .setRequired(false)
                )
                .addStringOption(option =>
                    option.setName("time")
                        .setDescription("Temps de traitement du test (en minutes)")
                        .setRequired(false)
                        .setMinLength(1)
                        .setMaxLength(3)
                )
        )
        .setContexts([InteractionContextType.Guild, InteractionContextType.PrivateChannel]),
    guildIds: ['872119977946263632', config.EVE_HOME_GUILD],
    execute: async (interaction: ChatInputCommandInteraction) => {
        await interaction.deferReply({
            flags: [MessageFlags.Ephemeral],
        });
        const subcommand = interaction.options.getSubcommand();
        const name = interaction.options.getString("nom_prenom", true);
        let result = interaction.options.getString("resultat", false) || undefined;
        const timeStr = interaction.options.getString("time", false) || undefined;
        console.log(`Labo: ${interaction.user.tag} requested a ${subcommand} analysis for ${name} with result ${result || "random"} and time ${timeStr || "default"}.`);
        let time: number;
        if (timeStr) {
            time = parseInt(timeStr, 10);
            // Calcul du prochain reboot à 6h du matin
            const now = new Date();
            const nextReboot = new Date(now);
            nextReboot.setHours(6, 0, 0, 0);
            if (now.getHours() >= 6) {
                nextReboot.setDate(now.getDate() + 1);
            }
            const timeBeforeReboot = Math.floor((nextReboot.getTime() - now.getTime()) / 60000);
            if (isNaN(time) || time < 1 || time > 999 || time > timeBeforeReboot) {
                await interaction.editReply({
                    embeds: [lsmsErrorEmbedGenerator("Le temps doit être un nombre entre 1 et 999, et inférieur au temps avant le redémarrage du bot (6h du matin).")],
                });
                return;
            }
        } else {
            switch (subcommand) {
                case "bloodgroup":
                    time = 5;
                    break;
                case "alcohole":
                    time = 3;
                    break;
                case "drugs":
                    time = 10;
                    break;
                case "diseases":
                    time = 15;
                    break;
                default:
                    time = 5;
            }
        }

        // Précalcule le résultat si non fourni
        if (!result) {
            if (subcommand === "bloodgroup") {
                // Tirage pondéré selon bloodTypePercentages
                const pool: string[] = [];
                Object.entries(bloodTypePercentages).forEach(([type, percent]) => {
                    for (let i = 0; i < percent; i++) pool.push(type);
                });
                result = pool[Math.floor(Math.random() * pool.length)];
            } else if (subcommand === "alcohole") {
                // Tirage pondéré selon alcoholeLevels
                const pool: string[] = [];
                Object.entries(alcoholeLevels).forEach(([level, percent]) => {
                    for (let i = 0; i < percent; i++) pool.push(level);
                });
                result = pool[Math.floor(Math.random() * pool.length)];
            } else if (subcommand === "drugs") {
                // Tirage pondéré selon drugTypes
                const pool: string[] = [];
                Object.entries(drugTypes).forEach(([type, percent]) => {
                    for (let i = 0; i < percent; i++) pool.push(type);
                });
                result = pool[Math.floor(Math.random() * pool.length)];
            } else if (subcommand === "diseases") {
                result = "Négatif";
            }
        }

        const waitingEmbed = lsmsEmbedGenerator()
            .setTitle("Analyse en cours")
            .setDescription(`Analyse de type **${subcommand}** pour **${name}** en cours. Résultat disponible dans environ **${time}** minute(s).`)
            .addFields(
                { name: "Type d'analyse", value: laboInQueryManager.getAnalyseType({ type: subcommand } as LaboInQueryEntry), inline: true },
                { name: "Nom de la personne", value: name, inline: true },
                { name: "Demandé par", value: `<@${interaction.user.id}>`, inline: true }
            );
        const message = await (interaction.channel as TextChannel)?.send({ embeds: [waitingEmbed] });
        if (!message) {
            await interaction.editReply({
                embeds: [lsmsErrorEmbedGenerator("Impossible d'envoyer le message dans ce canal.")],
            });
            return;
        }

        const entry: LaboInQueryEntry = {
            channeId: interaction.channelId,
            messageId: message.id,
            userId: interaction.user.id,
            startTime: new Date(),
            name,
            type: subcommand,
            result,
            time
        };
        if (subcommand === "drugs") {
            entry.type = interaction.options.getString("depistage", true); // type = dépistage pour drugs
        }
        addLaboInQuery(entry);
        await interaction.editReply({
            embeds: [lsmsSuccessEmbedGenerator("Demande d'analyse reçue. Résultat disponible dans environ " + time + " minute(s).")],
        });
        console.log(`Labo: New ${subcommand} analysis for ${name}, requested by ${interaction.user.tag}.`);
    }
};