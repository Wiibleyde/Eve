import { ChatInputCommandInteraction, MessageFlags, SlashCommandBuilder } from 'discord.js';
import type { ICommand } from '../command';
import { basicEmbedGenerator, errorEmbedGenerator, successEmbedGenerator } from '../../utils/embeds';
import { prisma } from '../../../utils/core/database';
import { client } from '../../bot';

const months = {
    '01': 'Janvier',
    '02': 'Février',
    '03': 'Mars',
    '04': 'Avril',
    '05': 'Mai',
    '06': 'Juin',
    '07': 'Juillet',
    '08': 'Août',
    '09': 'Septembre',
    '10': 'Octobre',
    '11': 'Novembre',
    '12': 'Décembre',
};

export const birthday: ICommand = {
    data: new SlashCommandBuilder()
        .setName('birthday')
        .setDescription('Commandes liées aux anniversaires')
        .addSubcommand((subcommand) =>
            subcommand
                .setName('set')
                .setDescription('Définir votre anniversaire')
                .addStringOption((option) =>
                    option.setName('date').setDescription('Date de votre anniversaire (JJ/MM)').setRequired(true)
                )
        )
        .addSubcommand((subcommand) =>
            subcommand.setName('get').setDescription('Obtenir la date de votre anniversaire')
        )
        .addSubcommand((subcommand) => subcommand.setName('remove').setDescription('Supprimer votre anniversaire'))
        .addSubcommand((subcommand) =>
            subcommand.setName('list').setDescription('Lister les anniversaires de la communauté')
        ),
    execute: async (interaction) => {
        await interaction.deferReply({ flags: [MessageFlags.Ephemeral] });
        const subcommand = (interaction as ChatInputCommandInteraction).options.getSubcommand();
        switch (subcommand) {
            case 'set': {
                const dateArg = interaction.options.get('date')?.value as string;
                const dateParts = dateArg.split('/');
                if (dateParts.length !== 3) {
                    const errorEmbed = errorBirthdayEmbedGenerator(
                        "La date d'anniversaire doit être au format JJ/MM/AAAA"
                    );
                    await interaction.editReply({ embeds: [errorEmbed] });
                    return;
                }
                const day = dateParts[0] || '';
                const month = dateParts[1] || '';
                const year = dateParts[2] || '';
                const date = new Date(`${year}-${month}-${day}`);
                if (isNaN(date.getTime())) {
                    const errorEmbed = errorBirthdayEmbedGenerator(
                        "La date d'anniversaire doit être au format JJ/MM/AAAA"
                    );
                    await interaction.editReply({ embeds: [errorEmbed] });
                    return;
                }
                if (date.getFullYear() < 1900 || date.getFullYear() > new Date().getFullYear()) {
                    const errorEmbed = errorBirthdayEmbedGenerator(
                        "L'année de votre anniversaire doit être comprise entre 1900 et l'année actuelle"
                    );
                    await interaction.editReply({ embeds: [errorEmbed] });
                    return;
                }
                if (date.getDate() !== parseInt(day) || date.getMonth() + 1 !== parseInt(month)) {
                    const errorEmbed = errorBirthdayEmbedGenerator(
                        "La date d'anniversaire doit être au format JJ/MM/AAAA"
                    );
                    await interaction.editReply({ embeds: [errorEmbed] });
                    return;
                }
                await prisma.globalUserData.upsert({
                    where: {
                        userId: interaction.user.id,
                    },
                    create: {
                        userId: interaction.user.id,
                        birthDate: date,
                    },
                    update: {
                        birthDate: date,
                    },
                });
                const successEmbed = successBirthdayEmbedGenerator(
                    "Votre date d'anniversaire a été définie avec succès"
                );
                await interaction.editReply({ embeds: [successEmbed] });
                break;
            }
            case 'get': {
                const userBirthday = await prisma.globalUserData.findFirst({
                    where: {
                        userId: interaction.user.id,
                        birthDate: {
                            not: null,
                        },
                    },
                    select: {
                        birthDate: true,
                    },
                });
                if (!userBirthday) {
                    const errorEmbed = errorBirthdayEmbedGenerator("Vous n'avez pas défini de date d'anniversaire");
                    await interaction.editReply({ embeds: [errorEmbed] });
                    return;
                }

                if (!userBirthday.birthDate) {
                    const errorEmbed = errorBirthdayEmbedGenerator(
                        "Erreur lors de la récupération de votre date d'anniversaire"
                    );
                    await interaction.editReply({ embeds: [errorEmbed] });
                    return;
                }

                const birthdayDate = new Date(userBirthday.birthDate);
                const day = birthdayDate.getDate().toString().padStart(2, '0');
                const month = (birthdayDate.getMonth() + 1).toString().padStart(2, '0');
                const formattedDate = `${day} ${months[month as keyof typeof months]}`;

                // Calculate next birthday timestamp
                const now = new Date();
                const nextBirthday = new Date(now.getFullYear(), birthdayDate.getMonth(), birthdayDate.getDate());
                if (nextBirthday < now) {
                    nextBirthday.setFullYear(now.getFullYear() + 1);
                }
                const timestamp = Math.floor(nextBirthday.getTime() / 1000);

                const birthdayEmbed = birthdayEmbedGenerator();
                birthdayEmbed.setTitle("Votre date d'anniversaire");
                birthdayEmbed.setDescription(`Votre date d'anniversaire est le ${formattedDate} (<t:${timestamp}:R>)`);
                await interaction.editReply({ embeds: [birthdayEmbed] });
                break;
            }
            case 'remove': {
                const userBirthdayToRemove = await prisma.globalUserData.findFirst({
                    where: {
                        userId: interaction.user.id,
                        birthDate: {
                            not: null,
                        },
                    },
                    select: {
                        birthDate: true,
                    },
                });
                if (!userBirthdayToRemove) {
                    const errorEmbed = errorBirthdayEmbedGenerator("Vous n'avez pas défini de date d'anniversaire");
                    await interaction.editReply({ embeds: [errorEmbed] });
                    return;
                }
                if (!userBirthdayToRemove.birthDate) {
                    const errorEmbed = errorBirthdayEmbedGenerator(
                        "Erreur lors de la récupération de votre date d'anniversaire"
                    );
                    await interaction.editReply({ embeds: [errorEmbed] });
                    return;
                }
                await prisma.globalUserData.update({
                    where: {
                        userId: interaction.user.id,
                    },
                    data: {
                        birthDate: null,
                    },
                });
                const successEmbed = successBirthdayEmbedGenerator(
                    "Votre date d'anniversaire a été supprimée avec succès"
                );
                await interaction.editReply({ embeds: [successEmbed] });
                break;
            }
            case 'list': {
                if (!interaction.guildId) {
                    await interaction.editReply({ content: 'Impossible de récupérer les anniversaires' });
                    return;
                }
                const usersOnGuild = await client.guilds
                    .fetch(interaction.guildId)
                    .then((guild) => guild.members.fetch());
                const userIds = usersOnGuild.map((user) => user.id);
                if (!userIds) {
                    await interaction.editReply({ content: 'Impossible de récupérer les anniversaires' });
                    return;
                }
                const birthdays = await prisma.globalUserData.findMany({
                    where: {
                        userId: {
                            in: userIds,
                        },
                        birthDate: {
                            not: null,
                        },
                    },
                });
                if (birthdays.length === 0) {
                    await interaction.editReply({ content: 'Aucun anniversaire enregistré' });
                    return;
                }
                birthdays.sort((a, b) => {
                    if (a.birthDate && b.birthDate) {
                        return a.birthDate.getMonth() - b.birthDate.getMonth();
                    }
                    return 0;
                });
                const monthsBirthdays: { [key: string]: string[] } = {};
                for (const month in months) {
                    for (const birthday of birthdays) {
                        if (birthday.birthDate?.getMonth() === parseInt(month as string) - 1) {
                            if (!monthsBirthdays[month]) {
                                monthsBirthdays[month] = [];
                            }
                            monthsBirthdays[month].push(
                                `<@${birthday.userId}> - ${birthday.birthDate.toLocaleDateString('fr-FR')}`
                            );
                        }
                    }
                }

                const birthdayEmbed = birthdayEmbedGenerator();
                birthdayEmbed.setTitle('Anniversaires de la communauté');
                birthdayEmbed.setDescription('Voici la liste des anniversaires de la communauté, triée par mois :');
                for (const month in monthsBirthdays) {
                    const monthBirthdays = monthsBirthdays[month];
                    if (monthBirthdays) {
                        birthdayEmbed.addFields({
                            name: months[month as keyof typeof months],
                            value: monthBirthdays.join('\n'),
                            inline: true,
                        });
                    }
                }

                await interaction.editReply({ embeds: [birthdayEmbed] });
                return;
            }
            default: {
                const errorEmbed = errorBirthdayEmbedGenerator("La commande demandée n'existe pas");
                await interaction.editReply({ embeds: [errorEmbed] });
                return;
            }
        }
    },
};

