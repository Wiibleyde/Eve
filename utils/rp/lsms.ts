import {
    ActionRowBuilder,
    ButtonBuilder,
    ButtonStyle,
    TextChannel,
    type APIEmbed,
    type Embed,
    type EmbedBuilder,
    type User,
} from 'discord.js';
import { basicEmbedGenerator } from '../../bot/utils/embeds';
import { prisma } from '../core/database';
import { client } from '../../bot/bot';
import { logger } from '../..';

// Interface used to save LSMS duty data and send a summary at the reboot
interface LsmsDutySaveData {
    usersOnDuty: string[];
    usersOnCall: string[];
}

export interface LsmsRadioEntry {
    name: string;
    frequency: string;
}

const lsmsDutySaveData: Map<string, LsmsDutySaveData> = new Map(); // Map to store duty data by guild ID

const LSMS_RADIO_ADD_BUTTON_ID = 'lsmsRadioAdd';
const LSMS_RADIO_REMOVE_BUTTON_ID = 'lsmsRadioRemove';
const LSMS_RADIO_EDIT_BUTTON_PREFIX = 'lsmsRadioEdit';
const LSMS_RADIO_ADD_MODAL_PREFIX = 'lsmsRadioAddModal';
const LSMS_RADIO_EDIT_MODAL_PREFIX = 'lsmsRadioEditModal';
const LSMS_RADIO_REMOVE_SELECT_PREFIX = 'lsmsRadioRemoveSelect';

const MAX_RADIO_BUTTON_ROWS = 4; // 4 rows available after the primary actions row
const MAX_RADIOS = MAX_RADIO_BUTTON_ROWS * 5; // 5 buttons per row

export function onDutyUser(guildId: string, userId: string) {
    if (!lsmsDutySaveData.has(guildId)) {
        lsmsDutySaveData.set(guildId, { usersOnDuty: [], usersOnCall: [] });
    }
    const dutyData = lsmsDutySaveData.get(guildId)!;
    if (!dutyData.usersOnDuty.includes(userId)) {
        dutyData.usersOnDuty.push(userId);
    }
}

export function onCallUser(guildId: string, userId: string) {
    if (!lsmsDutySaveData.has(guildId)) {
        lsmsDutySaveData.set(guildId, { usersOnDuty: [], usersOnCall: [] });
    }
    const dutyData = lsmsDutySaveData.get(guildId)!;
    if (!dutyData.usersOnCall.includes(userId)) {
        dutyData.usersOnCall.push(userId);
    }
}

export function getLsmsSummary(guildId: string): LsmsDutySaveData | undefined {
    return lsmsDutySaveData.get(guildId);
}

export function lsmsEmbedGenerator() {
    return basicEmbedGenerator().setAuthor({
        name: 'Eve - LSMS',
        iconURL: 'https://cdn.discordapp.com/icons/872119977946263632/2100e7fea046e84d057cc0009a2b240f.webp',
    });
}

export function lsmsErrorEmbedGenerator(reason: string) {
    return basicEmbedGenerator()
        .setAuthor({
            name: 'Eve - LSMS',
            iconURL: 'https://cdn.discordapp.com/icons/872119977946263632/2100e7fea046e84d057cc0009a2b240f.webp',
        })
        .setColor(0xff0000)
        .setTitle('Erreur')
        .setDescription(reason);
}

export function lsmsSuccessEmbedGenerator(message: string) {
    return basicEmbedGenerator()
        .setAuthor({
            name: 'Eve - LSMS',
            iconURL: 'https://cdn.discordapp.com/icons/872119977946263632/2100e7fea046e84d057cc0009a2b240f.webp',
        })
        .setColor(0x00ff00)
        .setTitle('Succès')
        .setDescription(message);
}

export function lsmsDutyUpdateEmbedGenerator(userUpdated: User, take: boolean) {
    return lsmsEmbedGenerator()
        .setColor(take ? 0x00ff00 : 0xff0000)
        .setTitle(`${take ? 'Prise de' : 'Fin de'} service`)
        .setDescription(`<@${userUpdated.id}> a ${take ? 'pris' : 'terminé'} son service.`);
}

export function lsmsOnCallUpdateEmbedGenerator(userUpdated: User, take: boolean) {
    return lsmsEmbedGenerator()
        .setColor(take ? 0x00ff00 : 0xff0000)
        .setTitle(`${take ? 'Début du' : 'Fin du'} semi service`)
        .setDescription(`<@${userUpdated.id}> a ${take ? 'débuté' : 'terminé'} son semi service.`);
}

