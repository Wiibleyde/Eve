import { SlashCommandBuilder, SlashCommandOptionsOnlyBuilder } from 'discord.js';

interface Formation {
    title: string;
    subFormations: string[];
}

const MARSFormations: Formation[] = [
    {
        title: 'Information générale',
        subFormations: [
            'Permis voiture',
            'Permis moto',
            'Permis poids lourd',
            'Licence hélicoptère',
            "Permis de port d'arme",
            'Visite hopital',
        ],
    },
    {
        title: 'Formation principale',
        subFormations: [
            'Appel coma',
            'Opération',
            'Rédaction des rapports',
            'Bobologie',
            'Don du sang',
            'Rappel / Parachute',
            'Natation / Plongée',
            'Intervention pompier',
        ],
    },
    {
        title: 'Formation secondaire',
        subFormations: ['Secret médical', 'Matériel', 'Communication radio'],
    },
];

const CompaniesFormations: Map<string, Formation[]> = new Map([['MARS', MARSFormations]]);

export const data: SlashCommandOptionsOnlyBuilder = new SlashCommandBuilder()
    .setName('createemployee')
    .setDescription('Créer une fiche employé')
    .addStringOption((option) =>
        option
            .setName('company')
            .setDescription("Nom de l'entreprise")
            .setRequired(true)
            .addChoices(
                Array.from(CompaniesFormations.keys()).map((company) => ({
                    name: company,
                    value: company,
                }))
            )
    )
    .addStringOption((option) => option.setName('name').setDescription("Nom de l'employé").setRequired(true))
    .addStringOption((option) => option.setName('firstname').setDescription("Prénom de l'employé").setRequired(true));