export function birthdayEmbedGenerator() {
    return basicEmbedGenerator().setAuthor({
        name: 'Eve - Anniversaires',
        iconURL:
            'https://cdn.discordapp.com/attachments/1373968524229218365/1373971343162343465/image.png?ex=682c5a07&is=682b0887&hm=db9c2f437f8475ba28510e252284c554f3f8d93864037b6ed03a58adbfa005e0&',
    });
}

function successBirthdayEmbedGenerator(reason: string) {
    return successEmbedGenerator(reason).setAuthor({
        name: 'Eve - Anniversaires',
        iconURL:
            'https://cdn.discordapp.com/attachments/1373968524229218365/1373971343162343465/image.png?ex=682c5a07&is=682b0887&hm=db9c2f437f8475ba28510e252284c554f3f8d93864037b6ed03a58adbfa005e0&',
    });
}

function errorBirthdayEmbedGenerator(reason: string) {
    return errorEmbedGenerator(reason).setAuthor({
        name: 'Eve - Anniversaires',
        iconURL:
            'https://cdn.discordapp.com/attachments/1373968524229218365/1373971343162343465/image.png?ex=682c5a07&is=682b0887&hm=db9c2f437f8475ba28510e252284c554f3f8d93864037b6ed03a58adbfa005e0&',
    });
}