export function lsmsDutyEmbedGenerator(
    onDutyPeople: User[],
    onCallPeople: User[]
): { embed: EmbedBuilder; actionRow: ActionRowBuilder<ButtonBuilder> } {
    const dutyList =
        onDutyPeople.length > 0
            ? onDutyPeople.map((user) => `<@${user.id}>`).join('\n')
            : "Personne n'est en service :(";
    const callList =
        onCallPeople.length > 0
            ? onCallPeople.map((user) => `<@${user.id}>`).join('\n')
            : "Personne n'est en semi service :(";
    const embed = lsmsEmbedGenerator()
        .setTitle('Gestionnaire de service')
        .setDescription('Cliquez sur les boutons ci-dessous pour gérer les services.')
        .addFields(
            {
                name: 'En service :',
                value: dutyList,
                inline: true,
            },
            {
                name: 'En semi service :',
                value: callList,
                inline: true,
            }
        )
        .setFooter({
            iconURL: lsmsEmbedGenerator().data.footer?.icon_url || '',
            text: '⚠️ Cela peut prendre 5 secondes pour que les changements soient pris en compte.',
        });
    const buttons = [
        new ButtonBuilder()
            .setCustomId('handleLsmsDuty')
            .setLabel('Prendre/Quitter le service')
            .setStyle(ButtonStyle.Primary),
        new ButtonBuilder()
            .setCustomId('handleLsmsOnCall')
            .setLabel('Prendre/Quitter le semi service')
            .setStyle(ButtonStyle.Secondary),
    ];
    const actionRow = new ActionRowBuilder<ButtonBuilder>().addComponents(buttons);

    return {
        embed,
        actionRow,
    };
}

export async function prepareLsmsSummary(): Promise<void> {
    try {
        logger.info('Préparation du résumé LSMS...');
        const uptime = process.uptime();

        const dutyManagers = await prisma.lsmsDutyManager.findMany();
        logger.info(`Found ${dutyManagers.length} LSMS duty managers.`);

        // Utiliser Promise.all pour attendre toutes les opérations asynchrones
        await Promise.all(
            dutyManagers.map(async (dutyManager) => {
                logger.info(`Processing duty manager for guild ${dutyManager.guildId}`);
                const lastRebootDutyData = getLsmsSummary(dutyManager.guildId);

                if (lastRebootDutyData) {
                    logger.info(
                        lastRebootDutyData.usersOnDuty.map((user) => `<@${user}>`).join(', '),
                        lastRebootDutyData.usersOnCall.map((user) => `<@${user}>`).join(', ')
                    );
                    const dutyFieldFormatted = lastRebootDutyData.usersOnDuty.map((user) => `<@${user}>`).join(', ');
                    const onCallFieldFormatted = lastRebootDutyData.usersOnCall.map((user) => `<@${user}>`).join(', ');
                    const dutyEmbed = lsmsEmbedGenerator()
                        .setTitle(`**Récapitulatif du service**`)
                        .setDescription(
                            `Période du <t:${(new Date(Date.now() - uptime * 1000).getTime() / 1000) | 0}:f> au <t:${(Date.now() / 1000) | 0}:f>`
                        )
                        .addFields(
                            { name: 'Service', value: dutyFieldFormatted || 'Aucun :(' },
                            { name: 'Semi service', value: onCallFieldFormatted || 'Aucun :(' }
                        );

                    logger.info(`Sending duty summary embed to logs channel for guild ${dutyManager.guildId}`);

                    if (dutyManager.logsChannelId) {
                        logger.info(`Logs channel ID: ${dutyManager.logsChannelId}`);
                        try {
                            const channel = (await client.channels.fetch(dutyManager.logsChannelId)) as TextChannel;
                            if (channel && channel.isTextBased()) {
                                logger.info(`Sending embed to channel ${channel.name} (${channel.id})`);
                                await channel.send({ embeds: [dutyEmbed] });
                                logger.info(`Embed sent successfully to channel ${channel.name} (${channel.id})`);
                            } else {
                                logger.error(
                                    `Le channel de logs LSMS n'est pas valide dans le serveur ${dutyManager.guildId}`
                                );
                            }
                        } catch (channelError) {
                            logger.error(
                                `Erreur lors de l'envoi du résumé LSMS au channel ${dutyManager.logsChannelId}: ${channelError}`
                            );
                        }
                    }
                } else {
                    logger.warn(`Aucune donnée de service trouvée pour le serveur ${dutyManager.guildId}`);
                }
            })
        );

        logger.info('Résumé LSMS préparé avec succès.');
    } catch (error) {
        logger.error(`Erreur lors de la préparation du résumé LSMS: ${error}`);
        throw error;
    }
}

export function encodeRadioName(name: string): string {
    return Buffer.from(name, 'utf8').toString('base64url');
}

export function decodeRadioName(encodedName: string): string {
    return Buffer.from(encodedName, 'base64url').toString('utf8');
}

