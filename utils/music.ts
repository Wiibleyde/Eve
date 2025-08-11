import { client } from "@bot/bot";
import { basicEmbedGenerator } from "@bot/utils/embeds";

export function musicEmbedGenerator() {
    return basicEmbedGenerator().setAuthor({
        name: 'Eve - Musique',
        iconURL: client.user?.displayAvatarURL() || '',
    })
}

export function musicErrorEmbedGenerator(reason: string) {
    return basicEmbedGenerator()
        .setAuthor({
            name: 'Eve - Musique',
            iconURL: client.user?.displayAvatarURL() || '',
        })
        .setColor(0xff0000)
        .setTitle('Erreur')
        .setDescription(reason);
}

export function musicSuccessEmbedGenerator(message: string) {
    return basicEmbedGenerator()
        .setAuthor({
            name: 'Eve - Musique',
            iconURL: client.user?.displayAvatarURL() || '',
        })
        .setTitle('Succès')
        .setDescription(message)
        .setColor(0x00ff00);
}
