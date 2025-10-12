import type { StringSelectMenuInteraction } from 'discord.js';
import { handleLsmsRadioRemoveSelect } from './handlers/rp/handleLsmsRadioRemoveSelect';

export const selectMenus: Record<string, (interaction: StringSelectMenuInteraction) => Promise<void>> = {
    lsmsRadioRemoveSelect: handleLsmsRadioRemoveSelect,
};
