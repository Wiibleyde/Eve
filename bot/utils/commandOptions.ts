import type { 
    ChatInputCommandInteraction, 
    User, 
    Role, 
    Attachment, 
    GuildBasedChannel 
} from 'discord.js';

/**
 * Utility functions to safely extract options from ChatInputCommandInteraction
 * These functions provide type safety and eliminate the need for casting
 */

/**
 * Get a string option value
 */
export function getStringOption(interaction: ChatInputCommandInteraction, name: string, required: true): string;
export function getStringOption(interaction: ChatInputCommandInteraction, name: string, required?: false): string | null;
export function getStringOption(interaction: ChatInputCommandInteraction, name: string, required = false): string | null {
    const option = interaction.options.getString(name, required as boolean);
    return option;
}

/**
 * Get a user option value
 */
export function getUserOption(interaction: ChatInputCommandInteraction, name: string, required: true): User;
export function getUserOption(interaction: ChatInputCommandInteraction, name: string, required?: false): User | null;
export function getUserOption(interaction: ChatInputCommandInteraction, name: string, required = false): User | null {
    const option = interaction.options.getUser(name, required as boolean);
    return option;
}

/**
 * Get a channel option value
 */
export function getChannelOption(interaction: ChatInputCommandInteraction, name: string, required: true): GuildBasedChannel;
export function getChannelOption(interaction: ChatInputCommandInteraction, name: string, required?: false): GuildBasedChannel | null;
export function getChannelOption(interaction: ChatInputCommandInteraction, name: string, required = false): GuildBasedChannel | null {
    const option = interaction.options.getChannel(name, required as boolean);
    return option as GuildBasedChannel | null;
}

/**
 * Get a role option value
 */
export function getRoleOption(interaction: ChatInputCommandInteraction, name: string, required: true): Role;
export function getRoleOption(interaction: ChatInputCommandInteraction, name: string, required?: false): Role | null;
export function getRoleOption(interaction: ChatInputCommandInteraction, name: string, required = false): Role | null {
    const option = interaction.options.getRole(name, required as boolean);
    return option as Role | null;
}

/**
 * Get an integer option value
 */
export function getIntegerOption(interaction: ChatInputCommandInteraction, name: string, required: true): number;
export function getIntegerOption(interaction: ChatInputCommandInteraction, name: string, required?: false): number | null;
export function getIntegerOption(interaction: ChatInputCommandInteraction, name: string, required = false): number | null {
    const option = interaction.options.getInteger(name, required as boolean);
    return option;
}

/**
 * Get a number option value
 */
export function getNumberOption(interaction: ChatInputCommandInteraction, name: string, required: true): number;
export function getNumberOption(interaction: ChatInputCommandInteraction, name: string, required?: false): number | null;
export function getNumberOption(interaction: ChatInputCommandInteraction, name: string, required = false): number | null {
    const option = interaction.options.getNumber(name, required as boolean);
    return option;
}

/**
 * Get a boolean option value
 */
export function getBooleanOption(interaction: ChatInputCommandInteraction, name: string, required: true): boolean;
export function getBooleanOption(interaction: ChatInputCommandInteraction, name: string, required?: false): boolean | null;
export function getBooleanOption(interaction: ChatInputCommandInteraction, name: string, required = false): boolean | null {
    const option = interaction.options.getBoolean(name, required as boolean);
    return option;
}

/**
 * Get an attachment option value
 */
export function getAttachmentOption(interaction: ChatInputCommandInteraction, name: string, required: true): Attachment;
export function getAttachmentOption(interaction: ChatInputCommandInteraction, name: string, required?: false): Attachment | null;
export function getAttachmentOption(interaction: ChatInputCommandInteraction, name: string, required = false): Attachment | null {
    const option = interaction.options.getAttachment(name, required as boolean);
    return option;
}

/**
 * Get the subcommand name
 */
export function getSubcommand(interaction: ChatInputCommandInteraction): string {
    return interaction.options.getSubcommand();
}

/**
 * Get the subcommand group name
 */
export function getSubcommandGroup(interaction: ChatInputCommandInteraction): string | null {
    return interaction.options.getSubcommandGroup();
}
