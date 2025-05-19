import { CommandInteraction, MessageFlags, SlashCommandBuilder } from 'discord.js';
import type { ICommand } from '../command';
import { basicEmbedGenerator } from '../../utils/embeds';

const PING_THRESHOLD = {
    VERY_GOOD: 50,
    GOOD: 100,
    CORRECT: 150,
    WEAK: 200,
    BAD: 500,
};

interface PingStatus {
    status: string;
    color: number;
}

export const ping: ICommand = {
    data: new SlashCommandBuilder().setName('ping').setDescription('Replies with Pong!'),
    execute: async (interaction: CommandInteraction) => {
        await interaction.deferReply({
            withResponse: true,
            flags: [MessageFlags.Ephemeral],
        });

        const memoryData = process.memoryUsage();
        const pingValue = interaction.client.ws.ping;
        const { status, color } = getPingStatus(pingValue);

        const pingEmbed = createPingEmbed(pingValue, status, color, memoryData);
        await interaction.editReply({ embeds: [pingEmbed] });
    },
};

/**
 * Determines ping status and color based on the ping value
 */
function getPingStatus(ping: number): PingStatus {
    if (ping <= 0) {
        return { status: 'Surprenant ! ⚫', color: 0x000000 };
    } else if (ping < PING_THRESHOLD.VERY_GOOD) {
        return { status: 'Très bon 🟢', color: 0x00ff00 };
    } else if (ping < PING_THRESHOLD.GOOD) {
        return { status: 'Bon 🟢', color: 0x00ff00 };
    } else if (ping < PING_THRESHOLD.CORRECT) {
        return { status: 'Correct 🟡', color: 0x00ff00 };
    } else if (ping < PING_THRESHOLD.WEAK) {
        return { status: 'Faible 🟠', color: 0xffa500 };
    } else if (ping < PING_THRESHOLD.BAD) {
        return { status: 'Mauvais 🔴', color: 0xff0000 };
    } else {
        return { status: 'Très mauvais 🔴', color: 0xff0000 };
    }
}

/**
 * Creates a formatted ping embed with system information
 */
function createPingEmbed(ping: number, status: string, color: number, memoryData: NodeJS.MemoryUsage) {
    const pingEmbed = basicEmbedGenerator();

    // Calcul des informations système formatées
    const memoryUsed = (memoryData.heapUsed / 1024 / 1024).toFixed(2);
    const uptime = formatUptime(process.uptime());

    // Calcul du pourcentage avec une valeur max de 100%
    const memoryPercent = Math.min(Math.round((memoryData.heapUsed / memoryData.rss) * 100), 100);

    pingEmbed
        .setTitle('📊 État du système')
        .setDescription(`**Le bot est opérationnel** ${getStatusEmoji(status)}`)
        .setTimestamp()
        .addFields(
            {
                name: '🏓 Latence',
                value: `\`\`\`\n${ping}ms\n${status}\`\`\``,
                inline: true,
            },
            {
                name: '💾 Mémoire',
                value: `\`\`\`\n${memoryUsed} MB / ${(memoryData.rss / 1024 / 1024).toFixed(2)} MB RSS\n${memoryPercent}% utilisé\`\`\``,
                inline: true,
            },
            {
                name: '⏱️ Uptime',
                value: `\`\`\`\n${uptime}\`\`\``,
                inline: true,
            }
        )
        .setAuthor({
            name: 'Eve - Ping',
            iconURL: "https://cdn.discordapp.com/attachments/1373968524229218365/1373975736943120455/image.png?ex=682c5e1e&is=682b0c9e&hm=cf2d1c4ea686b7a5e298142d6b40e509d978bccbce6c7fdfc6d66086e2c7e60c&",
        })
        .setColor(color);

    return pingEmbed;
}

/**
 * Format uptime in a more readable format
 */
function formatUptime(uptime: number): string {
    const days = Math.floor(uptime / 86400);
    const hours = Math.floor((uptime % 86400) / 3600);
    const minutes = Math.floor((uptime % 3600) / 60);
    const seconds = Math.floor(uptime % 60);

    if (days > 0) {
        return `${days}j ${hours}h ${minutes}m`;
    } else if (hours > 0) {
        return `${hours}h ${minutes}m ${seconds}s`;
    } else {
        return `${minutes}m ${seconds}s`;
    }
}

/**
 * Get appropriate emoji based on status
 */
function getStatusEmoji(status: string): string {
    if (status.includes('Très bon')) return '✅';
    if (status.includes('Bon')) return '✅';
    if (status.includes('Correct')) return '✓';
    if (status.includes('Faible')) return '⚠️';
    if (status.includes('Mauvais')) return '❌';
    if (status.includes('Très mauvais')) return '🆘';
    return '❓';
}
