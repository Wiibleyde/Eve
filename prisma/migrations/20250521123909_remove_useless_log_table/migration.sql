/*
  Warnings:

  - You are about to drop the `LogLevel` table. If the table is not empty, all the data it contains will be lost.
  - You are about to drop the `Logs` table. If the table is not empty, all the data it contains will be lost.

*/
-- DropForeignKey
ALTER TABLE `Logs` DROP FOREIGN KEY `Logs_levelId_fkey`;

-- DropTable
DROP TABLE `LogLevel`;

-- DropTable
DROP TABLE `Logs`;
