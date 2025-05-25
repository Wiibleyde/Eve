import {
    PermissionFlagsBits,
    GuildMember,
    CommandInteraction,
    ButtonInteraction,
    ModalSubmitInteraction,
    type CacheType,
    type Interaction,
    type Snowflake
} from 'discord.js';
import { config } from './config';

/**
 * Type for supported interaction types in permission checking
 */
export type SupportedInteraction = CommandInteraction | ButtonInteraction<CacheType> | ModalSubmitInteraction | Interaction<CacheType>;

/**
 * Type representing permission requirements
 * Can be either permission flag bits or permission name strings
 */
export type PermissionRequirement = bigint | keyof typeof PermissionFlagsBits;

/**
 * Checks if a user has the required permissions or roles for a given interaction.
 *
 * @param interaction - The interaction to check permissions for.
 * @param requiredPermissions - An array of permissions (as bigint values or permission names) to check against the user's permissions.
 * @param options - Additional options for permission checking.
 * @param options.exact - If true, the user must have all the permissions in the array. If false, the user only needs to have at least one of the permissions. Default is false.
 * @param options.allowedRoleIds - Array of role IDs that are allowed to use this command regardless of permissions.
 * @returns A promise that resolves to a boolean indicating whether the user has the required permissions or roles.
 */
export async function hasPermission(
    interaction: SupportedInteraction,
    requiredPermissions: PermissionRequirement[] = [],
    options: {
        exact?: boolean,
        allowedRoleIds?: Snowflake[]
    } = {}
): Promise<boolean> {
    // Default options
    const { exact = false, allowedRoleIds = [] } = options;

    // Bot owner always has permission
    if (config.OWNER_ID === interaction.user.id) return true;

    // Get member from the interaction
    const member = await getMember(interaction);
    if (!member) return false;

    // Administrators always have permission
    if (member.permissions.has(PermissionFlagsBits.Administrator)) return true;

    // Check for allowed roles if specified
    if (allowedRoleIds.length > 0 && hasAnyRole(member, allowedRoleIds)) {
        return true;
    }

    // If no permissions are required, and we've reached this point, deny access
    if (requiredPermissions.length === 0) return false;

    // Convert any permission names to their bitfield values
    const permissionBits = requiredPermissions.map(perm => 
        typeof perm === 'string' ? PermissionFlagsBits[perm] : perm
    );

    // Check required permissions
    if (exact) {
        return permissionBits.every(permission => member.permissions.has(permission));
    } else {
        return permissionBits.some(permission => member.permissions.has(permission));
    }
}

/**
 * Gets a guild member from an interaction.
 * 
 * @param interaction - The interaction to get the member from.
 * @returns The guild member, or null if not found.
 */
async function getMember(interaction: SupportedInteraction): Promise<GuildMember | null> {
    try {
        return await interaction.guild?.members.fetch(interaction.user.id) || null;
    } catch {
        return null;
    }
}

/**
 * Checks if a member has any of the specified roles.
 * 
 * @param member - The guild member to check.
 * @param roleIds - Array of role IDs to check for.
 * @returns True if the member has any of the roles, false otherwise.
 */
function hasAnyRole(member: GuildMember, roleIds: Snowflake[]): boolean {
    return roleIds.some(roleId => member.roles.cache.has(roleId));
}