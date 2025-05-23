import { PrismaClient } from '@prisma/client';

export const prisma = new PrismaClient();

export async function disconnectDatabase() {
    try {
        await prisma.$disconnect();
    } catch (error) {
        console.error('Erreur lors de la déconnexion de la base de données:', error);
    }
}