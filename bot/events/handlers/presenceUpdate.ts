import { ActivityType, Events } from "discord.js";
import type { Event } from "../event";
import { config } from "@utils/core/config";

interface NowPlayingMusic {
    title: string;
    artist: string;
    trackImageUrl: string;
}

export let currentMusic: NowPlayingMusic | null = null;

export const presenceUpdateEvent: Event<Events.PresenceUpdate> = {
    name: Events.PresenceUpdate,
    once: false,
    execute: async (oldPresence, newPresence) => {
        if (!newPresence || newPresence.userId !== config.OWNER_ID) return;

        const deezerActivity = newPresence.activities.find(
            (a) => a.type === ActivityType.Listening && a.name === "Deezer"
        );

        if (deezerActivity) {
            // Owner started listening to music on Deezer
            const details = deezerActivity.details || "Unknown Title";
            const state = deezerActivity.state || "Unknown Artist";
            const trackImageUrl = deezerActivity.assets?.largeImageURL() || "";
            currentMusic = { title: details, artist: state, trackImageUrl };
        } else {
            // Owner stopped listening to music on Deezer
            currentMusic = null;
        }
    }
};