import { client } from '@bot/bot';
import { logger } from '../..';
import { lsmsEmbedGenerator } from './lsms';

export interface LaboInQueryEntry {
    channeId: string;
    messageId: string;
    userId: string;
    startTime: Date;
    name: string;
    type: string;
    result?: string;
    time?: number;
}

class LaboInQueryManager {
    private entries: LaboInQueryEntry[] = [];
    private checkLoopStarted = false;
    private checkLoopInterval: NodeJS.Timeout | null = null;

    add(entry: LaboInQueryEntry) {
        this.entries.push(entry);
        if (!this.checkLoopStarted) this.startCheckLoop();
    }

    private startCheckLoop() {
        if (this.checkLoopStarted) return;
        this.checkLoopStarted = true;

        this.checkLoopInterval = setInterval(() => {
            const now = new Date();
            for (let i = this.entries.length - 1; i >= 0; i--) {
                const entry = this.entries[i];
                if (!entry) continue;
                const elapsed = now.getTime() - entry.startTime.getTime();
                if (elapsed >= (entry.time || 5) * 60000) {
                    this.notifyUserCompletion(entry);
                    this.entries.splice(i, 1);
                }
            }
            if (this.entries.length === 0 && this.checkLoopInterval) {
                clearInterval(this.checkLoopInterval);
                this.checkLoopInterval = null;
                this.checkLoopStarted = false;
            }
        }, 10000);
    }

    private async notifyUserCompletion(entry: LaboInQueryEntry) {
        const channel = await client.channels.fetch(entry.channeId);
        if (!channel || !channel.isTextBased()) {
            logger.warn(`Impossible de notifier l'utilisateur <@${entry.userId}> : canal introuvable ou non textuel.`);
            return;
        }
        const message = await channel.messages.fetch(entry.messageId).catch(() => null);
        if (message) {
            await message.edit({
                embeds: [
                    lsmsEmbedGenerator()
                        .setTitle('Analyse terminée')
                        .setDescription(`L'analyse pour **${entry.name}** est terminée.`)
                        .addFields(
                            { name: "Type d'analyse", value: this.getAnalyseType(entry), inline: true },
                            { name: 'Résultat', value: entry.result || "Erreur dans l'analyse", inline: true },
                            { name: 'Demandé par', value: `<@${entry.userId}>`, inline: true }
                        ),
                ],
            });
            const replyMessage = await message.reply({
                content: `<@${entry.userId}>, votre analyse est terminée. (Message supprimé automatiquement dans 1 minute)`,
            });

            logger.info(`Utilisateur <@${entry.userId}> notifié de la fin de l'analyse pour ${entry.name}.`);

            setTimeout(() => {
                replyMessage.delete().catch(() => null);
            }, 60000);
        } else {
            logger.warn(`Message introuvable pour notifier l'utilisateur <@${entry.userId}>.`);
        }

    }

    getAll(): LaboInQueryEntry[] {
        return [...this.entries];
    }

    public getAnalyseType(entry: LaboInQueryEntry): string {
        switch (entry.type) {
            case 'bloodgroup':
                return 'Groupe Sanguin';
            case 'alcohole':
                return "Taux d'Alcoolémie";
            case 'drugs':
                return 'Drogues';
            case 'diseases':
                return 'Maladies';
            default:
                return 'Analyse Inconnue';
        }
    }

    cancelByMessageId(messageId: string): { success: boolean; entry?: LaboInQueryEntry } {
        const index = this.entries.findIndex((e) => e.messageId === messageId);
        if (index !== -1) {
            const saveEntry = this.entries[index];
            this.entries.splice(index, 1);
            return { success: true, entry: saveEntry };
        }
        return { success: false };
    }
}

// Exporte une instance unique
export const laboInQueryManager = new LaboInQueryManager();

// Pour compatibilité, expose l'ancienne fonction via la classe
export function addLaboInQuery(entry: LaboInQueryEntry) {
    laboInQueryManager.add(entry);
}