function truncateLabel(label: string, maxLength = 80 - 'Modifier la radio '.length): string {
    if (label.length <= maxLength) {
        return label;
    }

    return `${label.slice(0, maxLength - 1)}…`;
}

export function getRadioCustomIds() {
    return {
        addButton: LSMS_RADIO_ADD_BUTTON_ID,
        removeButton: LSMS_RADIO_REMOVE_BUTTON_ID,
        editButtonPrefix: LSMS_RADIO_EDIT_BUTTON_PREFIX,
        addModalPrefix: LSMS_RADIO_ADD_MODAL_PREFIX,
        editModalPrefix: LSMS_RADIO_EDIT_MODAL_PREFIX,
        removeSelectPrefix: LSMS_RADIO_REMOVE_SELECT_PREFIX,
    } as const;
}

export function parseRadiosFromEmbed(embed?: APIEmbed | EmbedBuilder | Embed | null): LsmsRadioEntry[] {
    if (!embed) {
        return [];
    }

    const rawEmbed: APIEmbed =
        typeof (embed as EmbedBuilder).toJSON === 'function'
            ? ((embed as EmbedBuilder).toJSON() as APIEmbed)
            : embed instanceof Object && 'data' in embed
              ? ((embed as EmbedBuilder).data as APIEmbed)
              : (embed as APIEmbed);

    const fields = rawEmbed.fields;

    if (!fields) {
        return [];
    }

    return fields
        .map((field) => ({
            name: field.name?.trim() ?? '',
            frequency: field.value?.trim() ?? '',
        }))
        .filter((field) => field.name.length > 0 && field.frequency.length > 0);
}

export function buildRadioEmbed(radios: LsmsRadioEntry[]): EmbedBuilder {
    const embed = lsmsEmbedGenerator()
        .setTitle('Gestionnaire de radios')
        .setDescription(
            radios.length > 0
                ? 'Utilisez les boutons ci-dessous pour gérer les radios disponibles.'
                : 'Aucune radio configurée pour le moment.\nAjoutez une radio pour commencer.'
        );

    if (radios.length > 0) {
        embed.setFields(
            radios.map((radio) => ({
                name: radio.name,
                value: radio.frequency,
                inline: true,
            }))
        );
    }

    return embed;
}

export function buildRadioComponents(radios: LsmsRadioEntry[]): ActionRowBuilder<ButtonBuilder>[] {
    const rows: ActionRowBuilder<ButtonBuilder>[] = [];

    let currentRow = new ActionRowBuilder<ButtonBuilder>();
    radios.slice(0, MAX_RADIOS).forEach((radio, index) => {
        const button = new ButtonBuilder()
            .setCustomId(`${LSMS_RADIO_EDIT_BUTTON_PREFIX}--${encodeRadioName(radio.name)}`)
            .setLabel(`Modifier la radio ${truncateLabel(radio.name)}`)
            .setStyle(ButtonStyle.Secondary);

        currentRow.addComponents(button);

        if (currentRow.components.length === 5 || index === Math.min(radios.length, MAX_RADIOS) - 1) {
            rows.push(currentRow);
            currentRow = new ActionRowBuilder<ButtonBuilder>();
        }
    });

    const baseRow = new ActionRowBuilder<ButtonBuilder>().addComponents(
        new ButtonBuilder().setCustomId(LSMS_RADIO_ADD_BUTTON_ID).setLabel('+').setStyle(ButtonStyle.Success),
        new ButtonBuilder().setCustomId(LSMS_RADIO_REMOVE_BUTTON_ID).setLabel('-').setStyle(ButtonStyle.Danger)
    );

    rows.push(baseRow);

    if (radios.length === 0) {
        return rows;
    }

    return rows;
}

export function canAddRadio(radios: LsmsRadioEntry[]): boolean {
    return radios.length < MAX_RADIOS;
}

export function formatRadioModalId(prefix: string, channelId: string, messageId: string, extra?: string): string {
    const parts = [prefix, channelId, messageId];
    if (extra) {
        parts.push(extra);
    }
    return parts.join('--');
}

export function decodeRadioModalId(customId: string): {
    channelId: string | null;
    messageId: string | null;
    extra: string | null;
} {
    const [, channelId = null, messageId = null, extra = null] = customId.split('--');
    return { channelId, messageId, extra };
}

export function decodeRadioButtonId(customId: string): string | null {
    const [, encodedName] = customId.split('--');
    if (!encodedName) {
        return null;
    }

    try {
        return decodeRadioName(encodedName);
    } catch (error) {
        logger.warn(`Impossible de décoder le nom de la radio depuis le bouton: ${String(error)}`);
        return null;
    }
}

export function updateRadioMessage(radios: LsmsRadioEntry[]) {
    return {
        embed: buildRadioEmbed(radios),
        components: buildRadioComponents(radios),
    };
}
