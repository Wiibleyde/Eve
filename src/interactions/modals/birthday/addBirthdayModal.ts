import { prisma } from '@/utils/database';
import { errorEmbed, successEmbed } from '@/utils/embeds';
import { ModalSubmitInteraction } from 'discord.js';

/**
 * Handles the submission of a modal to add or update a user's birthday.
 *
 * @param interaction - The interaction object representing the modal submission.
 * @returns A promise that resolves when the interaction has been handled.
 *
 * This function performs the following steps:
 * 1. Extracts the user and birthday date text from the interaction.
 * 2. Validates the date format and parses the day, month, and year.
 * 3. Creates a Date object from the parsed values and validates it.
 * 4. Checks if the user already has a birthday record in the database.
 * 5. Updates the existing record or creates a new one with the provided birthday date.
 * 6. Sends a success or error message as a reply to the interaction.
 */
export async function addBirthdayModal(interaction: ModalSubmitInteraction): Promise<void> {
    const user = interaction.user;
    const birthdayDateText = interaction.fields.getTextInputValue('birthday');
    const dateParts = birthdayDateText.split('/');
    if (dateParts.length !== 3) {
        await interaction.reply({ embeds: [errorEmbed(interaction, new Error('Date invalide'))], ephemeral: true });
        return;
    }
    const day = parseInt(dateParts[0]) + 1;
    const month = parseInt(dateParts[1]) - 1;
    const year = parseInt(dateParts[2]);
    if (isNaN(day) || isNaN(month) || isNaN(year)) {
        await interaction.reply({ embeds: [errorEmbed(interaction, new Error('Date invalide'))], ephemeral: true });
        return;
    }
    const birthdayDate = new Date(year, month, day);
    birthdayDate.setHours(0, 0, 0, 0);
    if (isNaN(birthdayDate.getTime())) {
        await interaction.reply({ embeds: [errorEmbed(interaction, new Error('Date invalide'))], ephemeral: true });
        return;
    }
    const birthday = await prisma.globalUserData.findFirst({
        where: {
            userId: user.id,
        },
    });
    if (birthday) {
        await prisma.globalUserData.update({
            where: {
                uuid: birthday.uuid,
            },
            data: {
                birthDate: birthdayDate,
            },
        });
    } else {
        await prisma.globalUserData.create({
            data: {
                birthDate: birthdayDate,
                userId: user.id,
            },
        });
    }
    await interaction.reply({ embeds: [successEmbed(interaction, 'Anniversaire ajouté')], ephemeral: true });
}
