import { PrismaClient } from "@prisma/client";

export const prisma = new PrismaClient();

export async function connectDatabase() {
	try {
		await prisma.$connect();
		// Test the connection with a simple query
		await prisma.$queryRaw`SELECT 1`;
		return true;
	} catch (error) {
		console.error("Erreur lors de la connexion à la base de données:", error);
		return false;
	}
}

export async function disconnectDatabase() {
	try {
		await prisma.$disconnect();
	} catch (error) {
		console.error(
			"Erreur lors de la déconnexion de la base de données:",
			error,
		);
	}
}
