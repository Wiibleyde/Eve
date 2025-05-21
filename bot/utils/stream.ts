import type { StreamData, TwitchUser } from '../../utils/stream/twitch';

// Renamed functions with more descriptive names
export function handleStreamStarted(streamer: StreamData, userData: TwitchUser): void {
    // Implementation for when a stream goes online
}

export function handleStreamEnded(streamer: StreamData): void {
    // Implementation for when a stream goes offline
}

export function handleStreamUpdated(streamer: StreamData, userData: TwitchUser): void {
    // Implementation for when a stream's details change
}
