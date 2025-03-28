/* eslint-disable no-useless-escape */
const quoiRegexs = [
    // Quoi
    /qu+o+i+[\ ?]*\?*$/,
    /qu+o+i+[\ +]*\?*$/,

    // Koa
    /ko+a+[\ ?]*\?*$/,
    /ko+a+[\ +]*\?*$/,

    // Qoa
    /q+o+a+[\ ?]*\?*$/,
    /q+o+a+[\ +]*\?*$/,

    // Koi
    /ko+i+[\ ?]*\?*$/,
    /ko+i+[\ +]*\?*$/,

    // Kwa
    /kw+a+[\ ?]*\?*$/,
    /kw+a+[\ +]*\?*$/,

    // Kewa
    /k+e+w+a+[\ ?]*\?*$/,
    /k+e+w+a+[\ +]*\?*$/,
];

const possibleFeurResponses = [
    {
        response: 'Feur.',
        probability: 70,
    },
    {
        response: 'coubeh.',
        probability: 10,
    },
    {
        response: 'la 🐨',
        probability: 10,
    },
    {
        response: 'drilatère.',
        probability: 10,
    },
];

export function detectFeur(message: string): boolean {
    for (const regex of quoiRegexs) {
        if (regex.test(message.toLowerCase())) {
            return true;
        }
    }
    return false;
}

export function generateFeurResponse(): string {
    const random = Math.random() * 100;
    let cumulativeProbability = 0;
    for (const response of possibleFeurResponses) {
        cumulativeProbability += response.probability;
        if (random <= cumulativeProbability) {
            return response.response;
        }
    }
    return 'Feur.';
}
