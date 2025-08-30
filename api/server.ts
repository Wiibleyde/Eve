import express from 'express';
import cors from 'cors';
import { logger } from '..';
import { baseHandler, healthHandler, infoHandler, pingHandler } from './handlers/base';
import { currentMusicHandler } from './handlers/music';

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

api.get(`${BASE_API_PATH}/status`, baseHandler);
api.get(`${BASE_API_PATH}/health`, healthHandler);
api.get(`${BASE_API_PATH}/info`, infoHandler);
api.get(`${BASE_API_PATH}/ping`, pingHandler);
api.get(`${BASE_API_PATH}/music`, currentMusicHandler);