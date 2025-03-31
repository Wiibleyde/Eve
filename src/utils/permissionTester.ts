import {
    CommandInteraction,
    ButtonInteraction,
    ModalSubmitInteraction,
    CacheType,
    PermissionFlagsBits,
    Interaction,
} from 'discord.js';
import { config } from '@/config';

/**
 * Checks if a user has the required permissions for a given interaction.
 *
 * @param interaction - The interaction to check permissions for. Can be a CommandInteraction, ButtonInteraction, or ModalSubmitInteraction.
 * @param permissionsToTests - An array of permissions (as bigint values) to check against the user's permissions.
 * @param exact - If true, the user must have all the permissions in the array. If false, the user only needs to have at least one of the permissions.
 * @returns A promise that resolves to a boolean indicating whether the user has the required permissions.
 */
export async function hasPermission(
    interaction: CommandInteraction | ButtonInteraction<CacheType> | ModalSubmitInteraction | Interaction<CacheType>,
    permissionsToTests: bigint[],
    exact: boolean
): Promise<boolean> {
    if (config.OWNER_ID == interaction.user.id) return true;
    const member = await interaction.guild?.members.fetch(interaction.user.id);
    if (!member) return false;
    if (member.permissions.has(PermissionFlagsBits.Administrator)) return true;
    if (permissionsToTests.length == 0) return false;
    if (exact) {
        return permissionsToTests.every((permission) => member.permissions.has(permission));
    } else {
        return permissionsToTests.some((permission) => member.permissions.has(permission));
    }
}
