import {
    ActionRowBuilder,
    APIEmbedField,
    ButtonBuilder,
    ButtonInteraction,
    ButtonStyle,
    CommandInteraction,
    EmbedBuilder,
    ModalSubmitInteraction,
    StringSelectMenuInteraction,
} from 'discord.js';

export const radioImage = './assets/img/radio.png';

export interface RadioFrequencies {
    name: string;
    frequency: string;
}

/**
 * Formats a given service name into an internal name.
 * The internal name is converted to lowercase and spaces are replaced with underscores.
 *
 * @param serviceName - The name of the service to format.
 * @returns The formatted internal name.
 */
export function formatInternalName(serviceName: string): string {
    return serviceName.toLowerCase().replace(' ', '_');
}

/**
 * Generates an array of APIEmbedField objects from an array of RadioFrequencies.
 *
 * @param radios - An array of RadioFrequencies objects.
 * @returns An array of APIEmbedField objects with the radio name and frequency.
 */
function createFieldsForRadios(radios: RadioFrequencies[]): APIEmbedField[] {
    return radios.map((radio) => {
        return {
            name: `Radio ${radio.name}`,
            value: `Fréquence : ${radio.frequency}`,
            inline: true,
        };
    });
}

/**
 * Creates an array of ButtonBuilder instances for the given radio frequencies.
 *
 * @param {RadioFrequencies[]} radios - An array of radio frequency objects.
 * @returns {ButtonBuilder[]} An array of ButtonBuilder instances, each configured with a custom ID and label based on the radio frequency.
 */
function createButtonsForRadios(radios: RadioFrequencies[]): ButtonBuilder[] {
    const buttons = radios.map((radio) => {
        return new ButtonBuilder()
            .setCustomId(`changeRadio--${radio.name}`)
            .setLabel(`Changer la radio ${radio.name}`)
            .setStyle(ButtonStyle.Primary);
    });
    buttons.push(new ButtonBuilder().setCustomId('handleAddRadio').setLabel('+').setStyle(ButtonStyle.Success));
    buttons.push(new ButtonBuilder().setCustomId('handleRemoveRadio').setLabel('-').setStyle(ButtonStyle.Danger));
    return buttons;
}

/**
 * Creates an embed message for a radio interaction.
 *
 * @param interaction - The interaction object which can be of type CommandInteraction, ButtonInteraction, or ModalSubmitInteraction.
 * @param name - The name of the radio.
 * @param radio - An array of radio frequencies.
 * @returns An object containing the embed, action row, and optional files.
 */
export function creatEmbedForRadio(
    interaction: CommandInteraction | ButtonInteraction | ModalSubmitInteraction | StringSelectMenuInteraction,
    name: string,
    radio: RadioFrequencies[]
): { embed: EmbedBuilder; actionRow: ActionRowBuilder<ButtonBuilder>; files?: { attachment: string; name: string }[] } {
    const embed = new EmbedBuilder()
        .setTitle(`Radio du ${name}`)
        .setDescription('Voici les radios disponibles')
        .setColor('Aqua')
        .setTimestamp()
        .setThumbnail('attachment://radio.png')
        .setFooter({ text: `Eve – Toujours prête à vous aider.`, iconURL: interaction.client.user.displayAvatarURL() });

    const fields: APIEmbedField[] = createFieldsForRadios(radio);
    embed.addFields(fields);

    const buttons = createButtonsForRadios(radio);

    const actionRow = new ActionRowBuilder<ButtonBuilder>().addComponents(buttons);

    const files = [{ attachment: radioImage, name: 'radio.png' }];

    return { embed, actionRow, files };
}
