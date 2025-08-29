import { currentMusic } from '@bot/events/handlers/presenceUpdate';
import express from 'express';
import cors from 'cors';
import { logger } from '..';

const BASE_API_PATH = '/api/v1';

export const api = express()
    .use(cors({
        origin: '*',
        methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
        allowedHeaders: ['Content-Type', 'Authorization'],
        credentials: false
    }))
    .use(express.json());

export const startApiServer = (port: number) => {
    api.listen(port, () => {
        logger.info(`API server running on port ${port}`);
    });
};

api.get(`${BASE_API_PATH}/status`, (req, res) => {
    res.json({ status: 'ok', timestamp: new Date().toISOString() });
});

api.get(`${BASE_API_PATH}/health`, (req, res) => {
    res.json({ health: 'good', uptime: process.uptime() });
});

api.get(`${BASE_API_PATH}/info`, (req, res) => {
    res.json({ app: 'Eve Bot', version: '1.0.0', author: 'Your Name' });
});

api.get(`${BASE_API_PATH}/ping`, (req, res) => {
    res.json({ message: 'pong', timestamp: new Date().toISOString() });
});

api.get(`${BASE_API_PATH}/music`, (req, res) => {
    res.json({ currentlyPlaying: currentMusic });
});